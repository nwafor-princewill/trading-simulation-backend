

## 2. `backend/README.md`


# Trading Simulator – Backend API

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Gin](https://img.shields.io/badge/Gin-00D1D1?style=for-the-badge&logo=go&logoColor=white)
![MongoDB](https://img.shields.io/badge/MongoDB-4EA94B?style=for-the-badge&logo=mongodb&logoColor=white)
![Render](https://img.shields.io/badge/Render-46E3B7?style=for-the-badge&logo=render&logoColor=white)

> **Real-time trading engine** with WebSocket broadcasting, JWT auth, MongoDB persistence, and advanced order execution.

---

## Features

| Feature | Description |
|--------|-------------|
| **WebSocket Broadcast** | Live stock prices to **all connected clients** |
| **Auth** | Register / Login → JWT (7-day expiry) |
| **Orders** | Market, Limit, Stop-Loss, Take-Profit |
| **Portfolio** | Position tracking, P&L, cash balance |
| **Mock + Real Data** | Alpha Vantage fallback |
| **CORS** | Configured for Vercel frontend |
| **Rate Limiting** | 100 req/min per IP |

---

## Tech Stack

- **Language** – Go 1.22  
- **Framework** – Gin Web Framework  
- **Database** – MongoDB Atlas (free tier)  
- **WebSocket** – Native `gorilla/websocket`  
- **Deploy** – Render.com (free web service)

---

## Project Structure
/
├── handlers/        # API routes
├── models/          # MongoDB structs
├── websocket/       # Hub + broadcaster
├── main.go
├── go.mod
└── .env


---

## WebSocket – The Live Market Core


// websocket/hub.go
type Hub struct {
    clients   map[*Client]bool
    broadcast chan Stock
}

func (h *Hub) Run() {
    for stock := range h.broadcast {
        for client := range h.clients {
            client.send <- stock
        }
    }
}


Flow

simulateMarketData() → runs every 3s
Fetches 10 stocks (mock + Alpha Vantage)
hub.BroadcastStock(stock) → sends to all clients
Frontend receives → updates UI instantly


No polling – true real-time
Scalable – one goroutine per client

Getting Started

1. Clone & Install
bashgit clone https://github.com/nwafor-princewill/trading-simulator-backend.git
cd trading-simulator-backend
go mod tidy
2. Environment (.env)
envMONGODB_URI=mongodb+srv://<user>:<pass>@cluster0.mongodb.net/trading
JWT_SECRET=your-super-secret-jwt-key
ALPHA_VANTAGE_KEY=your-free-api-key
PORT=8080
3. Run Locally
bashgo run main.go
API: http://localhost:8080
WebSocket: ws://localhost:8080/ws

Method,         Route,              Description
POST,        /api/auth/register,    Create user
POST,       /api/auth/login,        Get JWT
GET,        /api/portfolio,         User positions
POST,      /api/orders,            Place order
GET,       /api/orders,           Order history
GET,      /ws,                   WebSocket feed

Deploy to Render

Connect GitHub repo
Set Build Command: go build -o main .
Start Command: ./main
Add env vars in dashboard


Auto-deploys on push


Scripts
bashgo run main.go          # dev
go build                # production binary

