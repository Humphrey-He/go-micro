package recommendation

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestEnrichItems(t *testing.T) {
	// Create mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	service := &Service{db: sqlxDB}

	ctx := context.Background()

	t.Run("enrich items with product info", func(t *testing.T) {
		items := []RecItem{
			{SkuID: 1001},
			{SkuID: 1002},
			{SkuID: 1003},
		}

		// Mock database query
		rows := sqlmock.NewRows([]string{"sku_id", "name", "price", "image"}).
			AddRow(1001, "iPhone 15 Pro", 799900, "https://example.com/iphone.jpg").
			AddRow(1002, "MacBook Pro", 1999900, "https://example.com/macbook.jpg").
			AddRow(1003, "AirPods Pro", 24900, "https://example.com/airpods.jpg")

		mock.ExpectQuery("SELECT sku_id, name, price, image FROM products WHERE sku_id IN").
			WithArgs(int64(1001), int64(1002), int64(1003)).
			WillReturnRows(rows)

		// Execute enrichment
		enriched := service.enrichItems(ctx, items)

		// Verify results
		assert.Len(t, enriched, 3)
		
		assert.Equal(t, int64(1001), enriched[0].SkuID)
		assert.Equal(t, "iPhone 15 Pro", enriched[0].Name)
		assert.Equal(t, 7999.00, enriched[0].Price)
		assert.Equal(t, "https://example.com/iphone.jpg", enriched[0].Image)

		assert.Equal(t, int64(1002), enriched[1].SkuID)
		assert.Equal(t, "MacBook Pro", enriched[1].Name)
		assert.Equal(t, 19999.00, enriched[1].Price)

		assert.Equal(t, int64(1003), enriched[2].SkuID)
		assert.Equal(t, "AirPods Pro", enriched[2].Name)
		assert.Equal(t, 249.00, enriched[2].Price)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("handle missing products with placeholder", func(t *testing.T) {
		items := []RecItem{
			{SkuID: 1001},
			{SkuID: 9999}, // Non-existent product
		}

		// Mock database query - only return one product
		rows := sqlmock.NewRows([]string{"sku_id", "name", "price", "image"}).
			AddRow(1001, "iPhone 15 Pro", 799900, "https://example.com/iphone.jpg")

		mock.ExpectQuery("SELECT sku_id, name, price, image FROM products WHERE sku_id IN").
			WithArgs(int64(1001), int64(9999)).
			WillReturnRows(rows)

		// Execute enrichment
		enriched := service.enrichItems(ctx, items)

		// Verify results
		assert.Len(t, enriched, 2)
		
		// First product should have real data
		assert.Equal(t, "iPhone 15 Pro", enriched[0].Name)
		assert.Equal(t, 7999.00, enriched[0].Price)

		// Second product should have placeholder
		assert.Equal(t, "商品9999", enriched[1].Name)
		assert.Equal(t, 0.0, enriched[1].Price)
		assert.Equal(t, "", enriched[1].Image)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("handle empty items list", func(t *testing.T) {
		items := []RecItem{}
		enriched := service.enrichItems(ctx, items)
		assert.Len(t, enriched, 0)
	})
}
