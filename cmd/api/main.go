package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
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

func main() {
	// Gin defaults to debug mode will need to update if ever sent to prod
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Health check
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Placeholder for future route
	r.GET("/address/balance/:address", func(c *gin.Context) {
		address := c.Param("address")
		c.JSON(http.StatusOK, gin.H{
			"address": address,
			"balance": "0",
			"unit":    "ETH",
		})
	})

	port := getPort()
	log.Printf("starting gin server on :%s", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

