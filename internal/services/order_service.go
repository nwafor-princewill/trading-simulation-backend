package services

import (
	"context"
	"fmt"
	"time"

	"trading-simulator/internal/models"
	"trading-simulator/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderService struct {
	orderCollection     *mongo.Collection
	portfolioCollection *mongo.Collection
	userCollection      *mongo.Collection
	marketService       *MarketDataService
}

func NewOrderService(marketService *MarketDataService) *OrderService {
	return &OrderService{
		orderCollection:     config.GetCollection("orders"),
		portfolioCollection: config.GetCollection("portfolio"),
		userCollection:      config.GetCollection("users"),
		marketService:       marketService,
	}
}

func (s *OrderService) PlaceOrder(order *models.Order) error {
	order.ID = primitive.NewObjectID()
	order.Timestamp = time.Now()
	order.Status = "filled"

	switch order.Type {
	case "buy":
		return s.executeBuyOrder(order)
	case "sell":
		return s.executeSellOrder(order)
	default:
		return fmt.Errorf("invalid order type: %s", order.Type)
	}
}

func (s *OrderService) executeBuyOrder(order *models.Order) error {
	cash := s.GetCashBalance(order.UserID)
	cost := order.Price * float64(order.Quantity)
	if cash < cost {
		return fmt.Errorf("insufficient funds. have $%.2f, need $%.2f", cash, cost)
	}

	_, err := s.orderCollection.InsertOne(context.Background(), order)
	if err != nil {
		return err
	}

	var pos models.Portfolio
	err = s.portfolioCollection.FindOne(context.Background(), bson.M{
		"user_id": order.UserID,
		"symbol":  order.Symbol,
	}).Decode(&pos)

	if err == mongo.ErrNoDocuments {
		pos = models.Portfolio{
			ID:      primitive.NewObjectID(),
			UserID:  order.UserID,
			Symbol:  order.Symbol,
			Shares:  order.Quantity,
			AvgCost: order.Price,
		}
		_, err = s.portfolioCollection.InsertOne(context.Background(), pos)
	} else if err == nil {
		totalCost := (pos.AvgCost * float64(pos.Shares)) + cost
		totalShares := pos.Shares + order.Quantity
		newAvg := totalCost / float64(totalShares)

		_, err = s.portfolioCollection.UpdateOne(
			context.Background(),
			bson.M{"_id": pos.ID},
			bson.M{"$set": bson.M{
				"shares":   totalShares,
				"avg_cost": newAvg,
			}},
		)
	}
	if err != nil {
		return err
	}

	userID, _ := primitive.ObjectIDFromHex(order.UserID)
	_, err = s.userCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": userID},
		bson.M{"$inc": bson.M{"cash_balance": -cost}},
	)
	return err
}

func (s *OrderService) executeSellOrder(order *models.Order) error {
	var pos models.Portfolio
	err := s.portfolioCollection.FindOne(context.Background(), bson.M{
		"user_id": order.UserID,
		"symbol":  order.Symbol,
	}).Decode(&pos)
	if err == mongo.ErrNoDocuments {
		return fmt.Errorf("you own no %s", order.Symbol)
	}
	if err != nil {
		return err
	}
	if pos.Shares < order.Quantity {
		return fmt.Errorf("insufficient shares: have %d, want %d", pos.Shares, order.Quantity)
	}

	_, err = s.orderCollection.InsertOne(context.Background(), order)
	if err != nil {
		return err
	}

	newShares := pos.Shares - order.Quantity
	if newShares == 0 {
		_, err = s.portfolioCollection.DeleteOne(context.Background(), bson.M{"_id": pos.ID})
	} else {
		_, err = s.portfolioCollection.UpdateOne(
			context.Background(),
			bson.M{"_id": pos.ID},
			bson.M{"$set": bson.M{"shares": newShares}},
		)
	}
	if err != nil {
		return err
	}

	revenue := order.Price * float64(order.Quantity)
	userID, _ := primitive.ObjectIDFromHex(order.UserID)
	_, err = s.userCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": userID},
		bson.M{"$inc": bson.M{"cash_balance": revenue}},
	)
	return err
}

func (s *OrderService) GetUserPortfolio(userID string) ([]models.Portfolio, error) {
	cur, err := s.portfolioCollection.Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())
	var list []models.Portfolio
	_ = cur.All(context.Background(), &list)
	return list, nil
}

func (s *OrderService) GetUserOrders(userID string) ([]models.Order, error) {
	cur, err := s.orderCollection.Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())
	var list []models.Order
	_ = cur.All(context.Background(), &list)
	return list, nil
}

func (s *OrderService) GetCashBalance(userID string) float64 {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 10000.0
	}
	var u models.User
	err = s.userCollection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&u)
	if err != nil {
		return 10000.0
	}
	return u.CashBalance
}

func (s *OrderService) GetTotalPortfolioValue(userID string) float64 {
	pos, err := s.GetUserPortfolio(userID)
	if err != nil {
		return 0
	}
	val := 0.0
	for _, p := range pos {
		stock, err := s.marketService.GetMockStockPrice(p.Symbol)
		if err == nil {
			val += stock.Price * float64(p.Shares)
		}
	}
	return val
}