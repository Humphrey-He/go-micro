package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go-micro/pkg/config"
)

// idempotencyEntry stores the state of an idempotency key
type idempotencyEntry struct {
	ResponseCode int
	ResponseBody []byte
	CreatedAt    time.Time
	Expiry       time.Time
}

// idempotencyStore holds idempotency keys in memory
type idempotencyStore struct {
	mu     sync.RWMutex
	keys   map[string]*idempotencyEntry
	ttl    time.Duration
	clean  chan struct{}
}

var (
	idempStore = &idempotencyStore{
		keys:  make(map[string]*idempotencyEntry),
		ttl:   24 * time.Hour,
		clean: make(chan struct{}),
	}
)

// Idempotency returns a middleware that ensures operations are idempotent
// Prevents duplicate order creation, payment processing, etc.
func Idempotency() gin.HandlerFunc {
	ttlSeconds := config.GetEnv("IDEMPOTENCY_TTL_SECONDS", "86400")
	ttl := time.Duration(parseInt(ttlSeconds)) * time.Second
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	idempStore.ttl = ttl

	go idempStore.cleanupExpired()

	return func(c *gin.Context) {
		method := c.Request.Method
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			c.Next()
			return
		}

		idempKey := c.GetHeader("X-Idempotency-Key")
		if idempKey == "" {
			idempKey = generateRequestHash(c)
		}

		if existing := idempStore.get(idempKey); existing != nil {
			c.Writer.Header().Set("X-Idempotency-Key", idempKey)
			c.Writer.Header().Set("X-Idempotency-Replayed", "true")
			c.Data(existing.ResponseCode, "application/json", existing.ResponseBody)
			c.Abort()
			return
		}

		idempStore.setPending(idempKey)
		c.Writer.Header().Set("X-Idempotency-Key", idempKey)

		// Create response buffer to capture response
		buf := &bytes.Buffer{}
		writer := &responseWriter{ResponseWriter: c.Writer, buffer: buf}
		c.Writer = writer

		c.Next()

		if c.Writer.Status() < 400 {
			idempStore.setComplete(idempKey, c.Writer.Status(), buf.Bytes())
		}
	}
}

type responseWriter struct {
	gin.ResponseWriter
	buffer *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.buffer.Write(b)
	return w.ResponseWriter.Write(b)
}

func generateRequestHash(c *gin.Context) string {
	body, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	hasher := sha256.New()
	hasher.Write([]byte(c.Request.Method))
	hasher.Write([]byte(c.Request.URL.Path))
	hasher.Write(body)
	return hex.EncodeToString(hasher.Sum(nil))[:32]
}

func (s *idempotencyStore) get(key string) *idempotencyEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if entry, ok := s.keys[key]; ok {
		if time.Now().Before(entry.Expiry) {
			return entry
		}
	}
	return nil
}

func (s *idempotencyStore) setPending(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.keys[key] = &idempotencyEntry{
		CreatedAt: time.Now(),
		Expiry:    time.Now().Add(s.ttl),
	}
}

func (s *idempotencyStore) setComplete(key string, code int, body []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entry, ok := s.keys[key]; ok {
		entry.ResponseCode = code
		entry.ResponseBody = make([]byte, len(body))
		copy(entry.ResponseBody, body)
	}
}

func (s *idempotencyStore) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			now := time.Now()
			for key, entry := range s.keys {
				if now.After(entry.Expiry) {
					delete(s.keys, key)
				}
			}
			s.mu.Unlock()
		case <-s.clean:
			return
		}
	}
}
