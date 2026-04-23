package job

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"go-micro/pkg/db"
)

const (
	weightPurchase = 10
	weightCart    = 3
	weightFavorite = 5
	minWeightThreshold = 0.05 // 5%
)

type PreferenceJob struct {
	db *sqlx.DB
}

func NewPreferenceJob() (*PreferenceJob, error) {
	dbx, err := db.NewMySQL()
	if err != nil {
		return nil, err
	}
	return &PreferenceJob{db: dbx}, nil
}

func (j *PreferenceJob) Run() error {
	log.Println("[PreferenceJob] starting...")
	start := time.Now()

	// Get all users with recent behavior
	users, err := j.getActiveUsers()
	if err != nil {
		log.Printf("[PreferenceJob] failed to get active users: %v", err)
		return err
	}

	log.Printf("[PreferenceJob] processing %d users", len(users))

	// Process each user
	for _, userID := range users {
		if err := j.computeUserPreference(userID); err != nil {
			log.Printf("[PreferenceJob] failed to compute preference for user %d: %v", userID, err)
		}
	}

	log.Printf("[PreferenceJob] completed in %v", time.Since(start))
	return nil
}

func (j *PreferenceJob) getActiveUsers() ([]int64, error) {
	rows, err := j.db.Query(`
		SELECT DISTINCT user_id
		FROM user_behavior_logs
		WHERE created_at > DATE_SUB(NOW(), INTERVAL 7 DAY)
			AND user_id > 0
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err == nil {
			users = append(users, userID)
		}
	}
	return users, nil
}

func (j *PreferenceJob) computeUserPreference(userID int64) error {
	// Get user behavior with category info
	rows, err := j.db.Query(`
		SELECT
			COALESCE(p.category_id, 0) as category_id,
			CASE b.behavior_type
				WHEN 'purchase' THEN weightPurchase
				WHEN 'cart' THEN weightCart
				WHEN 'favorite' THEN weightFavorite
			END as weight
		FROM user_behavior_logs b
		LEFT JOIN products p ON b.sku_id = p.sku_id
		WHERE b.user_id = ? AND b.created_at > DATE_SUB(NOW(), INTERVAL 30 DAY)
	`, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Aggregate weights by category
	categoryWeights := make(map[int64]float64)
	var totalWeight float64

	for rows.Next() {
		var catID int64
		var weight float64
		if err := rows.Scan(&catID, &weight); err == nil {
			categoryWeights[catID] += weight
			totalWeight += weight
		}
	}

	if totalWeight == 0 {
		return nil
	}

	// Delete old implicit preferences
	_, _ = j.db.Exec(`DELETE FROM user_category_preference WHERE user_id = ? AND source = 'implicit'`, userID)

	// Insert normalized preferences (only if weight > 5%)
	for catID, weight := range categoryWeights {
		normalized := weight / totalWeight
		if normalized > minWeightThreshold { // Only keep if > 5%
			_, err := j.db.Exec(`
				INSERT INTO user_category_preference (user_id, category_id, weight, source)
				VALUES (?, ?, ?, 'implicit')
				ON DUPLICATE KEY UPDATE weight = VALUES(weight)
			`, userID, catID, normalized)
			if err != nil {
				log.Printf("failed to insert preference for user %d, cat %d: %v", userID, catID, err)
			}
		}
	}

	return nil
}
