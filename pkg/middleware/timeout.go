package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestTimeout returns a middleware that enforces request timeout
// Prevents hanging requests from consuming resources
func RequestTimeout() gin.HandlerFunc {
	defaultTimeout := time.Duration(getEnvInt("REQUEST_TIMEOUT_MS", 30000)) * time.Millisecond

	return func(c *gin.Context) {
		timeout := getTimeoutForPath(c.Request.URL.Path, defaultTimeout)

		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		finished := make(chan struct{})

		go func() {
			c.Next()
			close(finished)
		}()

		select {
		case <-finished:
			// Request completed normally
			return
		case <-ctx.Done():
			// Timeout occurred
			c.AbortWithStatusJSON(504, gin.H{
				"code":    504,
				"message": "request timeout",
				"timeout": timeout.String(),
			})
		}
	}
}

func getTimeoutForPath(path string, defaultTimeout time.Duration) time.Duration {
	// Payment operations - longer timeout
	if hasPrefix(path, "/payments") {
		return 60 * time.Second
	}

	// Order creation - medium timeout
	if hasPrefix(path, "/orders") && hasSuffix(path, "") {
		return 30 * time.Second
	}

	// Inventory operations - short timeout
	if hasPrefix(path, "/inventory") {
		return 10 * time.Second
	}

	// Health checks - very short
	if hasSuffix(path, "/health", "/healthz", "/readyz") {
		return 5 * time.Second
	}

	// Default timeout
	return defaultTimeout
}
