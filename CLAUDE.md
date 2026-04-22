# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Full Stack
```bash
cp .env.example .env      # first-time setup (fill in Google OAuth creds)
make dev                  # docker (DB/Redis) + backend + frontend concurrently
make d                    # docker compose up -d (DB + Redis only)
make down                 # docker compose down
```

### Backend (Go)
```bash
cd backend
go run ./cmd/server       # start API server (auto-runs migrations on startup)
go test ./...             # all tests
go test ./internal/orderbook/...  # orderbook unit tests only
```

### Frontend (Node)
```bash
cd frontend
npm install
npm run dev               # Vite dev server → http://localhost:5173
npm run build             # tsc + vite build
```

### Docker (full stack)
```bash
docker compose --profile full up --build
```

## Architecture

StakeStock is a gamified virtual trading platform where asset prices are determined purely by supply/demand through a real order book. No external price feeds.

### Backend (`backend/`)

**Entry point**: `cmd/server/main.go` — loads config → runs migrations → creates pgxpool + Redis → boots engine registry → starts HTTP server.

**Internal packages** (each is a distinct concern):

| Package | Role |
|---|---|
| `config/` | Env var loading |
| `db/` | pgxpool + embedded migrations (golang-migrate) |
| `models/` | Shared types: User, Balance, Asset, Position, Order, Trade |
| `auth/` | Google OAuth2, cookie sessions, context helpers |
| `orderbook/` | Core matching engine — price-time priority, fully unit-tested |
| `engine/` | `AssetBook` goroutine wrapper + `Registry` (one goroutine per asset) |
| `accounts/` | Balance/position reservation helpers |
| `trading/` | Order intake: reservation TX → engine → settlement TX |
| `pubsub/` | Redis pub/sub + market data cache |
| `ws/` | WebSocket hub |
| `httpapi/` | chi v5 router + HTTP handlers |
| `pricerec/` | Background goroutine recording price history |

**Order lifecycle**: reservation TX (lock cash/qty in Postgres) → send to per-asset goroutine → settlement TX (write trades, update balances/positions, publish Redis events).

**Router** (`internal/httpapi/router.go`):
- `/auth/google/*` — OAuth flow
- `/ws` — WebSocket (subscribe via channel names)
- `/api/*` — all require auth cookie; POST `/api/me/username` requires onboarding gate

**Tech**: Go 1.22, chi v5, pgx/v5, go-redis v9, coder/websocket, golang-migrate, shopspring/decimal, emirpasic/gods.

### Frontend (`frontend/`)

**Vite proxy**: `/api/*`, `/auth/*`, `/ws` all proxy to `http://localhost:8080` in dev — no CORS friction locally.

**React Router** (`App.tsx`):
- Public: `/login`, `/onboarding`
- Protected (AuthGuard): `/` (Dashboard), `/asset/:symbol`, `/portfolio`, `/history`

**Data layer**: TanStack React Query for server state + custom fetch wrapper (`src/api/client.ts`) with `credentials: 'include'`. Zustand store for auth state only.

**Real-time**: WebSocket hook (`src/ws/`) receives `{"channel":"...","data":{...}}` envelopes. Channels: `asset:SYMBOL:depth`, `asset:SYMBOL:trades`, `user:ID:orders`.

**Tech**: React 18, Vite 7, TypeScript 5.5 (strict), React Router 6, React Query 5, Zustand 4, Recharts 2.

### Database

PostgreSQL 16 with embedded SQL migrations in `backend/migrations/`. Migrations run automatically on server start. Key tables: `users`, `balances`, `assets`, `positions`, `orders`, `trades`, `sessions`.

### Environment Variables

Required in `.env` (see `.env.example`):
- `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET` — Google OAuth app credentials
- `GOOGLE_REDIRECT_URL` — must match OAuth console (default: `http://localhost:8080/auth/google/callback`)
- `DATABASE_URL`, `REDIS_URL` — set automatically by Docker Compose
- `FRONTEND_ORIGIN` — for CORS (default: `http://localhost:5173`)
- `STARTING_COINS` — coins granted to new users (default: 1000)
