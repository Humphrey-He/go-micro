package algorithm

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

const (
	WeightPurchase    = 10
	WeightCart        = 3
	WeightFavorite    = 5
	DefaultTopN       = 100
	DefaultPeriod     = "30d"
	DefaultPeriodDays = 30 // Number of days for the behavior analysis period in SQL interval
)

type Bestseller struct {
	db   *sqlx.DB
	topN int
}

func NewBestseller(db *sqlx.DB) *Bestseller {
	return &Bestseller{
		db:   db,
		topN: DefaultTopN,
	}
}

// ComputeCategoryBestsellers calculates category bestseller rankings
func (b *Bestseller) ComputeCategoryBestsellers() error {
	query := fmt.Sprintf(`
		SELECT
			COALESCE(p.category_id, 0) as category_id,
			b.sku_id,
			SUM(CASE b.behavior_type
				WHEN 'purchase' THEN ?
				WHEN 'cart' THEN ?
				WHEN 'favorite' THEN ?
			END) as score
		FROM user_behavior_logs b
		LEFT JOIN products p ON b.sku_id = p.sku_id
		WHERE b.created_at > DATE_SUB(NOW(), INTERVAL %d DAY)
			AND b.user_id > 0
		GROUP BY COALESCE(p.category_id, 0), b.sku_id
	`, DefaultPeriodDays)
	rows, err := b.db.Query(query, WeightPurchase, WeightCart, WeightFavorite)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Group by category
	type skuScore struct {
		CategoryID int64
		SkuID      int64
		Score      float64
	}
	categoryScores := make(map[int64][]skuScore)

	for rows.Next() {
		var s skuScore
		if err := rows.Scan(&s.CategoryID, &s.SkuID, &s.Score); err != nil {
			log.Printf("failed to scan row: %v", err)
			continue
		}
		categoryScores[s.CategoryID] = append(categoryScores[s.CategoryID], s)
	}

	// Calculate rankings and save
	for catID, scores := range categoryScores {
		skuIDs := make([]int64, len(scores))
		scoreVals := make([]float64, len(scores))
		for i, s := range scores {
			skuIDs[i] = s.SkuID
			scoreVals[i] = s.Score
		}
		if err := b.saveCategoryBestsellers(catID, skuIDs, scoreVals, DefaultPeriod); err != nil {
			log.Printf("failed to save category %d bestsellers: %v", catID, err)
		}
	}

	return nil
}

// ComputeGlobalBestsellers calculates global bestseller rankings
func (b *Bestseller) ComputeGlobalBestsellers() error {
	query := fmt.Sprintf(`
		SELECT
			b.sku_id,
			SUM(CASE b.behavior_type
				WHEN 'purchase' THEN ?
				WHEN 'cart' THEN ?
				WHEN 'favorite' THEN ?
			END) as score
		FROM user_behavior_logs b
		WHERE b.created_at > DATE_SUB(NOW(), INTERVAL %d DAY)
			AND b.user_id > 0
		GROUP BY b.sku_id
		ORDER BY score DESC
		LIMIT ?
	`, DefaultPeriodDays)
	rows, err := b.db.Query(query, WeightPurchase, WeightCart, WeightFavorite, b.topN)
	if err != nil {
		return err
	}
	defer rows.Close()

	type skuScore struct {
		SkuID int64
		Score float64
	}
	var scores []skuScore

	for rows.Next() {
		var s skuScore
		if err := rows.Scan(&s.SkuID, &s.Score); err != nil {
			log.Printf("failed to scan row: %v", err)
			continue
		}
		scores = append(scores, s)
	}

	skuIDs := make([]int64, len(scores))
	scoreVals := make([]float64, len(scores))
	for i, s := range scores {
		skuIDs[i] = s.SkuID
		scoreVals[i] = s.Score
	}
	return b.saveGlobalBestsellers(skuIDs, scoreVals, DefaultPeriod)
}

func (b *Bestseller) saveCategoryBestsellers(categoryID int64, skuIDs []int64, scores []float64, period string) error {
	if len(skuIDs) == 0 {
		return nil
	}

	tx, err := b.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete old data
	_, _ = tx.Exec(`DELETE FROM category_bestsellers WHERE category_id = ? AND period = ?`, categoryID, period)

	// Insert new data
	for i, skuID := range skuIDs {
		_, err := tx.Exec(`
			INSERT INTO category_bestsellers (category_id, sku_id, sales_score, rank, period)
			VALUES (?, ?, ?, ?, ?)
		`, categoryID, skuID, scores[i], i+1, period)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (b *Bestseller) saveGlobalBestsellers(skuIDs []int64, scores []float64, period string) error {
	if len(skuIDs) == 0 {
		return nil
	}

	tx, err := b.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete old data
	_, _ = tx.Exec(`DELETE FROM global_bestsellers WHERE period = ?`, period)

	// Insert new data
	for i, skuID := range skuIDs {
		_, err := tx.Exec(`
			INSERT INTO global_bestsellers (sku_id, sales_score, rank, period)
			VALUES (?, ?, ?, ?)
		`, skuID, scores[i], i+1, period)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
