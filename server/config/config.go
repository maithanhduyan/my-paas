package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Server
	Listen string
	Domain string
	Secret string

	// Database
	DBDriver string // "sqlite" | "postgres"
	DBPath   string // SQLite only
	DBURL    string // PostgreSQL connection URL

	// Redis
	RedisURL     string
	RedisEnabled bool

	// Security
	RateLimitRPS   int
	RateLimitBurst int
	JWTExpiry      time.Duration
	RefreshExpiry  time.Duration
	EncryptionKey  string // hex-encoded 32-byte key

	// Workers
	WorkerCount int
	QueueSize   int

	// Features
	RegistrationOpen bool
	MaintenanceMode  bool

	// Builds
	BuildsDir string
	DataDir   string
}

func Load() *Config {
	c := &Config{
		Listen:         envOr("MYPAAS_LISTEN", ":8080"),
		Domain:         envOr("MYPAAS_DOMAIN", "localhost"),
		Secret:         envOr("MYPAAS_SECRET", ""),
		DBDriver:       envOr("MYPAAS_DB_DRIVER", "sqlite"),
		DBPath:         envOr("MYPAAS_DB", "/data/mypaas.db"),
		DBURL:          envOr("MYPAAS_DB_URL", ""),
		RedisURL:       envOr("MYPAAS_REDIS_URL", ""),
		EncryptionKey:  envOr("MYPAAS_ENCRYPTION_KEY", ""),
		BuildsDir:      envOr("MYPAAS_BUILDS_DIR", "/data/builds"),
		DataDir:        envOr("MYPAAS_DATA_DIR", "/data"),
		RateLimitRPS:   envInt("MYPAAS_RATE_LIMIT_RPS", 100),
		RateLimitBurst: envInt("MYPAAS_RATE_LIMIT_BURST", 200),
		WorkerCount:    envInt("MYPAAS_WORKER_COUNT", 2),
		QueueSize:      envInt("MYPAAS_QUEUE_SIZE", 100),
		JWTExpiry:      envDuration("MYPAAS_JWT_EXPIRY", 24*time.Hour),
		RefreshExpiry:  envDuration("MYPAAS_REFRESH_EXPIRY", 30*24*time.Hour),
		RegistrationOpen: envBool("MYPAAS_REGISTRATION_OPEN", false),
		MaintenanceMode:  envBool("MYPAAS_MAINTENANCE_MODE", false),
	}

	c.RedisEnabled = c.RedisURL != ""

	// Auto-generate secret if not set
	if c.Secret == "" {
		c.Secret = generateRandomHex(32)
	}

	// Auto-generate encryption key if not set
	if c.EncryptionKey == "" {
		c.EncryptionKey = generateRandomHex(32)
	}

	return c
}

func (c *Config) Validate() error {
	if c.DBDriver != "sqlite" && c.DBDriver != "postgres" {
		return fmt.Errorf("MYPAAS_DB_DRIVER must be 'sqlite' or 'postgres', got '%s'", c.DBDriver)
	}
	if c.DBDriver == "postgres" && c.DBURL == "" {
		return fmt.Errorf("MYPAAS_DB_URL is required when MYPAAS_DB_DRIVER=postgres")
	}
	if len(c.EncryptionKey) < 64 {
		return fmt.Errorf("MYPAAS_ENCRYPTION_KEY must be at least 32 bytes (64 hex characters)")
	}
	if c.WorkerCount < 1 {
		return fmt.Errorf("MYPAAS_WORKER_COUNT must be at least 1")
	}
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func generateRandomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
