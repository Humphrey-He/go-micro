package algorithm

import (
	"fmt"
	"log"
	"sort"

	"github.com/jmoiron/sqlx"
)

const (
	// Minimum support threshold - a rule must appear in at least this fraction of transactions
	MinSupport = 0.01 // 1% of transactions

	// Minimum confidence threshold - P(B|A) must be at least this
	MinConfidence = 0.1 // 10%

	// Minimum lift threshold - rule must increase probability by at least this factor
	MinLift = 1.5

	// Maximum number of items in an itemset to consider
	MaxItemsetSize = 3
)

// RecItem - Recommended product item (local type for algorithm package)
type RecItem struct {
	SkuID      int64   `json:"sku_id"`
	Name       string  `json:"name"`
	Price      float64 `json:"price"`
	Image      string  `json:"image"`
	Similarity float64 `json:"similarity,omitempty"`
	Reason     string  `json:"reason,omitempty"`
}

type Association struct {
	db *sqlx.DB
}

func NewAssociation(db *sqlx.DB) *Association {
	return &Association{db: db}
}

// Transaction represents a shopping transaction (order) with multiple items
type Transaction []int64

// FrequentItemset represents a frequent itemset with its support
type FrequentItemset struct {
	Items   []int64  `json:"items"`
	Support float64  `json:"support"`
	Count   int      `json:"count"`
}

// AssociationRule represents an association rule A -> B
type AssociationRule struct {
	SKUIDA          int64   `json:"sku_id_a" db:"sku_id_a"`
	SKUIDB          int64   `json:"sku_id_b" db:"sku_id_b"`
	Support         float64 `json:"support" db:"support"`                   // P(A∪B)
	Confidence      float64 `json:"confidence" db:"confidence"`               // P(B|A)
	Lift            float64 `json:"lift" db:"lift"`                         // confidence/P(B)
	CoPurchaseCount int     `json:"co_purchase_count" db:"co_purchase_count"`
	APurchaseCount  int     `json:"a_purchase_count" db:"a_purchase_count"`
}

// ComputeAssociationRules mines association rules from purchase data
// Uses a simplified Apriori-like approach for finding frequent co-purchases
func (a *Association) ComputeAssociationRules() error {
	// Step 1: Get all transactions (orders with multiple items)
	transactions, totalTransactions, err := a.getTransactions()
	if err != nil {
		return fmt.Errorf("failed to get transactions: %w", err)
	}

	if len(transactions) < 10 {
		log.Printf("Not enough transactions to compute association rules: %d", len(transactions))
		return nil
	}

	log.Printf("Computing association rules from %d transactions (total: %d)", len(transactions), totalTransactions)

	// Step 2: Count co-occurrences (2-itemsets)
	coOccurrences, err := a.countCoOccurrences(transactions)
	if err != nil {
		return fmt.Errorf("failed to count co-occurrences: %w", err)
	}

	// Step 3: Generate rules from co-occurrences
	rules := a.generateRules(coOccurrences, totalTransactions)

	// Step 4: Save rules to database
	if err := a.saveRules(rules); err != nil {
		return fmt.Errorf("failed to save rules: %w", err)
	}

	log.Printf("Generated %d association rules", len(rules))
	return nil
}

// getTransactions retrieves all transactions (orders) with multiple items
func (a *Association) getTransactions() ([]Transaction, int, error) {
	// Query orders with multiple items
	query := `
		SELECT oi.order_id, oi.sku_id
		FROM order_items oi
		JOIN orders o ON oi.order_id = o.order_id
		WHERE o.status = 'SUCCESS'
			AND oi.sku_id > 0
		ORDER BY oi.order_id`

	rows, err := a.db.Query(query)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	type orderItem struct {
		OrderID string `db:"order_id"`
		SKUID   int64  `db:"sku_id"`
	}

	var items []orderItem
	for rows.Next() {
		var it orderItem
		if err := rows.Scan(&it.OrderID, &it.SKUID); err != nil {
			return nil, 0, err
		}
		items = append(items, it)
	}

	// Group by order_id to form transactions
	orderMap := make(map[string][]int64)
	for _, it := range items {
		orderMap[it.OrderID] = append(orderMap[it.OrderID], it.SKUID)
	}

	// Filter transactions with at least 2 items
	var transactions []Transaction
	for _, skus := range orderMap {
		if len(skus) >= 2 {
			transactions = append(transactions, skus)
		}
	}

	return transactions, len(items), nil
}

// countCoOccurrences counts how often each pair of items appears together
func (a *Association) countCoOccurrences(transactions []Transaction) (map[string]int, error) {
	coCount := make(map[string]int)

	for _, trans := range transactions {
		// For each pair in the transaction
		for i := 0; i < len(trans); i++ {
			for j := i + 1; j < len(trans); j++ {
				// Ensure consistent ordering (smaller SKU first)
				skuA, skuB := trans[i], trans[j]
				if skuA > skuB {
					skuA, skuB = skuB, skuA
				}
				key := fmt.Sprintf("%d-%d", skuA, skuB)
				coCount[key]++
			}
		}
	}

	return coCount, nil
}

// generateRules creates association rules from co-occurrence counts
func (a *Association) generateRules(coOccurrences map[string]int, totalTransactions int) []AssociationRule {
	var rules []AssociationRule

	// Count individual item purchases from the database
	itemPurchaseCount := a.countItemPurchases(coOccurrences)

	minCount := int(float64(totalTransactions) * MinSupport)

	for key, coCount := range coOccurrences {
		if coCount < minCount {
			continue // Below minimum support
		}

		var skuA, skuB int64
		fmt.Sscanf(key, "%d-%d", &skuA, &skuB)

		countA := itemPurchaseCount[skuA]
		countB := itemPurchaseCount[skuB]

		if countA == 0 || countB == 0 {
			continue
		}

		support := float64(coCount) / float64(totalTransactions)
		confidence := float64(coCount) / float64(countA)
		lift := confidence / (float64(countB) / float64(totalTransactions))

		if confidence < MinConfidence || lift < MinLift {
			continue
		}

		// Only keep rules where A -> B (A is the premise)
		rules = append(rules, AssociationRule{
			SKUIDA:          skuA,
			SKUIDB:          skuB,
			Support:         support,
			Confidence:      confidence,
			Lift:            lift,
			CoPurchaseCount: coCount,
			APurchaseCount:  countA,
		})
	}

	// Sort by lift descending
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Lift > rules[j].Lift
	})

	// Limit to top rules per item
	return rules
}

// countItemPurchases counts how many times each item was purchased
func (a *Association) countItemPurchases(coOccurrences map[string]int) map[int64]int {
	itemCount := make(map[int64]int)

	// Query directly from order_items
	query := `
		SELECT oi.sku_id, COUNT(DISTINCT oi.order_id) as cnt
		FROM order_items oi
		JOIN orders o ON oi.order_id = o.order_id
		WHERE o.status = 'SUCCESS'
			AND oi.sku_id > 0
		GROUP BY oi.sku_id`

	rows, err := a.db.Query(query)
	if err != nil {
		log.Printf("failed to count item purchases: %v", err)
		return itemCount
	}
	defer rows.Close()

	for rows.Next() {
		var skuID int64
		var count int
		if err := rows.Scan(&skuID, &count); err != nil {
			continue
		}
		itemCount[skuID] = count
	}

	return itemCount
}

// saveRules persists association rules to the database
func (a *Association) saveRules(rules []AssociationRule) error {
	if len(rules) == 0 {
		return nil
	}

	tx, err := a.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear old rules
	_, _ = tx.Exec(`DELETE FROM association_rules`)

	// Insert new rules
	for _, rule := range rules {
		_, err := tx.Exec(`
			INSERT INTO association_rules (sku_id_a, sku_id_b, support, confidence, lift, co_purchase_count, a_purchase_count)
			VALUES (?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				support = VALUES(support),
				confidence = VALUES(confidence),
				lift = VALUES(lift),
				co_purchase_count = VALUES(co_purchase_count),
				a_purchase_count = VALUES(a_purchase_count)
		`, rule.SKUIDA, rule.SKUIDB, rule.Support, rule.Confidence, rule.Lift, rule.CoPurchaseCount, rule.APurchaseCount)
		if err != nil {
			return fmt.Errorf("failed to insert rule %d->%d: %w", rule.SKUIDA, rule.SKUIDB, err)
		}
	}

	return tx.Commit()
}

// GetCartAddons finds products to recommend as cart add-ons based on association rules
func (a *Association) GetCartAddons(cartSKUIDs []int64, limit int) ([]RecItem, error) {
	if len(cartSKUIDs) == 0 {
		return nil, nil
	}

	if limit <= 0 {
		limit = 10
	}

	// Find all association rules where any cart item is the premise (A)
	// and rank by lift and confidence
	query := `
		SELECT ar.sku_id_b, ar.sku_id_a, ar.confidence, ar.lift, ar.support,
			   p.name, p.price, p.original_price, p.image
		FROM association_rules ar
		LEFT JOIN products p ON ar.sku_id_b = p.sku_id
		WHERE ar.sku_id_a IN (?) AND ar.sku_id_b NOT IN (?)
		ORDER BY ar.lift DESC, ar.confidence DESC
		LIMIT ?`

	// Use sqlx.In for proper query building
	query, args, err := sqlx.In(query, cartSKUIDs, cartSKUIDs)
	if err != nil {
		return nil, err
	}
	query = a.db.Rebind(query)
	args = append(args, limit)

	rows, err := a.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type ruleResult struct {
		SKUIDB        int64   `db:"sku_id_b"`
		SKUIDA        int64   `db:"sku_id_a"`
		Confidence    float64 `db:"confidence"`
		Lift          float64 `db:"lift"`
		Support       float64 `db:"support"`
		Name          string  `db:"name"`
		Price         int64   `db:"price"`
		OriginalPrice int64   `db:"original_price"`
		Image         string  `db:"image"`
	}

	// Deduplicate by SKU_ID_B (same product from different rules)
	seen := make(map[int64]bool)
	var items []RecItem

	for rows.Next() {
		var r ruleResult
		if err := rows.Scan(&r.SKUIDB, &r.SKUIDA, &r.Confidence, &r.Lift, &r.Support,
			&r.Name, &r.Price, &r.OriginalPrice, &r.Image); err != nil {
			continue
		}

		if seen[r.SKUIDB] {
			continue
		}
		seen[r.SKUIDB] = true

		name := r.Name
		if name == "" {
			name = fmt.Sprintf("Product %d", r.SKUIDB)
		}

		items = append(items, RecItem{
			SkuID:      r.SKUIDB,
			Name:       name,
			Price:      float64(r.Price),
			Image:      r.Image,
			Similarity: r.Lift * 100, // Use lift as score
			Reason:     "常与已购商品一起购买",
		})
	}

	return items, nil
}

// GetPayCompleteRecommendations finds products to recommend after purchase completion
func (a *Association) GetPayCompleteRecommendations(purchasedSKUIDs []int64, limit int) ([]RecItem, error) {
	if len(purchasedSKUIDs) == 0 {
		return nil, nil
	}

	if limit <= 0 {
		limit = 10
	}

	// Find products frequently purchased together with purchased items
	// but prefer higher confidence rules
	query := `
		SELECT DISTINCT ar.sku_id_b, ar.confidence, ar.lift,
			   p.name, p.price, p.original_price, p.image
		FROM association_rules ar
		LEFT JOIN products p ON ar.sku_id_b = p.sku_id
		WHERE ar.sku_id_a IN (?)
			AND ar.sku_id_b NOT IN (?)
			AND ar.confidence > 0.2
		ORDER BY ar.confidence DESC, ar.lift DESC
		LIMIT ?`

	query, args, err := sqlx.In(query, purchasedSKUIDs, purchasedSKUIDs)
	if err != nil {
		return nil, err
	}
	query = a.db.Rebind(query)
	args = append(args, limit)

	rows, err := a.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seen := make(map[int64]bool)
	var items []RecItem

	for rows.Next() {
		var r struct {
			SKUIDB        int64   `db:"sku_id_b"`
			Confidence    float64 `db:"confidence"`
			Lift          float64 `db:"lift"`
			Name          string  `db:"name"`
			Price         int64   `db:"price"`
			OriginalPrice int64   `db:"original_price"`
			Image         string  `db:"image"`
		}
		if err := rows.Scan(&r.SKUIDB, &r.Confidence, &r.Lift,
			&r.Name, &r.Price, &r.OriginalPrice, &r.Image); err != nil {
			continue
		}

		if seen[r.SKUIDB] {
			continue
		}
		seen[r.SKUIDB] = true

		name := r.Name
		if name == "" {
			name = fmt.Sprintf("Product %d", r.SKUIDB)
		}

		items = append(items, RecItem{
			SkuID:      r.SKUIDB,
			Name:       name,
			Price:      float64(r.Price),
			Image:      r.Image,
			Similarity: r.Confidence * 100,
			Reason:     "为你推荐",
		})
	}

	return items, nil
}

// GetRelatedProducts finds products related to a given product via association rules
func (a *Association) GetRelatedProducts(skuID int64, limit int) ([]RecItem, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT ar.sku_id_b, ar.confidence, ar.lift,
			   p.name, p.price, p.original_price, p.image
		FROM association_rules ar
		LEFT JOIN products p ON ar.sku_id_b = p.sku_id
		WHERE ar.sku_id_a = ?
		ORDER BY ar.lift DESC, ar.confidence DESC
		LIMIT ?`

	rows, err := a.db.Query(query, skuID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []RecItem

	for rows.Next() {
		var r struct {
			SKUIDB        int64   `db:"sku_id_b"`
			Confidence    float64 `db:"confidence"`
			Lift          float64 `db:"lift"`
			Name          string  `db:"name"`
			Price         int64   `db:"price"`
			OriginalPrice int64   `db:"original_price"`
			Image         string  `db:"image"`
		}
		if err := rows.Scan(&r.SKUIDB, &r.Confidence, &r.Lift,
			&r.Name, &r.Price, &r.OriginalPrice, &r.Image); err != nil {
			continue
		}

		name := r.Name
		if name == "" {
			name = fmt.Sprintf("Product %d", r.SKUIDB)
		}

		items = append(items, RecItem{
			SkuID:      r.SKUIDB,
			Name:       name,
			Price:      float64(r.Price),
			Image:      r.Image,
			Similarity: r.Lift * 100,
			Reason:     "看了又看",
		})
	}

	return items, nil
}
