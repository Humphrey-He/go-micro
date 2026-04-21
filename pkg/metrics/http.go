package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func HTTPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		status := strconv.Itoa(c.Writer.Status())
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		RecordHTTPRequest(c.Request.Method, path, status, duration)
	}
}
