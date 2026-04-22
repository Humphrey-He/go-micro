package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	CircuitClosed CircuitBreakerState = iota
	CircuitOpen
	CircuitHalfOpen
)

func (s CircuitBreakerState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half_open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	FailureThreshold int           // failures before opening circuit
	SuccessThreshold  int           // successes in half-open to close
	OpenDuration      time.Duration // how long circuit stays open
}

// circuitBreaker implements the circuit breaker pattern
type circuitBreaker struct {
	state            CircuitBreakerState
	failures         int
	successes       int
	lastFailureTime  time.Time
	config           CircuitBreakerConfig
	mu               sync.RWMutex
}

// CircuitBreaker returns a middleware that protects against cascading failures
// Key features:
// - Per-endpoint or per-service circuit breaking
// - Configurable failure thresholds
// - Automatic recovery attempt (half-open state)
func CircuitBreaker(serviceName string) gin.HandlerFunc {
	config := CircuitBreakerConfig{
		FailureThreshold: getEnvInt("CIRCUIT_BREAKER_FAILURES", 5),
		SuccessThreshold: getEnvInt("CIRCUIT_BREAKER_SUCCESSES", 2),
		OpenDuration:     time.Duration(getEnvInt("CIRCUIT_BREAKER_OPEN_DURATION_MS", 30000)) * time.Millisecond,
	}

	cb := &circuitBreaker{
		state:  CircuitClosed,
		config: config,
	}

	return func(c *gin.Context) {
		key := serviceName + ":"+c.FullPath()

		if !cb.allow(key) {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"code":    503,
				"message": "service temporarily unavailable",
				"circuit": cb.getState(),
			})
			return
		}

		// Process request
		c.Next()

		// Record result
		status := c.Writer.Status()
		if status >= 500 {
			cb.recordFailure()
		} else if status < 400 {
			cb.recordSuccess()
		}
	}
}

func (cb *circuitBreaker) allow(key string) bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true

	case CircuitOpen:
		if time.Since(cb.lastFailureTime) > cb.config.OpenDuration {
			cb.state = CircuitHalfOpen
			cb.successes = 0
			return true
		}
		return false

	case CircuitHalfOpen:
		return true
	}

	return true
}

func (cb *circuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailureTime = time.Now()

	if cb.state == CircuitHalfOpen {
		cb.state = CircuitOpen
	} else if cb.failures >= cb.config.FailureThreshold {
		cb.state = CircuitOpen
	}
}

func (cb *circuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.successes++

	if cb.state == CircuitHalfOpen && cb.successes >= cb.config.SuccessThreshold {
		cb.state = CircuitClosed
		cb.failures = 0
	}
}

func (cb *circuitBreaker) getState() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state.String()
}
