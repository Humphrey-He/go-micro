package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go-micro/pkg/config"
)

type tokenBucket struct {
	tokens float64
	last   time.Time
}

var (
	rlMu      sync.Mutex
	rlBuckets = map[string]*tokenBucket{}
)

func RateLimit() gin.HandlerFunc {
	qps := float64(getInt("RATE_LIMIT_QPS", 100))
	burst := float64(getInt("RATE_LIMIT_BURST", 200))
	if qps <= 0 {
		qps = 100
	}
	if burst <= 0 {
		burst = 200
	}

	return func(c *gin.Context) {
		key := c.ClientIP()
		if !allow(key, qps, burst) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"code": 429, "message": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}

func allow(key string, qps, burst float64) bool {
	now := time.Now()
	rlMu.Lock()
	defer rlMu.Unlock()

	b, ok := rlBuckets[key]
	if !ok {
		rlBuckets[key] = &tokenBucket{tokens: burst - 1, last: now}
		return true
	}
	elapsed := now.Sub(b.last).Seconds()
	b.tokens += elapsed * qps
	if b.tokens > burst {
		b.tokens = burst
	}
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens -= 1
	return true
}

func getInt(key string, def int) int {
	v := config.GetEnv(key, "")
	if v == "" {
		return def
	}
	n := 0
	for i := 0; i < len(v); i++ {
		ch := v[i]
		if ch < '0' || ch > '9' {
			return def
		}
		n = n*10 + int(ch-'0')
	}
	if n == 0 {
		return def
	}
	return n
}
