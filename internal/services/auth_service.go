package services

import (
	"context"
	"errors"
	"log"
	"time"

	"trading-simulator/internal/models"
	"trading-simulator/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthService struct {
	userCollection *mongo.Collection
}

func NewAuthService() *AuthService {
	return &AuthService{
		userCollection: config.GetCollection("users"),
	}
}

// Register creates a new user
func (s *AuthService) Register(user *models.User) error {
	// Check if user already exists
	var existingUser models.User
	err := s.userCollection.FindOne(context.Background(), bson.M{
		"$or": []bson.M{
			{"username": user.Username},
			{"email": user.Email},
		},
	}).Decode(&existingUser)

	if err == nil {
		return errors.New("username or email already exists")
	} else if err != mongo.ErrNoDocuments {
		return err
	}

	// Hash password
	err = user.HashPassword()
	if err != nil {
		return err
	}

	// Set default values
	user.ID = primitive.NewObjectID()
	user.CashBalance = 10000.0 // Start with $10,000
	user.CreatedAt = time.Now()

	// Insert user
	_, err = s.userCollection.InsertOne(context.Background(), user)
	if err != nil {
		return err
	}

	log.Printf("✅ New user registered: %s", user.Username)
	return nil
}

// Login authenticates a user
func (s *AuthService) Login(username, password string) (*models.User, error) {
	var user models.User
	err := s.userCollection.FindOne(context.Background(), bson.M{
		"username": username,
	}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("invalid username or password")
		}
		return nil, err
	}

	// Check password
	if !user.CheckPassword(password) {
		return nil, errors.New("invalid username or password")
	}

	// Don't return password hash
	user.Password = ""
	return &user, nil
}

// GetUserByID returns a user by their ID
func (s *AuthService) GetUserByID(userID string) (*models.User, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var user models.User
	err = s.userCollection.FindOne(context.Background(), bson.M{
		"_id": objID,
	}).Decode(&user)

	if err != nil {
		return nil, err
	}

	user.Password = ""
	return &user, nil
}