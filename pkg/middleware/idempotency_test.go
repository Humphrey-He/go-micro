package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func newIdempRouter(t *testing.T) *gin.Engine {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		InitIdempotency(nil)
		rdb.Close()
	})
	InitIdempotency(rdb)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func do(r *gin.Engine, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// R2 验收（单测版）：同一请求打 5 次 → 业务执行 1 次、5 次同响应。
func TestIdempotency_ReplayFiveTimesSameResponse(t *testing.T) {
	r := newIdempRouter(t)
	var calls int64
	r.POST("/api/v1/orders", Idempotency(), func(c *gin.Context) {
		atomic.AddInt64(&calls, 1)
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"order_id": "O-1"}})
	})

	headers := map[string]string{"X-Idempotency-Key": "order-key-1"}
	var firstBody string
	for i := 0; i < 5; i++ {
		w := do(r, http.MethodPost, "/api/v1/orders", `{"sku_id":"S","qty":1}`, headers)
		if w.Code != http.StatusOK {
			t.Fatalf("attempt %d: status %d", i+1, w.Code)
		}
		if i == 0 {
			firstBody = w.Body.String()
			if w.Header().Get("X-Idempotency-Replayed") == "true" {
				t.Fatal("first request must not be marked replayed")
			}
			continue
		}
		if w.Body.String() != firstBody {
			t.Fatalf("attempt %d body = %s, want first response %s", i+1, w.Body.String(), firstBody)
		}
		if w.Header().Get("X-Idempotency-Replayed") != "true" {
			t.Fatalf("attempt %d missing replay header", i+1)
		}
	}
	if calls != 1 {
		t.Fatalf("handler executed %d times, want exactly 1", calls)
	}
}

func TestIdempotency_RequestIdKey(t *testing.T) {
	r := newIdempRouter(t)
	var calls int64
	r.POST("/orders", Idempotency(), func(c *gin.Context) {
		atomic.AddInt64(&calls, 1)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	h := map[string]string{HeaderRequestID: "rid-42"}
	do(r, http.MethodPost, "/orders", `{}`, h)
	do(r, http.MethodPost, "/orders", `{}`, h)
	if calls != 1 {
		t.Fatalf("same X-Request-Id replayed: calls = %d, want 1", calls)
	}
	// 不同 Request-Id 是不同请求
	do(r, http.MethodPost, "/orders", `{}`, map[string]string{HeaderRequestID: "rid-43"})
	if calls != 2 {
		t.Fatalf("different request id: calls = %d, want 2", calls)
	}
}

func TestIdempotency_DifferentCallersNotReplayed(t *testing.T) {
	r := newIdempRouter(t)
	var calls int64
	r.POST("/orders", Idempotency(), func(c *gin.Context) {
		atomic.AddInt64(&calls, 1)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// 无幂等键/请求ID → body hash 兜底；授权头不同 = 不同调用方
	do(r, http.MethodPost, "/orders", `{"item":"x"}`, map[string]string{"Authorization": "Bearer alice"})
	do(r, http.MethodPost, "/orders", `{"item":"x"}`, map[string]string{"Authorization": "Bearer bob"})
	if calls != 2 {
		t.Fatalf("different callers must not be treated as replay: calls = %d, want 2", calls)
	}
}

func TestIdempotency_InProgressReturns409(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		InitIdempotency(nil)
		rdb.Close()
	})
	InitIdempotency(rdb)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	block := make(chan struct{})
	r.POST("/orders", Idempotency(), func(c *gin.Context) {
		<-block // 模拟慢请求，首次仍在处理中
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	done := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		done <- do(r, http.MethodPost, "/orders", `{}`, map[string]string{"X-Idempotency-Key": "k1"})
	}()

	// 等首次请求进入 handler（pending 已写入）
	deadline := time.Now().Add(2 * time.Second)
	for {
		if n, _ := rdb.Keys(context.Background(), "idempotency:*").Result(); len(n) > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("pending key never appeared")
		}
		time.Sleep(10 * time.Millisecond)
	}

	w := do(r, http.MethodPost, "/orders", `{}`, map[string]string{"X-Idempotency-Key": "k1"})
	if w.Code != http.StatusConflict {
		t.Fatalf("in-flight duplicate status = %d, want 409", w.Code)
	}
	if !strings.Contains(w.Body.String(), "40901") {
		t.Fatalf("body = %s", w.Body.String())
	}

	close(block)
	if resp := <-done; resp.Code != http.StatusOK {
		t.Fatalf("first request status = %d, want 200", resp.Code)
	}
}
