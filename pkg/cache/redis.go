package cache

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"go-micro/pkg/config"
	"go-micro/pkg/logx"
	"go.uber.org/zap"
)

const NilValue = "__nil__"

type localItem struct {
	val   string
	isNil bool
	exp   time.Time
}

var (
	localCache sync.Map
	cacheHits  = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "cache",
		Name:      "hits_total",
		Help:      "Cache hits total",
	}, []string{"cache"})
	cacheMisses = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "cache",
		Name:      "misses_total",
		Help:      "Cache misses total",
	}, []string{"cache"})
)

func init() {
	prometheus.MustRegister(cacheHits, cacheMisses)
}

func NewRedis() *redis.Client {
	addr := config.GetEnv("REDIS_ADDR", "127.0.0.1:6379")
	password := config.GetEnv("REDIS_PASSWORD", "")
	dbStr := config.GetEnv("REDIS_DB", "0")
	db, _ := strconv.Atoi(dbStr)
	return redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})
}

func GetJSON(ctx context.Context, rdb *redis.Client, key string, out interface{}) (hit bool, isNil bool, err error) {
	cacheName := nameFromKey(key)
	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		cacheMisses.WithLabelValues(cacheName).Inc()
		return false, false, nil
	}
	if err != nil {
		logx.L().Warn("redis get failed", zap.String("key", key), zap.Error(err))
		if v, ok, localNil := getLocal(key); ok {
			cacheHits.WithLabelValues(cacheName).Inc()
			if localNil {
				return true, true, nil
			}
			if err := json.Unmarshal([]byte(v), out); err != nil {
				return false, false, err
			}
			return true, false, nil
		}
		cacheMisses.WithLabelValues(cacheName).Inc()
		return false, false, err
	}
	if val == NilValue {
		cacheHits.WithLabelValues(cacheName).Inc()
		return true, true, nil
	}
	if err := json.Unmarshal([]byte(val), out); err != nil {
		return false, false, err
	}
	cacheHits.WithLabelValues(cacheName).Inc()
	return true, false, nil
}

func SetJSON(ctx context.Context, rdb *redis.Client, key string, v interface{}, ttl time.Duration) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	setLocal(key, string(b), ttl, false)
	if err := rdb.Set(ctx, key, string(b), ttl).Err(); err != nil {
		logx.L().Warn("redis set failed", zap.String("key", key), zap.Error(err))
		return err
	}
	return nil
}

func SetNil(ctx context.Context, rdb *redis.Client, key string, ttl time.Duration) error {
	setLocal(key, NilValue, ttl, true)
	if err := rdb.Set(ctx, key, NilValue, ttl).Err(); err != nil {
		logx.L().Warn("redis set nil failed", zap.String("key", key), zap.Error(err))
		return err
	}
	return nil
}

// TryLock acquires a short-lived lock for cache rebuild to avoid thundering herd.
func TryLock(ctx context.Context, rdb *redis.Client, key string, ttl time.Duration) (bool, error) {
	return rdb.SetNX(ctx, key, "1", ttl).Result()
}

func Unlock(ctx context.Context, rdb *redis.Client, key string) error {
	return rdb.Del(ctx, key).Err()
}

func IncHit(cacheName string) {
	cacheHits.WithLabelValues(cacheName).Inc()
}

func IncMiss(cacheName string) {
	cacheMisses.WithLabelValues(cacheName).Inc()
}

func SetLocalString(key, val string, ttl time.Duration, isNil bool) {
	setLocal(key, val, ttl, isNil)
}

func GetLocalString(key string) (string, bool, bool) {
	return getLocal(key)
}

func setLocal(key, val string, ttl time.Duration, isNil bool) {
	if ttl <= 0 {
		return
	}
	localCache.Store(key, localItem{val: val, isNil: isNil, exp: time.Now().Add(ttl)})
}

func getLocal(key string) (string, bool, bool) {
	if v, ok := localCache.Load(key); ok {
		item := v.(localItem)
		if time.Now().After(item.exp) {
			localCache.Delete(key)
			return "", false, false
		}
		return item.val, true, item.isNil
	}
	return "", false, false
}

func nameFromKey(key string) string {
	if key == "" {
		return "unknown"
	}
	parts := strings.SplitN(key, ":", 2)
	if len(parts) == 0 || parts[0] == "" {
		return "unknown"
	}
	return parts[0]
}
