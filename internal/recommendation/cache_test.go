package recommendation

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, mr
}

func TestCache_SimilarProducts(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	cache := NewCache(client)
	ctx := context.Background()

	t.Run("cache miss returns false", func(t *testing.T) {
		items, ok := cache.GetSimilarProducts(ctx, 123, "cart")
		assert.False(t, ok)
		assert.Nil(t, items)
	})

	t.Run("set and get similar products", func(t *testing.T) {
		skuID := int64(100)
		scene := "cart"
		expectedItems := []RecItem{
			{SkuID: 101, Name: "Product 1", Price: 99.99, Similarity: 0.95},
			{SkuID: 102, Name: "Product 2", Price: 149.99, Similarity: 0.88},
		}

		err := cache.SetSimilarProducts(ctx, skuID, scene, expectedItems)
		require.NoError(t, err)

		items, ok := cache.GetSimilarProducts(ctx, skuID, scene)
		assert.True(t, ok)
		assert.Equal(t, expectedItems, items)
	})

	t.Run("different scenes have separate cache", func(t *testing.T) {
		skuID := int64(200)
		cartItems := []RecItem{{SkuID: 201, Name: "Cart Item"}}
		favoriteItems := []RecItem{{SkuID: 202, Name: "Favorite Item"}}

		cache.SetSimilarProducts(ctx, skuID, "cart", cartItems)
		cache.SetSimilarProducts(ctx, skuID, "favorite", favoriteItems)

		items1, ok1 := cache.GetSimilarProducts(ctx, skuID, "cart")
		items2, ok2 := cache.GetSimilarProducts(ctx, skuID, "favorite")

		assert.True(t, ok1)
		assert.True(t, ok2)
		assert.Equal(t, cartItems, items1)
		assert.Equal(t, favoriteItems, items2)
	})

	t.Run("TTL is set correctly", func(t *testing.T) {
		skuID := int64(300)
		scene := "purchase"
		items := []RecItem{{SkuID: 301}}

		cache.SetSimilarProducts(ctx, skuID, scene, items)

		key := "rec:similar:300:purchase"
		ttl := mr.TTL(key)
		assert.True(t, ttl > 0 && ttl <= 15*time.Minute)
	})
}

func TestCache_HomeRec(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	cache := NewCache(client)
	ctx := context.Background()

	t.Run("cache miss returns false", func(t *testing.T) {
		resp, ok := cache.GetHomeRec(ctx, 1, 1, 20)
		assert.False(t, ok)
		assert.Nil(t, resp)
	})

	t.Run("set and get home recommendations", func(t *testing.T) {
		userID := int64(1000)
		page := 1
		pageSize := 20

		expectedResp := &HomeRecResponse{
			Items: []RecItem{
				{SkuID: 1001, Name: "Rec 1", Price: 199.99},
				{SkuID: 1002, Name: "Rec 2", Price: 299.99},
			},
			Page:     page,
			PageSize: pageSize,
			Total:    2,
			Source:   "personalized",
		}

		err := cache.SetHomeRec(ctx, userID, page, pageSize, expectedResp)
		require.NoError(t, err)

		resp, ok := cache.GetHomeRec(ctx, userID, page, pageSize)
		assert.True(t, ok)
		assert.Equal(t, expectedResp, resp)
	})

	t.Run("different pages have separate cache", func(t *testing.T) {
		userID := int64(2000)
		page1Resp := &HomeRecResponse{Items: []RecItem{{SkuID: 2001}}, Page: 1, PageSize: 20, Total: 1, Source: "global"}
		page2Resp := &HomeRecResponse{Items: []RecItem{{SkuID: 2002}}, Page: 2, PageSize: 20, Total: 1, Source: "global"}

		cache.SetHomeRec(ctx, userID, 1, 20, page1Resp)
		cache.SetHomeRec(ctx, userID, 2, 20, page2Resp)

		resp1, ok1 := cache.GetHomeRec(ctx, userID, 1, 20)
		resp2, ok2 := cache.GetHomeRec(ctx, userID, 2, 20)

		assert.True(t, ok1)
		assert.True(t, ok2)
		assert.Equal(t, page1Resp, resp1)
		assert.Equal(t, page2Resp, resp2)
	})
}

func TestCache_InvalidateSimilarProducts(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	cache := NewCache(client)
	ctx := context.Background()

	t.Run("invalidate all scenes for a SKU", func(t *testing.T) {
		skuID := int64(400)
		cache.SetSimilarProducts(ctx, skuID, "cart", []RecItem{{SkuID: 401}})
		cache.SetSimilarProducts(ctx, skuID, "favorite", []RecItem{{SkuID: 402}})
		cache.SetSimilarProducts(ctx, skuID, "purchase", []RecItem{{SkuID: 403}})

		err := cache.InvalidateSimilarProducts(ctx, skuID)
		require.NoError(t, err)

		_, ok1 := cache.GetSimilarProducts(ctx, skuID, "cart")
		_, ok2 := cache.GetSimilarProducts(ctx, skuID, "favorite")
		_, ok3 := cache.GetSimilarProducts(ctx, skuID, "purchase")

		assert.False(t, ok1)
		assert.False(t, ok2)
		assert.False(t, ok3)
	})

	t.Run("invalidate does not affect other SKUs", func(t *testing.T) {
		cache.SetSimilarProducts(ctx, 500, "cart", []RecItem{{SkuID: 501}})
		cache.SetSimilarProducts(ctx, 600, "cart", []RecItem{{SkuID: 601}})

		cache.InvalidateSimilarProducts(ctx, 500)

		_, ok1 := cache.GetSimilarProducts(ctx, 500, "cart")
		items2, ok2 := cache.GetSimilarProducts(ctx, 600, "cart")

		assert.False(t, ok1)
		assert.True(t, ok2)
		assert.Equal(t, int64(601), items2[0].SkuID)
	})
}

func TestCache_GlobalBestsellers(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	cache := NewCache(client)
	ctx := context.Background()

	t.Run("set and get global bestsellers", func(t *testing.T) {
		limit := 10
		items := []RecItem{
			{SkuID: 701, Name: "Bestseller 1", Price: 99.99},
			{SkuID: 702, Name: "Bestseller 2", Price: 149.99},
		}

		err := cache.SetGlobalBestsellers(ctx, limit, items)
		require.NoError(t, err)

		result, ok := cache.GetGlobalBestsellers(ctx, limit)
		assert.True(t, ok)
		assert.Equal(t, items, result)
	})

	t.Run("different limits have separate cache", func(t *testing.T) {
		items10 := []RecItem{{SkuID: 801}}
		items20 := []RecItem{{SkuID: 802}}

		cache.SetGlobalBestsellers(ctx, 10, items10)
		cache.SetGlobalBestsellers(ctx, 20, items20)

		result10, ok10 := cache.GetGlobalBestsellers(ctx, 10)
		result20, ok20 := cache.GetGlobalBestsellers(ctx, 20)

		assert.True(t, ok10)
		assert.True(t, ok20)
		assert.Equal(t, items10, result10)
		assert.Equal(t, items20, result20)
	})
}

func TestCache_InvalidateAll(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	cache := NewCache(client)
	ctx := context.Background()

	cache.SetSimilarProducts(ctx, 900, "cart", []RecItem{{SkuID: 901}})
	cache.SetHomeRec(ctx, 1000, 1, 20, &HomeRecResponse{Items: []RecItem{{SkuID: 1001}}})
	cache.SetGlobalBestsellers(ctx, 10, []RecItem{{SkuID: 1002}})

	err := cache.InvalidateAll(ctx)
	require.NoError(t, err)

	_, ok1 := cache.GetSimilarProducts(ctx, 900, "cart")
	_, ok2 := cache.GetHomeRec(ctx, 1000, 1, 20)
	_, ok3 := cache.GetGlobalBestsellers(ctx, 10)

	assert.False(t, ok1)
	assert.False(t, ok2)
	assert.False(t, ok3)
}
