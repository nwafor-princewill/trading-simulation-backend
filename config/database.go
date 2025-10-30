package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Client

func ConnectDB() {
	mongoURI := os.Getenv("MONGODB_URI")
	
	if mongoURI == "" {
		log.Fatal("MONGODB_URI environment variable is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use mongo.Connect() instead of mongo.NewClient()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	// Ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	DB = client
	fmt.Println("✅ Connected to MongoDB")
}

// Getting database collections
func GetCollection(collectionName string) *mongo.Collection {
	databaseName := os.Getenv("DATABASE_NAME")
	if databaseName == "" {
		databaseName = "trading-simulator"
	}
	collection := DB.Database(databaseName).Collection(collectionName)
	return collection
}

// Disconnect closes the MongoDB connection
func DisconnectDB() {
	if DB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		if err := DB.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
		fmt.Println("MongoDB connection closed")
	}
}