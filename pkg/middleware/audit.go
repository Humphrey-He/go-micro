package middleware

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuditEvent represents an audit log entry
type AuditEvent struct {
	Timestamp   time.Time              `json:"timestamp"`
	RequestID   string                 `json:"request_id"`
	UserID      string                 `json:"user_id"`
	Username    string                 `json:"username"`
	Role        string                 `json:"role"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	ResourceID  string                 `json:"resource_id"`
	Method      string                 `json:"method"`
	Path        string                 `json:"path"`
	QueryParams map[string]string      `json:"query_params,omitempty"`
	StatusCode  int                    `json:"status_code"`
	ClientIP    string                 `json:"client_ip"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AuditLogger defines the interface for audit logging
type AuditLogger interface {
	Log(event *AuditEvent) error
}

// defaultAuditLogger uses zap for audit logging
type defaultAuditLogger struct {
	logger *zap.Logger
}

// Log implements AuditLogger interface
func (l *defaultAuditLogger) Log(event *AuditEvent) error {
	eventJSON, _ := json.Marshal(event)
	l.logger.Info("audit_event",
		zap.String("event", string(eventJSON)),
		zap.String("user_id", event.UserID),
		zap.String("action", event.Action),
		zap.String("resource", event.Resource),
		zap.Int("status", event.StatusCode),
		zap.Duration("duration", event.Duration),
	)
	return nil
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logger *zap.Logger) AuditLogger {
	return &defaultAuditLogger{logger: logger}
}

// Audit returns a middleware that logs admin operations for compliance
// Captures: user actions, resource modifications, query parameters, results
func Audit(action, resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Extract user info from context
		userID, _ := c.Get(CtxUserID)
		username, _ := c.Get(CtxName)
		role, _ := c.Get(CtxRole)
		requestID, _ := c.Get(HeaderRequestID)

		// Parse query params
		queryParams := make(map[string]string)
		for key, values := range c.Request.URL.Query() {
			if len(values) > 0 {
				// Mask sensitive parameters
				if isSensitiveParam(key) {
					queryParams[key] = "***"
				} else {
					queryParams[key] = values[0]
				}
			}
		}

		// Process request
		c.Next()

		// Build audit event
		event := &AuditEvent{
			Timestamp:   start,
			RequestID:   requestID.(string),
			UserID:      toString(userID),
			Username:    toString(username),
			Role:        toString(role),
			Action:      action,
			Resource:    resource,
			ResourceID:  extractResourceID(c),
			Method:      c.Request.Method,
			Path:        c.Request.URL.Path,
			QueryParams: queryParams,
			StatusCode:  c.Writer.Status(),
			ClientIP:    c.ClientIP(),
			UserAgent:   c.GetHeader("User-Agent"),
			Duration:    time.Since(start),
		}

		// Add error if present
		if len(c.Errors) > 0 {
			event.Error = c.Errors.String()
		}

		// Log the event
		logAuditEvent(event)

		// Add audit headers
		c.Writer.Header().Set("X-Audit-Request-ID", requestID.(string))
	}
}

// AdminAudit is a specialized middleware for admin panel operations
func AdminAudit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip GET requests for audit (unless it's a sensitive resource)
		if c.Request.Method == "GET" && !isSensitivePath(c.Request.URL.Path) {
			c.Next()
			return
		}

		start := time.Now()

		userID, _ := c.Get(CtxUserID)
		username, _ := c.Get(CtxName)
		role, _ := c.Get(CtxRole)
		requestID, _ := c.Get(HeaderRequestID)

		c.Next()

		// Only log mutations
		if c.Request.Method != "GET" {
			event := &AuditEvent{
				Timestamp:   start,
				RequestID:   requestID.(string),
				UserID:      toString(userID),
				Username:    toString(username),
				Role:        toString(role),
				Action:      classifyAction(c.Request.Method),
				Resource:    classifyResource(c.Request.URL.Path),
				ResourceID:  extractResourceID(c),
				Method:      c.Request.Method,
				Path:        c.Request.URL.Path,
				StatusCode:  c.Writer.Status(),
				ClientIP:    c.ClientIP(),
				UserAgent:   c.GetHeader("User-Agent"),
				Duration:    time.Since(start),
			}

			if errMsg := c.Errors.String(); errMsg != "" {
				event.Error = errMsg
			}

			logAuditEvent(event)
		}
	}
}

func logAuditEvent(event *AuditEvent) {
	// Log as structured JSON for easy parsing
	eventJSON, _ := json.Marshal(event)
	zap.L().Info("audit_event",
		zap.String("event", string(eventJSON)),
		zap.String("user_id", event.UserID),
		zap.String("action", event.Action),
		zap.String("resource", event.Resource),
		zap.Int("status", event.StatusCode),
		zap.Duration("duration", event.Duration),
	)
}

func extractResourceID(c *gin.Context) string {
	// Try to extract ID from path params
	for _, param := range []string{"id", "order_id", "payment_id", "user_id", "sku_id"} {
		if val := c.Param(param); val != "" {
			return val
		}
	}
	// Try to extract from query
	for _, param := range []string{"id", "order_id", "payment_id"} {
		if val := c.Query(param); val != "" {
			return val
		}
	}
	return ""
}

func isSensitiveParam(key string) bool {
	sensitive := []string{"password", "token", "secret", "key", "authorization", "api_key", "apikey"}
	for _, s := range sensitive {
		if key == s || contains(key, s) {
			return true
		}
	}
	return false
}

func isSensitivePath(path string) bool {
	sensitivePaths := []string{"/admin/users", "/admin/config", "/admin/secrets"}
	for _, sp := range sensitivePaths {
		if contains(path, sp) {
			return true
		}
	}
	return false
}

func classifyAction(method string) string {
	switch method {
	case "POST":
		return "CREATE"
	case "PUT", "PATCH":
		return "UPDATE"
	case "DELETE":
		return "DELETE"
	case "GET":
		return "READ"
	default:
		return method
	}
}

func classifyResource(path string) string {
	if contains(path, "/orders") {
		return "order"
	}
	if contains(path, "/payments") {
		return "payment"
	}
	if contains(path, "/inventory") {
		return "inventory"
	}
	if contains(path, "/users") {
		return "user"
	}
	if contains(path, "/refund") {
		return "refund"
	}
	if contains(path, "/activity") {
		return "activity"
	}
	if contains(path, "/price") {
		return "price"
	}
	return "unknown"
}
