// Package config loads application configuration from environment variables.
package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration for the application.
type Config struct {
	// Server
	Port            string
	Env             string
	LogLevel        string
	MaxBodySize     int
	ShutdownTimeout time.Duration
	RateLimitPerMin int

	// Database
	DatabaseURL string

	// Redis
	RedisURL string
	RedisTTL time.Duration

	// NATS
	NATSURL string

	// JWT
	JWTSecret string
	JWTExpiry time.Duration

	// Crypto
	EncryptionKey string

	// WebSocket
	WSMaxConnections int

	// Bootstrap admin (seeded on first start)
	AdminEmail    string
	AdminPassword string

	// Outbox relay
	OutboxPollInterval time.Duration
}

// Load reads configuration from the environment, optionally from a .env file.
func Load() *Config {
	_ = godotenv.Load() // .env is optional; explicit env vars take precedence
	return &Config{
		Port:               getStr("PORT", "8080"),
		Env:                getStr("ENV", "development"),
		LogLevel:           getStr("LOG_LEVEL", "debug"),
		MaxBodySize:        getInt("MAX_BODY_SIZE", 8<<20),
		ShutdownTimeout:    getDur("SHUTDOWN_TIMEOUT", 15*time.Second),
		RateLimitPerMin:    getInt("RATE_LIMIT_PER_MIN", 120),
		DatabaseURL:        getStr("DATABASE_URL", "postgres://aeroxe:secret@localhost:5432/sign-flow?sslmode=disable"),
		RedisURL:           getStr("REDIS_URL", "redis://localhost:6379"),
		RedisTTL:           getDur("REDIS_TTL", 15*time.Minute),
		NATSURL:            getStr("NATS_URL", "nats://localhost:4222"),
		JWTSecret:          getStr("JWT_SECRET", "change-me-in-production-please-32b!"),
		JWTExpiry:          getDur("JWT_EXPIRY", 24*time.Hour),
		EncryptionKey:      getStr("ENCRYPTION_KEY", "dev-only-encryption-key-change-me!"),
		WSMaxConnections:   getInt("WS_MAX_CONNECTIONS", 1000),
		AdminEmail:         strings.ToLower(getStr("ADMIN_EMAIL", "admin@signflow.local")),
		AdminPassword:      getStr("ADMIN_PASSWORD", "ChangeMe!123"),
		OutboxPollInterval: getDur("OUTBOX_POLL_INTERVAL", 2*time.Second),
	}
}

func getStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getDur(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
