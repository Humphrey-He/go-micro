package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go-micro/pkg/config"
)

// getEnvInt parses an environment variable to int with default fallback
func getEnvInt(key string, def int) int {
	v := config.GetEnv(key, "")
	if v == "" {
		return def
	}
	n := 0
	for i := 0; i < len(v); i++ {
		if v[i] < '0' || v[i] > '9' {
			return def
		}
		n = n*10 + int(v[i]-'0')
	}
	if n == 0 {
		return def
	}
	return n
}

// splitAndTrim splits a string by separator and trims whitespace
func splitAndTrim(s, sep string) []string {
	var result []string
	parts := strings.Split(s, sep)
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// containsAny checks if s contains any of the substrings
func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// parseInt parses a string to int
func parseInt(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			n = n*10 + int(s[i]-'0')
		}
	}
	return n
}

// hasSuffix checks if s ends with any of the suffixes
func hasSuffix(s string, suffixes ...string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}
	return false
}

// hasPrefix checks if s starts with any of the prefixes
func hasPrefix(s string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

// toString converts an interface to string
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// getClientIP extracts client IP from request
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For first (for proxied requests)
	xff := c.GetHeader("X-Forwarded-For")
	if xff != "" {
		ips := splitAndTrim(xff, ",")
		if len(ips) > 0 {
			return ips[0]
		}
	}

	// Check X-Real-IP
	xri := c.GetHeader("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to remote address
	return c.ClientIP()
}

// getRequestID extracts request ID from context
func getRequestID(c *gin.Context) string {
	if rid, exists := c.Get(HeaderRequestID); exists {
		if s, ok := rid.(string); ok {
			return s
		}
	}
	return ""
}

// futureDate returns RFC1123 date string for future time
func futureDate(d time.Duration) string {
	return time.Now().Add(d).UTC().Format(time.RFC1123)
}

// trim removes leading and trailing whitespace
func trim(s string) string {
	return strings.TrimSpace(s)
}
