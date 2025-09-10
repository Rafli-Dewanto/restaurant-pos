package middleware

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type RateLimitConfig struct {
	// Max number of requests allowed per window
	Max int
	// Duration of the rate limiting window
	Duration time.Duration
	// Key generator function to identify clients
	KeyGenerator func(c *fiber.Ctx) string
	// Handler to call when rate limit is exceeded
	LimitReached func(c *fiber.Ctx) error
	// Skip middleware based on condition
	Skip func(c *fiber.Ctx) bool
	// Logger for rate limit events
	Logger *logrus.Logger
}

type ClientInfo struct {
	Count     int
	ResetTime time.Time
	mutex     sync.RWMutex
}

type RateLimiter struct {
	clients map[string]*ClientInfo
	mutex   sync.RWMutex
	config  RateLimitConfig
	cleanup *time.Ticker
}

func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	// Set defaults
	if config.Max == 0 {
		config.Max = 100
	}
	if config.Duration == 0 {
		config.Duration = time.Hour
	}
	if config.KeyGenerator == nil {
		config.KeyGenerator = func(c *fiber.Ctx) string {
			return c.IP()
		}
	}
	if config.LimitReached == nil {
		config.LimitReached = func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Rate limit exceeded",
				"message": "Too many requests, please try again later",
			})
		}
	}

	rl := &RateLimiter{
		clients: make(map[string]*ClientInfo),
		config:  config,
	}

	// Start cleanup routine to remove expired entries
	rl.cleanup = time.NewTicker(config.Duration)
	go rl.cleanupExpired()

	return rl
}

func (rl *RateLimiter) cleanupExpired() {
	for range rl.cleanup.C {
		now := time.Now()
		rl.mutex.Lock()
		for key, client := range rl.clients {
			client.mutex.RLock()
			if now.After(client.ResetTime) {
				delete(rl.clients, key)
			}
			client.mutex.RUnlock()
		}
		rl.mutex.Unlock()
	}
}

func (rl *RateLimiter) Stop() {
	if rl.cleanup != nil {
		rl.cleanup.Stop()
	}
}

func (rl *RateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip if condition is met
		if rl.config.Skip != nil && rl.config.Skip(c) {
			return c.Next()
		}

		key := rl.config.KeyGenerator(c)
		now := time.Now()

		// Get or create client info
		rl.mutex.Lock()
		client, exists := rl.clients[key]
		if !exists {
			client = &ClientInfo{
				Count:     0,
				ResetTime: now.Add(rl.config.Duration),
			}
			rl.clients[key] = client
		}
		rl.mutex.Unlock()

		client.mutex.Lock()
		defer client.mutex.Unlock()

		// Reset if window has expired
		if now.After(client.ResetTime) {
			client.Count = 0
			client.ResetTime = now.Add(rl.config.Duration)
		}

		// Check if limit is exceeded
		if client.Count >= rl.config.Max {
			// Set rate limit headers
			c.Set("X-RateLimit-Limit", strconv.Itoa(rl.config.Max))
			c.Set("X-RateLimit-Remaining", "0")
			c.Set("X-RateLimit-Reset", strconv.FormatInt(client.ResetTime.Unix(), 10))

			// Log rate limit exceeded
			if rl.config.Logger != nil {
				rl.config.Logger.Warnf("Rate limit exceeded for client: %s", key)
			}

			return rl.config.LimitReached(c)
		}

		// Increment counter
		client.Count++

		// Set rate limit headers
		c.Set("X-RateLimit-Limit", strconv.Itoa(rl.config.Max))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(rl.config.Max-client.Count))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(client.ResetTime.Unix(), 10))

		return c.Next()
	}
}

// Predefined rate limit configurations

// BasicRateLimit creates a basic rate limiter (100 requests per hour)
func BasicRateLimit(logger *logrus.Logger) fiber.Handler {
	rl := NewRateLimiter(RateLimitConfig{
		Max:      100,
		Duration: time.Hour,
		Logger:   logger,
	})
	return rl.Middleware()
}

// StrictRateLimit creates a strict rate limiter (20 requests per minute)
func StrictRateLimit(logger *logrus.Logger) fiber.Handler {
	rl := NewRateLimiter(RateLimitConfig{
		Max:      20,
		Duration: time.Minute,
		Logger:   logger,
	})
	return rl.Middleware()
}

// AuthRateLimit creates rate limiting for auth endpoints
func AuthRateLimit(logger *logrus.Logger) fiber.Handler {
	rl := NewRateLimiter(RateLimitConfig{
		Max:      5, // 5 login/register attempts per 15 minutes
		Duration: 15 * time.Minute,
		Logger:   logger,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Too many authentication attempts",
				"message": "Please wait before trying again",
			})
		},
	})
	return rl.Middleware()
}

// PaymentRateLimit creates rate limiting for payment endpoints
func PaymentRateLimit(logger *logrus.Logger) fiber.Handler {
	rl := NewRateLimiter(RateLimitConfig{
		Max:      10, // 10 payment requests per hour
		Duration: time.Hour,
		Logger:   logger,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Payment rate limit exceeded",
				"message": "Too many payment requests, please contact support if you need assistance",
			})
		},
	})
	return rl.Middleware()
}

// UserBasedRateLimit creates a rate limiter based on user ID
func UserBasedRateLimit(max int, duration time.Duration, logger *logrus.Logger) fiber.Handler {
	rl := NewRateLimiter(RateLimitConfig{
		Max:      max,
		Duration: duration,
		Logger:   logger,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Try to get user ID from JWT claims
			if userID := c.Locals("userID"); userID != nil {
				return fmt.Sprintf("user:%v", userID)
			}
			// Fallback to IP if no user ID
			return c.IP()
		},
	})
	return rl.Middleware()
}

// IPBasedRateLimit creates a rate limiter based on IP address
func IPBasedRateLimit(max int, duration time.Duration, logger *logrus.Logger) fiber.Handler {
	rl := NewRateLimiter(RateLimitConfig{
		Max:      max,
		Duration: duration,
		Logger:   logger,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	})
	return rl.Middleware()
}
