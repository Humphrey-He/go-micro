package algorithm

import (
	"math"
	"sort"

	"github.com/jmoiron/sqlx"
)

const (
	DefaultMinCoUsers    = 2
	SimilarityThreshold  = 0.01
	BatchSize            = 1000
	DefaultLimit         = 10
)

type ItemCF struct {
	db         *sqlx.DB
	minCoUsers int
}

func NewItemCF(db *sqlx.DB) *ItemCF {
	return &ItemCF{
		db:         db,
		minCoUsers: DefaultMinCoUsers,
	}
}

// ComputeSimilarity computes item-item similarity based on co-occurrence
func (ic *ItemCF) ComputeSimilarity() error {
	// Step 1: Count users per item
	itemUserCounts, err := ic.countItemUsers()
	if err != nil {
		return err
	}

	// Step 2: Find co-occurring item pairs
	pairs, err := ic.countCoOccurrences()
	if err != nil {
		return err
	}

	// Step 3: Calculate similarity and save
	var batch []struct {
		SkuIDA     int64
		SkuIDB     int64
		Scene      string
		Similarity float64
		Weight     int
	}
	for _, pair := range pairs {
		countA := itemUserCounts[pair.SkuID]
		countB := itemUserCounts[pair.SkuIDB]
		if countA == 0 || countB == 0 {
			continue
		}

		// Cosine similarity
		sim := float64(pair.CoUsers) / math.Sqrt(float64(countA)*float64(countB))

		if sim > SimilarityThreshold && pair.CoUsers >= ic.minCoUsers {
			batch = append(batch, struct {
				SkuIDA     int64
				SkuIDB     int64
				Scene      string
				Similarity float64
				Weight     int
			}{
				SkuIDA:     pair.SkuID,
				SkuIDB:     pair.SkuIDB,
				Scene:      pair.Scene,
				Similarity: sim,
				Weight:     pair.CoUsers,
			})
		}

		// Batch insert when reaching BatchSize
		if len(batch) >= BatchSize {
			if err := ic.saveBatch(batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	// Save remaining
	if len(batch) > 0 {
		if err := ic.saveBatch(batch); err != nil {
			return err
		}
	}

	return nil
}

type itemUserCount struct {
	SkuID int64
	Count int
}

type coOccurrence struct {
	SkuID   int64
	SkuIDB  int64
	Scene   string
	CoUsers int
}

func (ic *ItemCF) countItemUsers() (map[int64]int, error) {
	counts := make(map[int64]int)
	rows, err := ic.db.Query(`
		SELECT sku_id, COUNT(DISTINCT user_id) as cnt
		FROM user_behavior_logs
		WHERE created_at > DATE_SUB(NOW(), INTERVAL 30 DAY)
			AND user_id > 0
		GROUP BY sku_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c itemUserCount
		if err := rows.Scan(&c.SkuID, &c.Count); err == nil {
			counts[c.SkuID] = c.Count
		}
	}
	return counts, nil
}

func (ic *ItemCF) countCoOccurrences() ([]coOccurrence, error) {
	pairs := make([]coOccurrence, 0)
	rows, err := ic.db.Query(`
		SELECT
			a.sku_id as sku_id_a,
			b.sku_id as sku_id_b,
			a.behavior_type as scene,
			COUNT(DISTINCT a.user_id) as co_users
		FROM user_behavior_logs a
		JOIN user_behavior_logs b ON a.user_id = b.user_id
			AND a.sku_id != b.sku_id
			AND a.behavior_type = b.behavior_type
		WHERE a.created_at > DATE_SUB(NOW(), INTERVAL 30 DAY)
			AND b.created_at > DATE_SUB(NOW(), INTERVAL 30 DAY)
			AND a.user_id > 0
			AND b.user_id > 0
		GROUP BY a.sku_id, b.sku_id, a.behavior_type
		HAVING co_users >= ?
	`, ic.minCoUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p coOccurrence
		if err := rows.Scan(&p.SkuID, &p.SkuIDB, &p.Scene, &p.CoUsers); err == nil {
			pairs = append(pairs, p)
		}
	}
	return pairs, nil
}

func (ic *ItemCF) saveBatch(batch []struct {
	SkuIDA     int64
	SkuIDB     int64
	Scene      string
	Similarity float64
	Weight     int
}) error {
	if len(batch) == 0 {
		return nil
	}

	tx, err := ic.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete old similarity for these pairs
	for _, b := range batch {
		_, _ = tx.Exec(`
			DELETE FROM product_similarity
			WHERE (sku_id_a = ? AND sku_id_b = ?) OR (sku_id_a = ? AND sku_id_b = ?)
		`, b.SkuIDA, b.SkuIDB, b.SkuIDB, b.SkuIDA)
	}

	// Insert new similarity
	for _, b := range batch {
		_, err := tx.Exec(`
			INSERT INTO product_similarity (sku_id_a, sku_id_b, scene, similarity, weight)
			VALUES (?, ?, ?, ?, ?)
		`, b.SkuIDA, b.SkuIDB, b.Scene, b.Similarity, b.Weight)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetSimilarItems returns similar items for a given SKU
func (ic *ItemCF) GetSimilarItems(skuID int64, scene string, limit int) ([]SimilarItem, error) {
	if limit <= 0 {
		limit = DefaultLimit
	}

	type result struct {
		SkuIDB     int64   `db:"sku_id_b"`
		Similarity float64 `db:"similarity"`
		Weight     int     `db:"weight"`
	}

	var results []result
	err := ic.db.Select(&results, `
		SELECT sku_id_b, similarity, weight
		FROM (
			SELECT sku_id_b, similarity, weight
			FROM product_similarity
			WHERE sku_id_a = ? AND scene = ?
			UNION ALL
			SELECT sku_id_a as sku_id_b, similarity, weight
			FROM product_similarity
			WHERE sku_id_b = ? AND scene = ?
		) combined
		ORDER BY similarity DESC
		LIMIT ?
	`, skuID, scene, skuID, scene, limit)
	if err != nil {
		return nil, err
	}

	items := make([]SimilarItem, 0, len(results))
	for _, r := range results {
		items = append(items, SimilarItem{
			SkuID:      r.SkuIDB,
			Similarity: r.Similarity,
		})
	}
	return items, nil
}

type SimilarItem struct {
	SkuID      int64
	Similarity float64
}

// SortSimilarItems sorts similar items by similarity descending
func SortSimilarItems(items []SimilarItem) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].Similarity > items[j].Similarity
	})
}