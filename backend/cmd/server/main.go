package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/stakestock/backend/internal/auth"
	"github.com/stakestock/backend/internal/config"
	"github.com/stakestock/backend/internal/db"
	"github.com/stakestock/backend/internal/engine"
	"github.com/stakestock/backend/internal/httpapi"
	"github.com/stakestock/backend/internal/pricerec"
	"github.com/stakestock/backend/internal/pubsub"
	"github.com/stakestock/backend/internal/trading"
	"github.com/stakestock/backend/internal/ws"
)

func main() {
	// Load .env file if present (dev convenience)
	loadDotEnv()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run migrations
	if err := db.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("migrations: %v", err)
	}
	log.Println("migrations: ok")

	// Postgres pool
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db pool: %v", err)
	}
	defer pool.Close()
	log.Println("database: connected")

	// Redis
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis url: %v", err)
	}
	rdb := redis.NewClient(redisOpts)
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis ping: %v", err)
	}
	defer rdb.Close()
	log.Println("redis: connected")

	bus := pubsub.NewBus(rdb)

	// Engine registry (builds asset books + rehydrates)
	reg := engine.NewRegistry()
	if err := reg.Build(ctx, pool); err != nil {
		log.Fatalf("engine registry: %v", err)
	}

	// Price recorder (background goroutine)
	go pricerec.New(pool).Run(ctx)

	// Services
	tradingSvc := trading.NewService(pool, reg, bus)

	// Auth
	googleProvider := auth.NewGoogleProvider(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL)
	sessionStore := auth.NewSessionStore(pool, cfg.SessionCookieSecure, cfg.StartingCoins)

	// WebSocket hub
	hub := ws.NewHub(bus)

	// HTTP handler
	handler := httpapi.NewHandler(pool, googleProvider, sessionStore, reg, tradingSvc, bus, hub, cfg.FrontendOrigin)
	r := handler.Routes()

	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func loadDotEnv() {
	data, err := os.ReadFile(".env")
	if err != nil {
		return // no .env file, that's fine
	}
	for _, line := range splitLines(string(data)) {
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		for i, c := range line {
			if c == '=' {
				key := line[:i]
				val := line[i+1:]
				if os.Getenv(key) == "" {
					os.Setenv(key, val)
				}
				break
			}
		}
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
