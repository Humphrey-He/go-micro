package middleware

import (
	"github.com/gin-gonic/gin"
	"go-micro/pkg/config"
)

// CORS returns a middleware that handles Cross-Origin Resource Sharing
// Supports e-commerce scenarios: web admin, mobile apps, third-party integrations
func CORS() gin.HandlerFunc {
	allowedOrigins := config.GetEnv("CORS_ALLOWED_ORIGINS", "*")
	allowedMethods := config.GetEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,PATCH,OPTIONS")
	allowedHeaders := config.GetEnv("CORS_ALLOWED_HEADERS", "Content-Type,Authorization,X-Request-Id,X-Trace-ID")
	exposeHeaders := config.GetEnv("CORS_EXPOSE_HEADERS", "X-Request-Id,X-Trace-ID,X-Span-ID")
	maxAge := config.GetEnv("CORS_MAX_AGE", "86400")
	allowCredentials := config.GetEnv("CORS_ALLOW_CREDENTIALS", "true")

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Check if origin is allowed (supports wildcard)
		if allowedOrigins != "*" && !isOriginAllowed(origin, allowedOrigins) {
			c.AbortWithStatusJSON(403, gin.H{
				"code":    403,
				"message": "origin not allowed",
			})
			return
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", allowedMethods)
		c.Writer.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
		c.Writer.Header().Set("Access-Control-Expose-Headers", exposeHeaders)
		c.Writer.Header().Set("Access-Control-Max-Age", maxAge)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", allowCredentials)

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func isOriginAllowed(origin, allowedOrigins string) bool {
	// Simple check - in production use a more robust method
	origins := splitAndTrim(allowedOrigins, ",")
	for _, o := range origins {
		if o == origin || o == "*" {
			return true
		}
	}
	return false
}
