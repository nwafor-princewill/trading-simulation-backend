package handlers

import (
	"net/http"
	"time"

	"trading-simulator/internal/models"
	"trading-simulator/internal/services"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService *services.OrderService
}

func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// PlaceOrderRequest - for regular market/limit orders
type PlaceOrderRequest struct {
	Symbol    string  `json:"symbol" binding:"required"`
	Type      string  `json:"type" binding:"required"`      // "buy" or "sell"
	OrderType string  `json:"orderType" binding:"required"` // "market" or "limit"
	Quantity  int     `json:"quantity" binding:"required,min=1"`
	Price     float64 `json:"price" binding:"required,min=0.01"`
}

func (h *OrderHandler) PlaceOrder(c *gin.Context) {
	// Get authenticated user ID from JWT
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req PlaceOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Validate order type
	if req.OrderType != "market" && req.OrderType != "limit" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order type. Must be 'market' or 'limit'"})
		return
	}

	// Validate order type (buy/sell)
	if req.Type != "buy" && req.Type != "sell" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order type. Must be 'buy' or 'sell'"})
		return
	}

	// Create order object
	order := &models.Order{
		UserID:    userID.(string),
		Symbol:    req.Symbol,
		Type:      req.Type,
		OrderType: req.OrderType,
		Quantity:  req.Quantity,
		Price:     req.Price,
		Status:    "filled", // Immediate execution
		Timestamp: time.Now(),
	}

	// Execute the order
	err := h.orderService.PlaceOrder(order)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order placed successfully",
		"order":   order,
	})
}

func (h *OrderHandler) GetPortfolio(c *gin.Context) {
	// Get authenticated user ID from JWT
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	portfolio, err := h.orderService.GetUserPortfolio(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch portfolio: " + err.Error()})
		return
	}

	cashBalance := h.orderService.GetCashBalance(userID.(string))

	c.JSON(http.StatusOK, gin.H{
		"portfolio":    portfolio,
		"cashBalance":  cashBalance,
		"totalAssets":  cashBalance + h.orderService.GetTotalPortfolioValue(userID.(string)),
	})
}

func (h *OrderHandler) GetOrders(c *gin.Context) {
	// Get authenticated user ID from JWT
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	orders, err := h.orderService.GetUserOrders(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}