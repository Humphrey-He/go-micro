package algorithm

import (
	"math"

	"github.com/jmoiron/sqlx"
)

type UserCF struct {
	db            *sqlx.DB
	topK          int
	minBehaviors  int
}

func NewUserCF(db *sqlx.DB) *UserCF {
	return &UserCF{
		db:           db,
		topK:         20,
		minBehaviors: 10,
	}
}

const (
	UserCFWeightPurchase = 10.0
	UserCFWeightCart     = 3.0
	UserCFWeightFavorite = 5.0
)

type userBehavior struct {
	UserID int64
	SkuID  int64
	Type   string
	Weight float64
}

type itemScore struct {
	SkuID int64
	Score float64
}

// GetUserRecommendations returns SKU IDs recommended for a user
func (uc *UserCF) GetUserRecommendations(userID int64, limit int) ([]int64, error) {
	if limit <= 0 {
		limit = 20
	}

	// Get user's behaviors
	userItems := uc.getUserBehaviors(userID)
	if len(userItems) < uc.minBehaviors {
		return nil, nil // Not enough data
	}

	// Find similar users
	similarUsers := uc.findSimilarUsers(userItems)
	if len(similarUsers) == 0 {
		return nil, nil
	}

	// Predict scores for items
	scores := uc.predictScores(userItems, similarUsers)

	// Sort by score descending
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].Score > scores[i].Score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	// Return top N
	result := make([]int64, 0, limit)
	for i := 0; i < len(scores) && i < limit; i++ {
		result = append(result, scores[i].SkuID)
	}
	return result, nil
}

func (uc *UserCF) getUserBehaviors(userID int64) []userBehavior {
	rows, err := uc.db.Query(`
		SELECT user_id, sku_id, behavior_type
		FROM user_behavior_logs
		WHERE user_id = ? AND created_at > DATE_SUB(NOW(), INTERVAL 30 DAY)
	`, userID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var behaviors []userBehavior
	for rows.Next() {
		var b userBehavior
		var typ string
		if err := rows.Scan(&b.UserID, &b.SkuID, &typ); err == nil {
			b.Weight = uc.getWeight(typ)
			behaviors = append(behaviors, b)
		}
	}
	return behaviors
}

func (uc *UserCF) getWeight(typ string) float64 {
	switch typ {
	case "purchase":
		return UserCFWeightPurchase
	case "cart":
		return UserCFWeightCart
	case "favorite":
		return UserCFWeightFavorite
	default:
		return 1.0
	}
}

func (uc *UserCF) findSimilarUsers(userItems []userBehavior) map[int64]float64 {
	if len(userItems) == 0 {
		return nil
	}

	// Build SKU set for user
	skuSet := make(map[int64]bool)
	for _, item := range userItems {
		skuSet[item.SkuID] = true
	}

	// Find users who interacted with same items
	similarUsers := make(map[int64]float64)

	skuIDs := make([]int64, 0, len(userItems))
	for _, item := range userItems {
		skuIDs = append(skuIDs, item.SkuID)
	}

	rows, err := uc.db.Query(`
		SELECT DISTINCT user_id
		FROM user_behavior_logs
		WHERE sku_id IN (SELECT sku_id FROM user_behavior_logs WHERE user_id = (SELECT MIN(user_id) FROM user_behavior_logs WHERE sku_id IN (SELECT sku_id FROM user_behavior_logs WHERE user_id = ? LIMIT 1)))
			AND user_id != ?
	`, userItems[0].UserID, userItems[0].UserID)
	if err != nil {
		return similarUsers
	}
	defer rows.Close()

	for rows.Next() {
		var otherUserID int64
		if err := rows.Scan(&otherUserID); err == nil {
			// Calculate similarity
			otherBehaviors := uc.getUserBehaviors(otherUserID)
			if len(otherBehaviors) == 0 {
				continue
			}

			common := 0
			otherSkuSet := make(map[int64]bool)
			for _, b := range otherBehaviors {
				otherSkuSet[b.SkuID] = true
			}
			for skuID := range skuSet {
				if otherSkuSet[skuID] {
					common++
				}
			}

			if common > 0 {
				sim := float64(common) / math.Sqrt(float64(len(userItems)*len(otherBehaviors)))
				if sim > 0.01 {
					similarUsers[otherUserID] = sim
				}
			}
		}
	}

	return similarUsers
}

func (uc *UserCF) predictScores(userItems []userBehavior, similarUsers map[int64]float64) []itemScore {
	userSkuSet := make(map[int64]bool)
	for _, b := range userItems {
		userSkuSet[b.SkuID] = true
	}

	itemScores := make(map[int64]float64)
	for otherUserID, sim := range similarUsers {
		otherBehaviors := uc.getUserBehaviors(otherUserID)
		for _, b := range otherBehaviors {
			if !userSkuSet[b.SkuID] { // Don't recommend items user already has
				itemScores[b.SkuID] += sim * b.Weight
			}
		}
	}

	scores := make([]itemScore, 0, len(itemScores))
	for skuID, score := range itemScores {
		scores = append(scores, itemScore{SkuID: skuID, Score: score})
	}
	return scores
}
