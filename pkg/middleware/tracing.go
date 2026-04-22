package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Tracing provides distributed tracing for observability
// Supports OpenTelemetry-style trace context propagation
func Tracing(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		ctx := c.Request.Context()

		// Extract trace context from headers
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = c.GetHeader("X-Request-Id")
		}
		spanID := generateSpanID()
		parentSpanID := c.GetHeader("X-Parent-Span-ID")

		// Add trace context to response headers
		c.Writer.Header().Set("X-Trace-ID", traceID)
		c.Writer.Header().Set("X-Span-ID", spanID)

		// Create span context for downstream propagation
		spanCtx := newSpanContext(traceID, spanID, parentSpanID)
		c.Set("span_ctx", spanCtx)

		// Process request
		c.Next()

		// Record span completion
		latency := time.Since(start)
		status := c.Writer.Status()

		// Log with trace context
		zap.L().Info("span_completed",
			zap.String("service", serviceName),
			zap.String("trace_id", traceID),
			zap.String("span_id", spanID),
			zap.String("parent_span_id", parentSpanID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.Int("response_size", c.Writer.Size()),
		)

		// Store span in context for propagation
		ctx = context.WithValue(ctx, "span_end", time.Now())
		c.Request = c.Request.WithContext(ctx)
	}
}

type spanContext struct {
	TraceID      string
	SpanID       string
	ParentSpanID string
}

func newSpanContext(traceID, spanID, parentSpanID string) *spanContext {
	return &spanContext{
		TraceID:      traceID,
		SpanID:       spanID,
		ParentSpanID: parentSpanID,
	}
}

func generateSpanID() string {
	// Simple span ID generation
	b := make([]byte, 8)
	for i := range b {
		b[i] = byte(time.Now().UnixNano() >> uint(i*8) & 0xff)
	}
	return fmt.Sprintf("%x", b)
}

func generateID(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = byte(time.Now().UnixNano() >> uint(i*8) & 0xff)
	}
	return fmt.Sprintf("%x", b)
}
