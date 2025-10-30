
---

## 2. `backend/README.md`

```markdown
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
