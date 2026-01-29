package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/haflettjm/eth-balance-service/internal/eth"
"github.com/haflettjm/eth-balance-service/internal/util"

)

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return "8080"
	}
	if _, err := strconv.Atoi(port); err != nil {
		log.Printf("invalid PORT %q, falling back to 8080", port)
		return "8080"
	}
	return port
}

type balanceService interface {
	GetBalanceETH(ctx context.Context, address string, block string) (string, error)
}

type server struct {
	balances balanceService
}

func (s server) registerRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/address/balance/:address", func(c *gin.Context) {
		address := c.Param("address")
		if !util.IsValidEthAddress(address) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ethereum address"})
			return
		}

		block := c.DefaultQuery("block", "latest")
		switch block {
		case "latest", "pending":
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid block (use latest|pending)"})
			return
		}

		// Use request context so Cloud Run cancels on client disconnect / deadline.
		bal, err := s.balances.GetBalanceETH(c.Request.Context(), address, block)
		if err != nil {
			// Upstream dependency failed
			c.JSON(http.StatusBadGateway, gin.H{"error": "upstream rpc error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"balance": bal, // string is safer than float
			"unit":    "ETH",
			"block":   block,
			"source":  "infura",
		})
	})
}

func main() {
	// Gin mode: default to release unless explicitly set for dev.
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	infuraKey := os.Getenv("INFURA_API_KEY")
	client, err := eth.NewInfuraClient(infuraKey)
	if err != nil {
		log.Fatalf("failed to init infura client: %v", err)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	s := server{balances: client}
	s.registerRoutes(r)

	port := getPort()

	// Production-ish HTTP server (timeouts help against slowloris)
	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("listening on :%s", port)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
