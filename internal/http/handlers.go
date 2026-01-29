package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type BalanceGetter interface {
	GetBalanceETH(ctx *gin.Context, address string, block string) (string, error)
}

type Handlers struct {
	Balances BalanceGetter
}

func (h Handlers) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h Handlers) Balance(c *gin.Context) {
	addr := c.Param("address")
	block := c.DefaultQuery("block", "latest") // optional: latest|pending

	bal, err := h.Balances.GetBalanceETH(c, addr, block)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance": bal,
		"unit":    "ETH",
	})
}

