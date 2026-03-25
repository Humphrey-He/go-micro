package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type sample struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestGetSetJSON(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()

	key := "k1"
	obj := sample{ID: "1", Name: "n1"}
	if err := SetJSON(ctx, rdb, key, obj, time.Minute); err != nil {
		t.Fatal(err)
	}

	var out sample
	hit, isNil, err := GetJSON(ctx, rdb, key, &out)
	if err != nil || !hit || isNil {
		t.Fatalf("unexpected get result: hit=%v nil=%v err=%v", hit, isNil, err)
	}
	if out.ID != "1" {
		t.Fatalf("unexpected value: %+v", out)
	}
}

func TestSetNil(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()

	key := "k2"
	if err := SetNil(ctx, rdb, key, time.Minute); err != nil {
		t.Fatal(err)
	}

	var out sample
	hit, isNil, err := GetJSON(ctx, rdb, key, &out)
	if err != nil || !hit || !isNil {
		t.Fatalf("unexpected get result: hit=%v nil=%v err=%v", hit, isNil, err)
	}
}
