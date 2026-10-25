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
	"ecomm-micro/pkg/config"
)

type responseData struct {
	ResponseCode int
	ResponseBody []byte
}

// pendingCode 标记"首次请求仍在处理中"的哨兵状态码
const pendingCode = -1

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

		idempKey := resolveIdempotencyKey(c)

		ctx := context.Background()
		if existing := idempStore.get(ctx, idempKey); existing != nil {
			replyStored(c, idempKey, existing)
			return
		}

		if !idempStore.setPending(ctx, idempKey) {
			existing := idempStore.get(ctx, idempKey)
			if existing != nil {
				replyStored(c, idempKey, existing)
				return
			}
		}
		c.Writer.Header().Set("X-Idempotency-Key", idempKey)

		buf := &bytes.Buffer{}
		writer := &responseWriter{ResponseWriter: c.Writer, buffer: buf}
		c.Writer = writer

		c.Next()

		// 仅缓存成功响应；失败响应不缓存，允许客户端重试
		if c.Writer.Status() < 400 {
			idempStore.setComplete(ctx, idempKey, c.Writer.Status(), buf.Bytes())
		} else {
			idempStore.del(ctx, idempKey)
		}
	}
}

// resolveIdempotencyKey 幂等键优先级：
// 1. X-Idempotency-Key（客户端显式声明）
// 2. X-Request-Id（网关注入的请求标识，拼接 method+path 防跨接口碰撞）
// 3. method+path+授权头+body 哈希（含调用方身份，防止不同用户同 body 误判重放）
func resolveIdempotencyKey(c *gin.Context) string {
	if k := c.GetHeader("X-Idempotency-Key"); k != "" {
		return k
	}
	if rid := c.GetHeader(HeaderRequestID); rid != "" {
		return c.Request.Method + ":" + c.Request.URL.Path + ":" + rid
	}
	return generateRequestHash(c)
}

// replyStored 重放首次响应；仍在处理中的请求返回 409 让客户端稍后重试。
func replyStored(c *gin.Context, key string, resp *responseData) {
	c.Writer.Header().Set("X-Idempotency-Key", key)
	if resp.ResponseCode == pendingCode {
		c.AbortWithStatusJSON(409, gin.H{"code": 40901, "message": "request with same key is in progress"})
		return
	}
	c.Writer.Header().Set("X-Idempotency-Replayed", "true")
	c.Data(resp.ResponseCode, "application/json", resp.ResponseBody)
	c.Abort()
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
	// 授权头参与哈希：不同用户提交相同 body 不构成重放
	hasher.Write([]byte(c.GetHeader("Authorization")))
	hasher.Write(body)
	return hex.EncodeToString(hasher.Sum(nil))[:32]
}

func (s *idempotencyStore) del(ctx context.Context, key string) {
	s.rdb.Del(ctx, "idempotency:"+key)
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
