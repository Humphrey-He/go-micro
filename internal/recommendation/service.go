package recommendation

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"go-micro/internal/recommendation/algorithm"
)

type Service struct {
	db         *sqlx.DB
	redis     *redis.Client
	cache     *Cache
	itemCF    *algorithm.ItemCF
	userCF    *algorithm.UserCF
	assoc     *algorithm.Association
}

const FiveMinuteBucket = 300

func NewService(db *sqlx.DB, redis *redis.Client, cache *Cache) *Service {
	return &Service{
		db:    db,
		redis: redis,
		cache: cache,
		itemCF: algorithm.NewItemCF(db),
		userCF: algorithm.NewUserCF(db),
		assoc: algorithm.NewAssociation(db),
	}
}

// ReportBehavior - Report user behavior
func (s *Service) ReportBehavior(ctx context.Context, req *BehaviorReportRequest, userID int64) error {
	// Calculate 5-minute time bucket
	timeBucket := int(time.Now().Unix() / FiveMinuteBucket)

	_, err := s.db.Exec(`
		INSERT INTO user_behavior_logs (user_id, anonymous_id, sku_id, behavior_type, source, stay_duration, time_bucket, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE id=id
	`, userID, req.AnonymousID, req.SkuID, req.BehaviorType, req.Source, req.StayDuration, timeBucket)

	return err
}

// Response types

// SimilarProductsResponse - Similar products response
type SimilarProductsResponse struct {
	Scene string    `json:"scene"`
	Items []RecItem `json:"items"`
}

// HomeRecResponse - Home recommendation response
type HomeRecResponse struct {
	Items    []RecItem `json:"items"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
	Total    int       `json:"total"`
	Source   string    `json:"source"` // personalized | category | global
}

// ColdStartResponse - Cold start response
type ColdStartResponse struct {
	HotItems      []RecItem       `json:"hot_items"`
	CategoryPrefs []CategoryPref  `json:"category_prefs"`
}

// CategoryPref - Category preference
type CategoryPref struct {
	CategoryID int64  `json:"category_id"`
	Name       string `json:"name"`
	Image      string `json:"image"`
}

// GetSimilarProducts - Get similar products
func (s *Service) GetSimilarProducts(ctx context.Context, skuID int64, scene string, limit int) (*SimilarProductsResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	// Try cache first
	if s.cache != nil {
		if items, ok := s.cache.GetSimilarProducts(ctx, skuID, scene); ok {
			return &SimilarProductsResponse{Scene: scene, Items: items}, nil
		}
	}

	// Cache miss - compute from database
	// TODO: Implement actual DB query using ItemCF

	items := []RecItem{} // Placeholder

	// Enrich items (TODO: fetch real product info)
	for i := range items {
		if items[i].Name == "" {
			items[i].Name = fmt.Sprintf("Product %d", items[i].SkuID)
		}
	}

	// Update cache
	if s.cache != nil && len(items) > 0 {
		s.cache.SetSimilarProducts(ctx, skuID, scene, items)
	}

	return &SimilarProductsResponse{
		Scene: scene,
		Items: items,
	}, nil
}

// GetHomeRecommendations - Get home recommendations
func (s *Service) GetHomeRecommendations(ctx context.Context, userID int64, page, pageSize int) (*HomeRecResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	// Try cache first
	if s.cache != nil {
		if resp, ok := s.cache.GetHomeRec(ctx, userID, page, pageSize); ok {
			return resp, nil
		}
	}

	// Cache miss - compute recommendations
	// For now, return global bestsellers as fallback

	items := []RecItem{} // TODO: Implement User-CF

	// Enrich items
	for i := range items {
		if items[i].Name == "" {
			items[i].Name = fmt.Sprintf("Product %d", items[i].SkuID)
		}
	}

	resp := &HomeRecResponse{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    len(items),
		Source:   "global",
	}

	// Update cache
	if s.cache != nil && len(items) > 0 {
		s.cache.SetHomeRec(ctx, userID, page, pageSize, resp)
	}

	return resp, nil
}

// GetColdStartData - Get cold start data
func (s *Service) GetColdStartData(ctx context.Context) (*ColdStartResponse, error) {
	return &ColdStartResponse{
		HotItems: []RecItem{},
		CategoryPrefs: []CategoryPref{
			{CategoryID: 1, Name: "手机数码", Image: "https://placeholder.com/phone.png"},
			{CategoryID: 2, Name: "服装鞋帽", Image: "https://placeholder.com/clothing.png"},
			{CategoryID: 3, Name: "家用电器", Image: "https://placeholder.com/electronics.png"},
		},
	}, nil
}

// SetUserPreference - Set user preference
func (s *Service) SetUserPreference(ctx context.Context, userID int64, categoryIDs []int64) error {
	if userID <= 0 || len(categoryIDs) == 0 {
		return nil
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear old explicit preferences
	_, err = tx.Exec(`DELETE FROM user_category_preference WHERE user_id = ? AND source = 'explicit'`, userID)
	if err != nil {
		return err
	}

	// Insert new preferences
	for _, catID := range categoryIDs {
		_, err := tx.Exec(`
			INSERT INTO user_category_preference (user_id, category_id, weight, source)
			VALUES (?, ?, 1.0, 'explicit')
		`, userID, catID)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	// Clear cache
	cacheKey := fmt.Sprintf("rec:user_pref:%d", userID)
	s.redis.Del(ctx, cacheKey)

	return nil
}

// CartAddonResponse - Cart add-on recommendation response
type CartAddonResponse struct {
	Items []RecItem `json:"items"`
}

// GetCartAddons - Get cart add-on recommendations
func (s *Service) GetCartAddons(ctx context.Context, cartSKUIDs []int64, limit int) (*CartAddonResponse, error) {
	if len(cartSKUIDs) == 0 {
		return &CartAddonResponse{Items: []RecItem{}}, nil
	}
	if limit <= 0 {
		limit = 10
	}

	algoItems, err := s.assoc.GetCartAddons(cartSKUIDs, limit)
	if err != nil {
		return nil, err
	}

	items := make([]RecItem, len(algoItems))
	for i, item := range algoItems {
		items[i] = RecItem{
			SkuID:      item.SkuID,
			Name:       item.Name,
			Price:      item.Price,
			Image:      item.Image,
			Similarity: item.Similarity,
			Reason:     item.Reason,
		}
	}

	return &CartAddonResponse{Items: items}, nil
}

// PayCompleteResponse - Pay complete recommendation response
type PayCompleteResponse struct {
	Items []RecItem `json:"items"`
}

// GetPayCompleteRecommendations - Get recommendations after purchase
func (s *Service) GetPayCompleteRecommendations(ctx context.Context, purchasedSKUIDs []int64, limit int) (*PayCompleteResponse, error) {
	if len(purchasedSKUIDs) == 0 {
		return &PayCompleteResponse{Items: []RecItem{}}, nil
	}
	if limit <= 0 {
		limit = 10
	}

	algoItems, err := s.assoc.GetPayCompleteRecommendations(purchasedSKUIDs, limit)
	if err != nil {
		return nil, err
	}

	items := make([]RecItem, len(algoItems))
	for i, item := range algoItems {
		items[i] = RecItem{
			SkuID:      item.SkuID,
			Name:       item.Name,
			Price:      item.Price,
			Image:      item.Image,
			Similarity: item.Similarity,
			Reason:     item.Reason,
		}
	}

	return &PayCompleteResponse{Items: items}, nil
}