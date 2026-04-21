# StakeStock — Virtual Trading Platform

Gamified trading platform where users trade virtual assets with in-game coins.
Prices are determined purely by supply and demand through a real order book matching engine.

## Stack

| Layer      | Technology                                |
|------------|-------------------------------------------|
| Matching   | Go — `backend/internal/orderbook` (price-time priority) |
| Backend    | Go + chi, pgx/v5, go-redis, coder/ws      |
| Auth       | Google OAuth 2.0 + cookie sessions        |
| Database   | PostgreSQL 16                              |
| Cache/PubSub | Redis 7                                 |
| Frontend   | React 18 + Vite + TypeScript              |
| Infra      | Docker Compose                            |

## Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- A Google OAuth 2.0 app — create one at <https://console.cloud.google.com/apis/credentials>
  - Authorised redirect URI: `http://localhost:8080/auth/google/callback`

## Quick Start

### 1. Configure environment

```bash
cp .env.example .env
# Edit .env and fill in GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET
```

### 2. Start infrastructure (Postgres + Redis)

```bash
docker compose up -d
```

### 3. Start backend

```bash
cd backend
go mod tidy          # first time only — downloads dependencies
go run ./cmd/server
```

The backend runs migrations automatically on startup.

### 4. Start frontend (separate terminal)

```bash
cd frontend
npm install          # first time only
npm run dev
```

Open <http://localhost:5173>.

---

## Architecture

```
Browser <—WS/HTTP—> Go backend (chi)
                      |
             ┌────────┴────────┐
         PostgreSQL          Redis
         (durable state)   (pub/sub + cache)
                      |
               Engine Registry
               (one AssetBook goroutine per asset)
                      |
                orderbook lib
                (price-time priority matching)
```

### Trading mechanics

- Every user starts with **1000 coins**.
- **Limit orders** — placed at a specific price; rest in the book if not immediately matched.
- **Market orders** — execute at best available price; fully-covered sells only.
- **Short selling** — limit-only. Backend reserves `qty × price` cash as collateral when a sell would take your position below zero.
- Prices are 100% demand/supply — there is no reference price until the first trade.

### Order lifecycle

1. Reservation TX — lock cash or position in Postgres.
2. Engine call — send order to the per-asset goroutine.
3. Settlement TX — write trades, update balances/positions, publish Redis events.

### Real-time

WebSocket at `/ws`. Send `{"subscribe":["asset:GOLD:depth","asset:GOLD:trades","user:<id>:orders"]}`. Receives `{"channel":"...","data":{...}}` envelopes.

---

## Project layout

```
stakestock/
├── backend/
│   ├── cmd/server/main.go   # entry point
│   ├── internal/
│   │   ├── config/          # env loading
│   │   ├── db/              # pgxpool + embedded migrations
│   │   ├── models/          # shared types
│   │   ├── auth/            # Google OAuth2, sessions
│   │   ├── engine/          # AssetBook goroutine wrapper + registry
│   │   ├── accounts/        # balance/position reservation helpers
│   │   ├── trading/         # order intake + settlement
│   │   ├── pubsub/          # Redis pub/sub + cache
│   │   ├── ws/              # WebSocket hub
│   │   └── httpapi/         # chi router + handlers
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── pages/           # Login, Onboarding, Dashboard, Asset, Portfolio, History
│   │   ├── components/      # OrderBookLadder, OrderForm, TradesFeed, NavBar
│   │   ├── api/             # fetch client + TypeScript types
│   │   ├── ws/              # WebSocket hook
│   │   └── store/           # Zustand auth store
│   └── Dockerfile
├── docker-compose.yml
└── .env.example
```

## Docker full-stack build

```bash
cp .env.example .env   # fill in Google OAuth credentials
docker compose --profile full up --build
```

Frontend → <http://localhost:5173>  
Backend  → <http://localhost:8080>
