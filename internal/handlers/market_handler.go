package handlers

import (
	"net/http"
	"trading-simulator/internal/services"
	"github.com/gin-gonic/gin"
)

type MarketHandler struct {
	marketService *services.MarketDataService
}

func NewMarketHandler(marketService *services.MarketDataService) *MarketHandler {
	return &MarketHandler{marketService: marketService}
}

func (h *MarketHandler) GetStockPrice(c *gin.Context) {
	symbol := c.Param("symbol")
	
	stock, err := h.marketService.GetStockPrice(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stock)
}