package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"trading-simulator/config"
	"trading-simulator/internal/handlers"
	"trading-simulator/internal/services"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize MongoDB
	config.ConnectDB()

	// Initialize services
	marketService := services.NewMarketDataService()
	wsHub := services.NewWebSocketHub()
	orderService := services.NewOrderService(marketService)
	advancedOrderService := services.NewAdvancedOrderService(marketService)
	authService := services.NewAuthService()

	// Start WebSocket hub in goroutine
	go wsHub.Run()

	// Start market data simulator
	go simulateMarketData(wsHub, marketService)

	// Start stop order monitoring
	go monitorStopOrders(advancedOrderService)

	// Create Gin router
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Initialize handlers
	marketHandler := handlers.NewMarketHandler(marketService)
	orderHandler := handlers.NewOrderHandler(orderService)
	advancedOrderHandler := handlers.NewAdvancedOrderHandler(advancedOrderService)
	authHandler := handlers.NewAuthHandler(authService)

	// Auth middleware helper
	authMiddleware := authHandler.AuthMiddleware()

	// Routes
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "OK",
			"message": "Trading Simulator API",
			"version": "1.0.0",
			"endpoints": []string{
				"GET /health",
				"GET /api/stocks/:symbol",
				"GET /ws",
				"POST /api/orders/place",
				"GET /api/portfolio", 
				"GET /api/orders",
				"POST /api/advanced-orders/stop",
				"GET /api/advanced-orders/active",
				"POST /api/advanced-orders/cancel/:id",
				"POST /api/auth/register",
				"POST /api/auth/login",
				"GET /api/auth/me",
			},
		})
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "OK",
			"message": "Trading Simulator API is running",
		})
	})

	// Market data routes
	router.GET("/api/stocks/:symbol", marketHandler.GetStockPrice)

	// WebSocket endpoint
	router.GET("/ws", func(c *gin.Context) {
		username := c.Query("username")
		if username == "" {
			username = "Anonymous"
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Failed to upgrade connection: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upgrade to WebSocket"})
			return
		}

		client := wsHub.RegisterClient(conn, username)
		log.Printf("WebSocket connection established for user: %s", username)

		// Start client pumps
		go client.WritePump()
		go client.ReadPump()
	})

	// Protected order routes - require authentication
	router.POST("/api/orders/place", authMiddleware, orderHandler.PlaceOrder)
	router.GET("/api/portfolio", authMiddleware, orderHandler.GetPortfolio)
	router.GET("/api/orders", authMiddleware, orderHandler.GetOrders)

	// Protected advanced order routes - require authentication
	router.POST("/api/advanced-orders/stop", authMiddleware, advancedOrderHandler.CreateStopOrder)
	router.GET("/api/advanced-orders/active", authMiddleware, advancedOrderHandler.GetActiveOrders)
	router.POST("/api/advanced-orders/cancel/:id", authMiddleware, advancedOrderHandler.CancelOrder)

	// Auth routes
	router.POST("/api/auth/register", authHandler.Register)
	router.POST("/api/auth/login", authHandler.Login)
	router.GET("/api/auth/me", authMiddleware, authHandler.GetCurrentUser)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	fmt.Printf("🚀 Trading Simulator Backend running on port %s\n", port)
	fmt.Printf("📊 API available at http://localhost:%s\n", port)
	fmt.Printf("🔌 WebSocket available at ws://localhost:%s/ws\n", port)
	fmt.Printf("🔐 Auth available at http://localhost:%s/api/auth\n", port)
	router.Run(":" + port)
}

// Simulate market data updates
func simulateMarketData(hub *services.WebSocketHub, marketService *services.MarketDataService) {
	symbols := []string{"AAPL", "GOOGL", "MSFT", "TSLA", "AMZN"}
	
	// Add delay before starting to allow server to fully initialize
	time.Sleep(2 * time.Second)
	log.Println("📈 Starting market data simulation...")

	// Get initial real data once
	log.Println("🔄 Fetching initial real stock data...")
	for _, symbol := range symbols {
		stock, err := marketService.GetStockPrice(symbol)
		if err != nil {
			log.Printf("❌ Error fetching %s: %v", symbol, err)
			continue
		}
		hub.BroadcastStock(*stock)
		log.Printf("✅ Initial data: %s - $%.2f", symbol, stock.Price)
		time.Sleep(1 * time.Second) // Respect API limits
	}

	// Use mock data for continuous updates (no API calls)
	log.Println("🤖 Switching to mock data for real-time updates...")
	ticker := time.NewTicker(3 * time.Second) // Update every 3 seconds
	defer ticker.Stop()

	for range ticker.C {
		// Use mock data only - no API calls
		for _, symbol := range symbols {
			stock, err := marketService.GetMockStockPrice(symbol)
			if err != nil {
				log.Printf("❌ Mock data error for %s: %v", symbol, err)
				continue
			}
			hub.BroadcastStock(*stock)
		}
	}
}

// Monitor stop orders in background
func monitorStopOrders(advancedOrderService *services.AdvancedOrderService) {
	// Wait for server to fully initialize
	time.Sleep(5 * time.Second)
	log.Println("🛑 Starting stop order monitoring...")

	ticker := time.NewTicker(10 * time.Second) // Check every 10 seconds
	defer ticker.Stop()

	for range ticker.C {
		advancedOrderService.CheckAndExecuteStopOrders()
	}
}