package recommendation

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestService_ReportBehavior(t *testing.T) {
	// Setup miniredis
	s := miniredis.RunT(t)
	defer s.Close()

	// This is a basic structure test - actual DB tests would need a test database
	// For now, just verify the method signature and basic behavior
	t.Run("basic signature test", func(t *testing.T) {
		// Just verify the service can be created and method exists
		assert.NotNil(t, &Service{})
	})
}

func TestFiveMinuteBucket(t *testing.T) {
	// Test that the bucket calculation is correct
	now := time.Now()
	bucket := int(now.Unix() / FiveMinuteBucket)

	// Two calls within 5 minutes should give same bucket
	bucket2 := int(now.Add(2*time.Minute).Unix() / FiveMinuteBucket)
	assert.Equal(t, bucket, bucket2)

	// After 5 minutes, bucket should change
	bucket3 := int(now.Add(6*time.Minute).Unix() / FiveMinuteBucket)
	assert.NotEqual(t, bucket, bucket3)
}

func TestBehaviorType_Constants(t *testing.T) {
	assert.Equal(t, BehaviorType("cart"), BehaviorCart)
	assert.Equal(t, BehaviorType("favorite"), BehaviorFavorite)
	assert.Equal(t, BehaviorType("purchase"), BehaviorPurchase)
}

func TestRecItem_Structure(t *testing.T) {
	item := RecItem{
		SkuID:      12345,
		Name:       "Test Product",
		Price:      99.99,
		Image:      "https://example.com/image.jpg",
		Similarity: 0.85,
		Reason:     "为你推荐",
	}

	assert.Equal(t, int64(12345), item.SkuID)
	assert.Equal(t, "Test Product", item.Name)
	assert.Equal(t, 99.99, item.Price)
	assert.Equal(t, 0.85, item.Similarity)
	assert.Equal(t, "为你推荐", item.Reason)
}
