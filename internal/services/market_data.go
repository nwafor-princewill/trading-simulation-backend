package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"trading-simulator/internal/models"
)

type AlphaVantageResponse struct {
	GlobalQuote struct {
		Symbol        string `json:"01. symbol"`
		Price         string `json:"05. price"`
		Change        string `json:"09. change"`
		ChangePercent string `json:"10. change percent"`
	} `json:"Global Quote"`
}

type AlphaVantageError struct {
	Information string `json:"Information"`
}

type MarketDataService struct {
	apiKey         string
	useMockData    bool
	lastAPISuccess time.Time
	mockPrices     map[string]float64
}

func NewMarketDataService() *MarketDataService {
	apiKey := os.Getenv("ALPHA_VANTAGE_API_KEY")
	if apiKey == "" {
		log.Fatal("ALPHA_VANTAGE_API_KEY not set in environment variables")
	}

	// Initialize mock prices with realistic values
	mockPrices := map[string]float64{
		"AAPL":  175.50,
		"GOOGL": 138.25,
		"MSFT":  330.80,
		"TSLA":  210.75,
		"AMZN":  178.90,
	}

	return &MarketDataService{
		apiKey:         apiKey,
		useMockData:    false, // Start with real API
		lastAPISuccess: time.Now(),
		mockPrices:     mockPrices,
	}
}

func (m *MarketDataService) GetStockPrice(symbol string) (*models.Stock, error) {
	// Try real API first (if we haven't been using mock data for too long)
	if !m.useMockData || time.Since(m.lastAPISuccess) > 30*time.Minute {
		stock, err := m.getRealStockPrice(symbol)
		if err == nil {
			m.lastAPISuccess = time.Now()
			m.useMockData = false // Real API worked, switch back
			return stock, nil
		}

		// If real API fails, switch to mock data
		log.Printf("⚠️ Real API failed for %s, switching to mock data: %v", symbol, err)
		m.useMockData = true
	}

	// Use mock data
	return m.getMockStockPrice(symbol)
}

func (m *MarketDataService) getRealStockPrice(symbol string) (*models.Stock, error) {
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=%s", symbol, m.apiKey)

	// Create HTTP client with timeout
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Check for API rate limit errors
	var apiError AlphaVantageError
	if err := json.Unmarshal(body, &apiError); err == nil && apiError.Information != "" {
		if strings.Contains(apiError.Information, "rate limit") {
			return nil, fmt.Errorf("API rate limit exceeded: %s", apiError.Information)
		}
	}

	var alphaResponse AlphaVantageResponse
	err = json.Unmarshal(body, &alphaResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	// Check if we got valid data
	if alphaResponse.GlobalQuote.Symbol == "" || alphaResponse.GlobalQuote.Price == "" {
		return nil, fmt.Errorf("no data returned for symbol %s", symbol)
	}

	// Parse price with better error handling
	price, err := parsePrice(alphaResponse.GlobalQuote.Price)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price: %v", err)
	}

	change, err := parsePrice(alphaResponse.GlobalQuote.Change)
	if err != nil {
		change = 0 // Default to 0 if change parsing fails
	}

	changePercent, err := parseChangePercent(alphaResponse.GlobalQuote.ChangePercent)
	if err != nil {
		changePercent = 0 // Default to 0 if percent parsing fails
	}

	stock := &models.Stock{
		Symbol:        strings.ToUpper(alphaResponse.GlobalQuote.Symbol),
		Name:          getStockName(alphaResponse.GlobalQuote.Symbol),
		Price:         price,
		Change:        change,
		ChangePercent: changePercent,
		Volume:        0, // Alpha Vantage doesn't provide volume in this endpoint
		Timestamp:     time.Now(),
	}

	log.Printf("✅ Real API: %s - $%.2f (%.2f%%)", stock.Symbol, stock.Price, stock.ChangePercent)
	return stock, nil
}

func (m *MarketDataService) getMockStockPrice(symbol string) (*models.Stock, error) {
	// Get base price or use default
	basePrice, exists := m.mockPrices[symbol]
	if !exists {
		basePrice = 100.0 // Default base price
		m.mockPrices[symbol] = basePrice
	}

	// Generate realistic price movement (±2%)
	changePercent := (rand.Float64()*4 - 2) // -2% to +2%
	change := basePrice * changePercent / 100
	newPrice := basePrice + change

	// Update mock price for next call
	m.mockPrices[symbol] = newPrice

	stock := &models.Stock{
		Symbol:        strings.ToUpper(symbol),
		Name:          getStockName(symbol),
		Price:         newPrice,
		Change:        change,
		ChangePercent: changePercent,
		Volume:        rand.Int63n(10000000) + 1000000, // Random volume
		Timestamp:     time.Now(),
	}

	log.Printf("🤖 Mock Data: %s - $%.2f (%+.2f%%)", stock.Symbol, stock.Price, stock.ChangePercent)
	return stock, nil
}

func parsePrice(priceStr string) (float64, error) {
	if priceStr == "" {
		return 0, fmt.Errorf("empty price string")
	}

	// Remove any non-numeric characters except decimal point and minus
	cleaned := strings.TrimSpace(priceStr)

	// Parse the float
	price, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse '%s' as float: %v", cleaned, err)
	}

	return price, nil
}

func parseChangePercent(percentStr string) (float64, error) {
	if percentStr == "" {
		return 0, fmt.Errorf("empty percent string")
	}

	// Remove percentage sign and trim
	cleaned := strings.TrimSpace(strings.TrimSuffix(percentStr, "%"))

	percent, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse '%s' as float: %v", cleaned, err)
	}

	return percent, nil
}

func getStockName(symbol string) string {
	names := map[string]string{
		"AAPL":  "Apple Inc.",
		"GOOGL": "Alphabet Inc.",
		"MSFT":  "Microsoft Corporation",
		"TSLA":  "Tesla Inc.",
		"AMZN":  "Amazon.com Inc.",
		"NVDA":  "NVIDIA Corporation",
		"META":  "Meta Platforms Inc.",
		"JPM":   "JPMorgan Chase & Co.",
	}

	if name, exists := names[strings.ToUpper(symbol)]; exists {
		return name
	}

	return fmt.Sprintf("%s Corporation", symbol)
}

// GetMultipleStockPrices fetches prices for multiple symbols
func (m *MarketDataService) GetMultipleStockPrices(symbols []string) ([]models.Stock, error) {
	var stocks []models.Stock

	for _, symbol := range symbols {
		stock, err := m.GetStockPrice(symbol)
		if err != nil {
			log.Printf("Error fetching %s: %v", symbol, err)
			continue // Skip failed requests but continue with others
		}
		stocks = append(stocks, *stock)

		// Add small delay between requests
		time.Sleep(100 * time.Millisecond)
	}

	return stocks, nil
}

// GetMockStockPrice generates realistic mock stock data without API calls
func (m *MarketDataService) GetMockStockPrice(symbol string) (*models.Stock, error) {
	// Get base price from our mock prices
	basePrice, exists := m.mockPrices[symbol]
	if !exists {
		// Set realistic base prices for each symbol
		realisticPrices := map[string]float64{
			"AAPL":  269.00,
			"GOOGL": 267.47,
			"MSFT":  542.07,
			"TSLA":  460.55,
			"AMZN":  229.25,
		}
		basePrice = realisticPrices[symbol]
		m.mockPrices[symbol] = basePrice
	}

	// Generate realistic price movement (±1.5%)
	changePercent := (rand.Float64()*3 - 1.5) // -1.5% to +1.5%
	change := basePrice * changePercent / 100
	newPrice := basePrice + change

	// Update mock price for next call (with some momentum)
	m.mockPrices[symbol] = newPrice

	// Generate realistic volume
	volume := rand.Int63n(5000000) + 1000000

	stock := &models.Stock{
		Symbol:        strings.ToUpper(symbol),
		Name:          getStockName(symbol),
		Price:         newPrice,
		Change:        change,
		ChangePercent: changePercent,
		Volume:        volume,
		Timestamp:     time.Now(),
	}

	log.Printf("🤖 Mock Data: %s - $%.2f (%+.2f%%)", stock.Symbol, stock.Price, stock.ChangePercent)
	return stock, nil
}