package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// BusinessRateLimitConfig holds rate limit configuration per resource
type BusinessRateLimitConfig struct {
	Order   RateLimitRule
	Payment RateLimitRule
	General RateLimitRule
}

// RateLimitRule defines rate limit for a specific resource
type RateLimitRule struct {
	Window          time.Duration
	MaxRequests     int
	MaxRequestsUser int // per user
}

// businessRateLimiter implements per-resource and per-user rate limiting
type businessRateLimiter struct {
	mu      sync.RWMutex
	counters map[string]*userCounter
	config   BusinessRateLimitConfig
}

type userCounter struct {
	total    int
	byUser   map[string]int
	windowStart time.Time
}

func newBusinessRateLimiter() *businessRateLimiter {
	config := BusinessRateLimitConfig{
		Order: RateLimitRule{
			Window:          60 * time.Second,
			MaxRequests:     100,
			MaxRequestsUser: 10,
		},
		Payment: RateLimitRule{
			Window:          60 * time.Second,
			MaxRequests:     50,
			MaxRequestsUser: 5,
		},
		General: RateLimitRule{
			Window:          60 * time.Second,
			MaxRequests:     200,
			MaxRequestsUser: 20,
		},
	}

	return &businessRateLimiter{
		counters: make(map[string]*userCounter),
		config:   config,
	}
}

var bizRateLimiter = newBusinessRateLimiter()

// BusinessRateLimit returns a middleware that enforces business-specific rate limits
// Order creation: max 10 per user/minute, 100 total/minute
// Payment processing: max 5 per user/minute, 50 total/minute
func BusinessRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		resource := classifyBusinessResource(c.Request.URL.Path)
		rule := getBusinessRule(resource)

		userID, _ := c.Get(CtxUserID)
		userIDStr := toString(userID)
		if userIDStr == "" {
			userIDStr = c.ClientIP()
		}

		key := resource + ":" + c.ClientIP()
		counter := bizRateLimiter.getOrCreate(key)

		bizRateLimiter.mu.Lock()
		now := time.Now()

		// Reset window if expired
		if now.Sub(counter.windowStart) > rule.Window {
			counter.total = 0
			counter.byUser = make(map[string]int)
			counter.windowStart = now
		}

		// Check global limit
		if counter.total >= rule.MaxRequests {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "rate limit exceeded for " + resource,
				"retry_after": int(rule.Window.Seconds()),
			})
			bizRateLimiter.mu.Unlock()
			return
		}

		// Check per-user limit
		if counter.byUser[userIDStr] >= rule.MaxRequestsUser {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "rate limit exceeded for user",
				"retry_after": int(rule.Window.Seconds()),
			})
			bizRateLimiter.mu.Unlock()
			return
		}

		// Increment counters
		counter.total++
		counter.byUser[userIDStr]++

		c.Writer.Header().Set("X-RateLimit-Limit", strconv.Itoa(rule.MaxRequests))
		c.Writer.Header().Set("X-RateLimit-Remaining", strconv.Itoa(rule.MaxRequests-counter.total))
		c.Writer.Header().Set("X-RateLimit-Window", rule.Window.String())

		bizRateLimiter.mu.Unlock()
		c.Next()
	}
}

// OrderCreationLimit specifically limits order creation attempts
// Prevents cart abandonment abuse and order spam
func OrderCreationLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get(CtxUserID)
		userIDStr := toString(userID)
		if userIDStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "authentication required",
			})
			return
		}

		key := "order_create:" + userIDStr
		limit := 10 // max 10 orders per minute
		window := 60 * time.Second

		bizRateLimiter.mu.Lock()
		counter := bizRateLimiter.counters[key]
		if counter == nil {
			counter = &userCounter{
				byUser:     make(map[string]int),
				windowStart: time.Now(),
			}
			bizRateLimiter.counters[key] = counter
		}

		now := time.Now()
		if now.Sub(counter.windowStart) > window {
			counter.total = 0
			counter.byUser = make(map[string]int)
			counter.windowStart = now
		}

		userCount := counter.byUser[userIDStr]
		if userCount >= limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":         429,
				"message":      "order creation rate limit exceeded",
				"retry_after":  int(window.Seconds()),
			})
			bizRateLimiter.mu.Unlock()
			return
		}

		counter.byUser[userIDStr]++
		counter.total++
		bizRateLimiter.mu.Unlock()

		c.Next()
	}
}

// PaymentRateLimit prevents payment API abuse
// Critical for financial security
func PaymentRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get(CtxUserID)
		userIDStr := toString(userID)
		if userIDStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "authentication required",
			})
			return
		}

		key := "payment:" + userIDStr
		limit := 5 // max 5 payments per minute
		window := 60 * time.Second

		bizRateLimiter.mu.Lock()
		counter := bizRateLimiter.counters[key]
		if counter == nil {
			counter = &userCounter{
				byUser:     make(map[string]int),
				windowStart: time.Now(),
			}
			bizRateLimiter.counters[key] = counter
		}

		now := time.Now()
		if now.Sub(counter.windowStart) > window {
			counter.total = 0
			counter.byUser = make(map[string]int)
			counter.windowStart = now
		}

		userCount := counter.byUser[userIDStr]
		if userCount >= limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":         429,
				"message":      "payment rate limit exceeded",
				"retry_after":  int(window.Seconds()),
			})
			bizRateLimiter.mu.Unlock()
			return
		}

		counter.byUser[userIDStr]++
		counter.total++
		bizRateLimiter.mu.Unlock()

		c.Next()
	}
}

func (b *businessRateLimiter) getOrCreate(key string) *userCounter {
	b.mu.RLock()
	if c, ok := b.counters[key]; ok {
		b.mu.RUnlock()
		return c
	}
	b.mu.RUnlock()

	b.mu.Lock()
	defer b.mu.Unlock()
	if c, ok := b.counters[key]; ok {
		return c
	}
	c := &userCounter{
		byUser:      make(map[string]int),
		windowStart: time.Now(),
	}
	b.counters[key] = c
	return c
}

func classifyBusinessResource(path string) string {
	if contains(path, "/orders") {
		return "order"
	}
	if contains(path, "/payments") {
		return "payment"
	}
	if contains(path, "/refund") {
		return "refund"
	}
	if contains(path, "/inventory") {
		return "inventory"
	}
	return "general"
}

func getBusinessRule(resource string) RateLimitRule {
	switch resource {
	case "order":
		return bizRateLimiter.config.Order
	case "payment":
		return bizRateLimiter.config.Payment
	default:
		return bizRateLimiter.config.General
	}
}
