package cache

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go-micro/pkg/config"
)

const NilValue = "__nil__"

func NewRedis() *redis.Client {
	addr := config.GetEnv("REDIS_ADDR", "127.0.0.1:6379")
	password := config.GetEnv("REDIS_PASSWORD", "")
	dbStr := config.GetEnv("REDIS_DB", "0")
	db, _ := strconv.Atoi(dbStr)
	return redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})
}

func GetJSON(ctx context.Context, rdb *redis.Client, key string, out interface{}) (hit bool, isNil bool, err error) {
	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}
	if val == NilValue {
		return true, true, nil
	}
	if err := json.Unmarshal([]byte(val), out); err != nil {
		return false, false, err
	}
	return true, false, nil
}

func SetJSON(ctx context.Context, rdb *redis.Client, key string, v interface{}, ttl time.Duration) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, key, string(b), ttl).Err()
}

func SetNil(ctx context.Context, rdb *redis.Client, key string, ttl time.Duration) error {
	return rdb.Set(ctx, key, NilValue, ttl).Err()
}
