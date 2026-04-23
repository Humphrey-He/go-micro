package recommendation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	CacheTTL             = 15 * time.Minute
	SimilarProductsCache = "rec:similar:%d:%s"  // sku_id:scene
	HomeRecCache         = "rec:home:%d:%d:%d"   // user_id:page:pageSize
	GlobalBestsellerCache = "rec:bestseller:global:%d"  // limit
)

type Cache struct {
	rdb *redis.Client
}

func NewCache(rdb *redis.Client) *Cache {
	return &Cache{rdb: rdb}
}

// GetSimilarProducts gets similar products from cache
func (c *Cache) GetSimilarProducts(ctx context.Context, skuID int64, scene string) ([]RecItem, bool) {
	key := fmt.Sprintf(SimilarProductsCache, skuID, scene)
	data, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, false
	}

	var items []RecItem
	if err := json.Unmarshal([]byte(data), &items); err != nil {
		return nil, false
	}
	return items, true
}

// SetSimilarProducts sets similar products in cache
func (c *Cache) SetSimilarProducts(ctx context.Context, skuID int64, scene string, items []RecItem) error {
	key := fmt.Sprintf(SimilarProductsCache, skuID, scene)
	data, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, data, CacheTTL).Err()
}

// GetHomeRec gets home recommendations from cache
func (c *Cache) GetHomeRec(ctx context.Context, userID int64, page, pageSize int) (*HomeRecResponse, bool) {
	key := fmt.Sprintf(HomeRecCache, userID, page, pageSize)
	data, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, false
	}

	var resp HomeRecResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		return nil, false
	}
	return &resp, true
}

// SetHomeRec sets home recommendations in cache
func (c *Cache) SetHomeRec(ctx context.Context, userID int64, page, pageSize int, resp *HomeRecResponse) error {
	key := fmt.Sprintf(HomeRecCache, userID, page, pageSize)
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, data, CacheTTL).Err()
}

// InvalidateSimilarProducts invalidates similar products cache for a SKU
func (c *Cache) InvalidateSimilarProducts(ctx context.Context, skuID int64) error {
	pattern := fmt.Sprintf("rec:similar:%d:*", skuID)
	keys, err := c.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return c.rdb.Del(ctx, keys...).Err()
	}
	return nil
}

// GetGlobalBestsellers gets global bestsellers from cache
func (c *Cache) GetGlobalBestsellers(ctx context.Context, limit int) ([]RecItem, bool) {
	key := fmt.Sprintf(GlobalBestsellerCache, limit)
	data, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, false
	}

	var items []RecItem
	if err := json.Unmarshal([]byte(data), &items); err != nil {
		return nil, false
	}
	return items, true
}

// SetGlobalBestsellers sets global bestsellers in cache
func (c *Cache) SetGlobalBestsellers(ctx context.Context, limit int, items []RecItem) error {
	key := fmt.Sprintf(GlobalBestsellerCache, limit)
	data, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, data, CacheTTL).Err()
}

// InvalidateAll invalidates all recommendation caches
func (c *Cache) InvalidateAll(ctx context.Context) error {
	patterns := []string{
		"rec:similar:*",
		"rec:home:*",
		"rec:bestseller:*",
	}

	for _, pattern := range patterns {
		keys, err := c.rdb.Keys(ctx, pattern).Result()
		if err != nil {
			continue
		}
		if len(keys) > 0 {
			c.rdb.Del(ctx, keys...)
		}
	}
	return nil
}