package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/my-paas/server/cache"
	"github.com/my-paas/server/config"
)

// RateLimiter provides rate limiting with optional Redis backend.
// Falls back to in-memory sliding window when Redis is unavailable.
func RateLimiter(cfg *config.Config, redis *cache.Redis) fiber.Handler {
	if redis != nil {
		return redisRateLimiter(cfg, redis)
	}
	return memoryRateLimiter(cfg)
}

func redisRateLimiter(cfg *config.Config, redis *cache.Redis) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.IP()
		if userID, ok := c.Locals("user_id").(string); ok && userID != "" {
			key = "user:" + userID
		}

		allowed, remaining, err := redis.CheckRateLimit(c.Context(), key, cfg.RateLimitRPS, time.Second)
		if err != nil {
			// If Redis fails, allow the request
			return c.Next()
		}

		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.RateLimitRPS))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if !allowed {
			c.Set("Retry-After", "1")
			return c.Status(429).JSON(fiber.Map{"error": "rate limit exceeded"})
		}
		return c.Next()
	}
}

// In-memory rate limiter using sliding window
type memLimiter struct {
	mu      sync.Mutex
	windows map[string]*slidingWindow
}

type slidingWindow struct {
	count   int
	resetAt time.Time
}

func memoryRateLimiter(cfg *config.Config) fiber.Handler {
	limiter := &memLimiter{
		windows: make(map[string]*slidingWindow),
	}

	// Cleanup goroutine
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			limiter.mu.Lock()
			now := time.Now()
			for k, w := range limiter.windows {
				if now.After(w.resetAt) {
					delete(limiter.windows, k)
				}
			}
			limiter.mu.Unlock()
		}
	}()

	return func(c *fiber.Ctx) error {
		key := c.IP()
		now := time.Now()

		limiter.mu.Lock()
		w, exists := limiter.windows[key]
		if !exists || now.After(w.resetAt) {
			w = &slidingWindow{count: 0, resetAt: now.Add(time.Second)}
			limiter.windows[key] = w
		}
		w.count++
		count := w.count
		limiter.mu.Unlock()

		remaining := cfg.RateLimitRPS - count
		if remaining < 0 {
			remaining = 0
		}

		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.RateLimitRPS))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if count > cfg.RateLimitBurst {
			c.Set("Retry-After", "1")
			return c.Status(429).JSON(fiber.Map{"error": "rate limit exceeded"})
		}
		return c.Next()
	}
}

// SecurityHeaders adds standard security headers for enterprise deployments.
func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		if c.Protocol() == "https" {
			c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		return c.Next()
	}
}

// RequestID adds a unique request ID to each request for tracing.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get("X-Request-ID")
		if id == "" {
			id = fmt.Sprintf("%d", time.Now().UnixNano())
		}
		c.Set("X-Request-ID", id)
		c.Locals("request_id", id)
		return c.Next()
	}
}
