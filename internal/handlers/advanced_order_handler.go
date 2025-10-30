package handlers

import (
	"net/http"
	"time"

	"trading-simulator/internal/models"
	"trading-simulator/internal/services"
	"github.com/gin-gonic/gin"
)

type AdvancedOrderHandler struct {
	service *services.AdvancedOrderService
}

func NewAdvancedOrderHandler(service *services.AdvancedOrderService) *AdvancedOrderHandler {
	return &AdvancedOrderHandler{service: service}
}

type StopOrderRequest struct {
	Symbol     string  `json:"symbol" binding:"required"`
	Type       string  `json:"type" binding:"required"`
	OrderType  string  `json:"orderType" binding:"required"`
	Quantity   int     `json:"quantity" binding:"required,min=1"`
	Price      float64 `json:"price" binding:"required,min=0.01"`
	StopPrice  float64 `json:"stopPrice" binding:"required,min=0.01"`
	LimitPrice float64 `json:"limitPrice,omitempty"`
}

func (h *AdvancedOrderHandler) CreateStopOrder(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	var req StopOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	o := &models.Order{
		UserID:     userID.(string),
		Symbol:     req.Symbol,
		Type:       req.Type,
		OrderType:  req.OrderType,
		Quantity:   req.Quantity,
		Price:      req.Price,
		StopPrice:  req.StopPrice,
		LimitPrice: req.LimitPrice,
		Status:     "active",
		Timestamp:  time.Now(),
	}

	if err := h.service.CreateStopOrder(o); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Stop order created",
		"order":   o,
	})
}

func (h *AdvancedOrderHandler) GetActiveOrders(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	list, err := h.service.GetActiveStopOrders(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"orders": list})
}

// userID is extracted but not used in service → keep it for consistency
func (h *AdvancedOrderHandler) CancelOrder(c *gin.Context) {
	_, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	orderID := c.Param("id")

	if err := h.service.CancelStopOrder(orderID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "order cancelled"})
}