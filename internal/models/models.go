package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Stock struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Symbol    string             `bson:"symbol" json:"symbol"`
	Name      string             `bson:"name" json:"name"`
	Price     float64            `bson:"price" json:"price"`
	Change    float64            `bson:"change" json:"change"`
	ChangePercent float64        `bson:"change_percent" json:"changePercent"`
	Volume    int64              `bson:"volume" json:"volume"`
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
}

type Order struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID          string             `bson:"user_id" json:"userId"`
	Symbol          string             `bson:"symbol" json:"symbol"`
	Type            string             `bson:"type" json:"type"`                         // "buy" or "sell"
	OrderType       string             `bson:"order_type" json:"orderType"`             // "market", "limit", "stop", "stop_limit", "trailing_stop"
	Quantity        int                `bson:"quantity" json:"quantity"`
	Price           float64            `bson:"price" json:"price"`                      // Execution price for market/limit, limit price for stop-limit
	StopPrice       float64            `bson:"stop_price,omitempty" json:"stopPrice"`   // Trigger price for stop orders
	LimitPrice      float64            `bson:"limit_price,omitempty" json:"limitPrice"` // Limit price for stop-limit orders
	TrailingPercent float64            `bson:"trailing_percent,omitempty" json:"trailingPercent"`
	Status          string             `bson:"status" json:"status"` // "pending", "filled", "cancelled", "active", "triggered"
	Timestamp       time.Time          `bson:"timestamp" json:"timestamp"`
	TriggeredAt     time.Time          `bson:"triggered_at,omitempty" json:"triggeredAt"`
}
type Portfolio struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID  string             `bson:"user_id" json:"userId"`
	Symbol  string             `bson:"symbol" json:"symbol"`
	Shares  int                `bson:"shares" json:"shares"`
	AvgCost float64            `bson:"avg_cost" json:"avgCost"`
}