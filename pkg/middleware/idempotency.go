package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go-micro/pkg/config"
)

type responseData struct {
	ResponseCode int
	ResponseBody []byte
}

type idempotencyStore struct {
	rdb *redis.Client
	ttl time.Duration
}

var (
	idempStore *idempotencyStore
	redisOnce  bool
)

func InitIdempotency(rdb *redis.Client) {
	idempStore = &idempotencyStore{
		rdb: rdb,
		ttl: 24 * time.Hour,
	}
}

func Idempotency() gin.HandlerFunc {
	if idempStore == nil || idempStore.rdb == nil {
		// Fallback to no-op if Redis not initialized
		return func(c *gin.Context) {
			c.Next()
		}
	}

	ttlSeconds := config.GetEnv("IDEMPOTENCY_TTL_SECONDS", "86400")
	ttl := time.Duration(parseInt(ttlSeconds)) * time.Second
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	idempStore.ttl = ttl

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

		ctx := context.Background()
		if existing := idempStore.get(ctx, idempKey); existing != nil {
			c.Writer.Header().Set("X-Idempotency-Key", idempKey)
			c.Writer.Header().Set("X-Idempotency-Replayed", "true")
			c.Data(existing.ResponseCode, "application/json", existing.ResponseBody)
			c.Abort()
			return
		}

		if !idempStore.setPending(ctx, idempKey) {
			existing := idempStore.get(ctx, idempKey)
			if existing != nil {
				c.Writer.Header().Set("X-Idempotency-Key", idempKey)
				c.Writer.Header().Set("X-Idempotency-Replayed", "true")
				c.Data(existing.ResponseCode, "application/json", existing.ResponseBody)
				c.Abort()
				return
			}
		}
		c.Writer.Header().Set("X-Idempotency-Key", idempKey)

		buf := &bytes.Buffer{}
		writer := &responseWriter{ResponseWriter: c.Writer, buffer: buf}
		c.Writer = writer

		c.Next()

		if c.Writer.Status() < 400 {
			idempStore.setComplete(ctx, idempKey, c.Writer.Status(), buf.Bytes())
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

func (s *idempotencyStore) get(ctx context.Context, key string) *responseData {
	redisKey := "idempotency:" + key
	data, err := s.rdb.Get(ctx, redisKey).Bytes()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return nil
	}
	var resp responseData
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil
	}
	return &resp
}

func (s *idempotencyStore) setPending(ctx context.Context, key string) bool {
	redisKey := "idempotency:" + key
	pending := responseData{ResponseCode: -1}
	data, _ := json.Marshal(pending)
	ok, err := s.rdb.SetNX(ctx, redisKey, data, s.ttl).Result()
	return ok && err == nil
}

func (s *idempotencyStore) setComplete(ctx context.Context, key string, code int, body []byte) {
	redisKey := "idempotency:" + key
	resp := responseData{ResponseCode: code, ResponseBody: body}
	data, _ := json.Marshal(resp)
	s.rdb.Set(ctx, redisKey, data, s.ttl)
}
