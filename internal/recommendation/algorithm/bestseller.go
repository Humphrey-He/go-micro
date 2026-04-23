package algorithm

import (
	"log"

	"github.com/jmoiron/sqlx"
)

type Bestseller struct {
	db         *sqlx.DB
	periodDays int
	topN       int
}

func NewBestseller(db *sqlx.DB) *Bestseller {
	return &Bestseller{
		db:         db,
		periodDays: 30,
		topN:       100,
	}
}

// BehaviorWeights maps behavior types to their weights
var BehaviorWeights = map[string]float64{
	"purchase": 10.0,
	"cart":     3.0,
	"favorite": 5.0,
}

// ComputeCategoryBestsellers calculates category bestseller rankings
func (b *Bestseller) ComputeCategoryBestsellers() error {
	cutoff := "DATE_SUB(NOW(), INTERVAL 30 DAY)"

	// Query to get scores per SKU per category
	rows, err := b.db.Query(`
		SELECT
			COALESCE(p.category_id, 0) as category_id,
			b.sku_id,
			SUM(CASE b.behavior_type
				WHEN 'purchase' THEN 10
				WHEN 'cart' THEN 3
				WHEN 'favorite' THEN 5
			END) as score
		FROM user_behavior_logs b
		LEFT JOIN products p ON b.sku_id = p.sku_id
		WHERE b.created_at > ` + cutoff + `
			AND b.user_id > 0
		GROUP BY COALESCE(p.category_id, 0), b.sku_id
	`)
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
		if err := rows.Scan(&s.CategoryID, &s.SkuID, &s.Score); err == nil {
			categoryScores[s.CategoryID] = append(categoryScores[s.CategoryID], s)
		}
	}

	// Calculate rankings and save
	for catID, scores := range categoryScores {
		if err := b.saveCategoryBestsellers(catID, scores, "30d"); err != nil {
			log.Printf("failed to save category %d bestsellers: %v", catID, err)
		}
	}

	return nil
}

// ComputeGlobalBestsellers calculates global bestseller rankings
func (b *Bestseller) ComputeGlobalBestsellers() error {
	cutoff := "DATE_SUB(NOW(), INTERVAL 30 DAY)"

	rows, err := b.db.Query(`
		SELECT
			b.sku_id,
			SUM(CASE b.behavior_type
				WHEN 'purchase' THEN 10
				WHEN 'cart' THEN 3
				WHEN 'favorite' THEN 5
			END) as score
		FROM user_behavior_logs b
		WHERE b.created_at > ` + cutoff + `
			AND b.user_id > 0
		GROUP BY b.sku_id
		ORDER BY score DESC
		LIMIT ?
	`, b.topN)
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
		if err := rows.Scan(&s.SkuID, &s.Score); err == nil {
			scores = append(scores, s)
		}
	}

	return b.saveGlobalBestsellers(scores, "30d")
}

func (b *Bestseller) saveCategoryBestsellers(categoryID int64, scores []skuScore, period string) error {
	if len(scores) == 0 {
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
	for rank, s := range scores {
		_, err := tx.Exec(`
			INSERT INTO category_bestsellers (category_id, sku_id, sales_score, rank, period)
			VALUES (?, ?, ?, ?, ?)
		`, categoryID, s.SkuID, s.Score, rank+1, period)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (b *Bestseller) saveGlobalBestsellers(scores []skuScore, period string) error {
	if len(scores) == 0 {
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
	for rank, s := range scores {
		_, err := tx.Exec(`
			INSERT INTO global_bestsellers (sku_id, sales_score, rank, period)
			VALUES (?, ?, ?, ?)
		`, s.SkuID, s.Score, rank+1, period)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

type skuScore struct {
	CategoryID int64
	SkuID      int64
	Score      float64
}
