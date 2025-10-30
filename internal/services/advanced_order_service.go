package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"trading-simulator/internal/models"
	"trading-simulator/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AdvancedOrderService struct {
	orderCollection     *mongo.Collection
	portfolioCollection *mongo.Collection
	marketDataService   *MarketDataService
	orderService        *OrderService
}

func NewAdvancedOrderService(marketDataService *MarketDataService) *AdvancedOrderService {
	return &AdvancedOrderService{
		orderCollection:     config.GetCollection("advanced_orders"),
		portfolioCollection: config.GetCollection("portfolio"),
		marketDataService:   marketDataService,
		orderService:        NewOrderService(marketDataService), // fixed: pass marketDataService
	}
}

func (s *AdvancedOrderService) CreateStopOrder(order *models.Order) error {
	order.ID = primitive.NewObjectID()
	order.Timestamp = time.Now()
	order.Status = "active"

	if order.Type == "sell" {
		var portfolio models.Portfolio
		err := s.portfolioCollection.FindOne(context.Background(), bson.M{
			"user_id": order.UserID,
			"symbol":  order.Symbol,
		}).Decode(&portfolio)

		if err != nil || portfolio.Shares < order.Quantity {
			return fmt.Errorf("insufficient shares for stop loss order")
		}
	}

	_, err := s.orderCollection.InsertOne(context.Background(), order)
	if err != nil {
		return err
	}

	log.Printf("STOP Order Created: %s %s %d shares @ $%.2f trigger for user %s",
		order.Symbol, order.Type, order.Quantity, order.StopPrice, order.UserID)
	return nil
}

func (s *AdvancedOrderService) CheckAndExecuteStopOrders() {
	cursor, err := s.orderCollection.Find(context.Background(), bson.M{
		"status": "active",
		"order_type": bson.M{"$in": []string{"stop", "stop_limit", "trailing_stop"}},
	})
	if err != nil {
		return
	}
	defer cursor.Close(context.Background())

	var activeOrders []models.Order
	if err = cursor.All(context.Background(), &activeOrders); err != nil {
		return
	}

	for _, order := range activeOrders {
		currentPrice := s.getCurrentPrice(order.Symbol)

		if s.shouldTriggerStopOrder(order, currentPrice) {
			s.executeStopOrder(&order, currentPrice)
		}
	}
}

func (s *AdvancedOrderService) getCurrentPrice(symbol string) float64 {
	stock, err := s.marketDataService.GetStockPrice(symbol)
	if err != nil {
		return 100.0
	}
	return stock.Price
}

func (s *AdvancedOrderService) shouldTriggerStopOrder(order models.Order, currentPrice float64) bool {
	switch order.OrderType {
	case "stop":
		if order.Type == "sell" {
			return currentPrice <= order.StopPrice
		}
		return currentPrice >= order.StopPrice
	case "stop_limit":
		if order.Type == "sell" {
			return currentPrice <= order.StopPrice && currentPrice >= order.LimitPrice
		}
		return currentPrice >= order.StopPrice && currentPrice <= order.LimitPrice
	case "trailing_stop":
		if order.Type == "sell" {
			return currentPrice <= order.StopPrice
		}
		return currentPrice >= order.StopPrice
	}
	return false
}

func (s *AdvancedOrderService) executeStopOrder(order *models.Order, currentPrice float64) {
	_, err := s.orderCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": order.ID},
		bson.M{"$set": bson.M{
			"status":       "triggered",
			"triggered_at": time.Now(),
			"price":        currentPrice,
		}},
	)
	if err != nil {
		log.Printf("Error updating stop order: %v", err)
		return
	}

	executionOrder := &models.Order{
		UserID:    order.UserID,
		Symbol:    order.Symbol,
		Type:      order.Type,
		OrderType: "market",
		Quantity:  order.Quantity,
		Price:     currentPrice,
	}

	if err = s.orderService.PlaceOrder(executionOrder); err != nil {
		log.Printf("Error executing stop order: %v", err)
	} else {
		log.Printf("STOP Order Triggered: %s %s %d shares @ $%.2f for user %s",
			order.Symbol, order.Type, order.Quantity, currentPrice, order.UserID)
	}
}

func (s *AdvancedOrderService) GetActiveStopOrders(userID string) ([]models.Order, error) {
	cursor, err := s.orderCollection.Find(context.Background(), bson.M{
		"user_id": userID,
		"status":  "active",
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var orders []models.Order
	err = cursor.All(context.Background(), &orders)
	return orders, err
}

func (s *AdvancedOrderService) CancelStopOrder(orderID string) error {
	objID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return err
	}

	_, err = s.orderCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"status": "cancelled"}},
	)
	return err
}