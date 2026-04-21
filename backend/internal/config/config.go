package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port                string
	DatabaseURL         string
	RedisURL            string
	GoogleClientID      string
	GoogleClientSecret  string
	GoogleRedirectURL   string
	SessionCookieSecure bool
	FrontendOrigin      string
	StartingCoins       string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:               getenv("PORT", "8080"),
		DatabaseURL:        getenv("DATABASE_URL", "postgres://stake:stake@localhost:5432/stakestock"),
		RedisURL:           getenv("REDIS_URL", "redis://localhost:6379/0"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  getenv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/google/callback"),
		FrontendOrigin:     getenv("FRONTEND_ORIGIN", "http://localhost:5173"),
		StartingCoins:      getenv("STARTING_COINS", "1000"),
	}

	secureCookie, _ := strconv.ParseBool(getenv("SESSION_COOKIE_SECURE", "false"))
	cfg.SessionCookieSecure = secureCookie

	if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set")
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
