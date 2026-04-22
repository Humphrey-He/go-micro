package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go-micro/pkg/config"
)

// CacheControl provides caching directives for client and CDN caching
// Supports: browser cache, CDN cache, cache invalidation patterns
func CacheControl() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		// Only apply to GET requests
		if method != "GET" {
			c.Next()
			return
		}

		// Determine cache policy based on resource type
		cachePolicy := getCachePolicy(path)

		c.Writer.Header().Set("Cache-Control", cachePolicy.Directive)
		c.Writer.Header().Set("Expires", cachePolicy.Expires)
		c.Writer.Header().Set("Vary", "Accept-Encoding")

		if cachePolicy.ETag {
			// Generate simple ETag based on path and timestamp
			etag := generateETag(path)
			c.Writer.Header().Set("ETag", etag)

			// Check If-None-Match
			if c.GetHeader("If-None-Match") == etag {
				c.AbortWithStatus(304)
				return
			}
		}

		c.Next()
	}
}

// CachePolicy defines caching behavior
type CachePolicy struct {
	Directive string
	Expires   string
	ETag      bool
}

func getCachePolicy(path string) CachePolicy {
	// Static assets - long cache
	if hasSuffix(path, ".css", ".js", ".svg", ".png", ".jpg", ".jpeg", ".gif", ".ico") {
		return CachePolicy{
			Directive: "public, max-age=31536000, immutable",
			Expires:   futureDate(365 * 24 * time.Hour),
			ETag:      false,
		}
	}

	// Admin dashboard - no cache
	if hasPrefix(path, "/admin/dashboard") {
		return CachePolicy{
			Directive: "no-cache, no-store, must-revalidate",
			Expires:   "0",
			ETag:      false,
		}
	}

	// Order/payment data - short cache or no cache
	if hasSuffix(path, "/orders", "/payments", "/refunds") {
		return CachePolicy{
			Directive: "private, no-cache, max-age=0",
			Expires:   "0",
			ETag:      true,
		}
	}

	// Inventory - short cache
	if hasPrefix(path, "/inventory") {
		return CachePolicy{
			Directive: "public, max-age=30, stale-while-revalidate=60",
			Expires:   futureDate(30 * time.Second),
			ETag:      true,
		}
	}

	// Health/readiness - no cache
	if hasSuffix(path, "/health", "/healthz", "/readyz", "/ready") {
		return CachePolicy{
			Directive: "no-cache, no-store, must-revalidate",
			Expires:   "0",
			ETag:      false,
		}
	}

	// Default - short cache
	maxAge := config.GetEnv("CACHE_DEFAULT_MAX_AGE", "60")
	maxAgeSec, _ := strconv.Atoi(maxAge)
	if maxAgeSec <= 0 {
		maxAgeSec = 60
	}

	return CachePolicy{
		Directive: "public, max-age=" + maxAge + ", stale-while-revalidate=" + strconv.Itoa(maxAgeSec*2),
		Expires:   futureDate(time.Duration(maxAgeSec) * time.Second),
		ETag:      true,
	}
}

func generateETag(path string) string {
	// Simple ETag - in production use content hash
	return `"` + path + `-` + strconv.FormatInt(time.Now().Unix(), 10) + `"`
}
