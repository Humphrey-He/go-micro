package recommendation

import "time"

type BehaviorType string

const (
	BehaviorCart     BehaviorType = "cart"
	BehaviorFavorite BehaviorType = "favorite"
	BehaviorPurchase BehaviorType = "purchase"
)

type BehaviorLog struct {
	ID           int64        `db:"id"`
	UserID       int64        `db:"user_id"`
	AnonymousID  string       `db:"anonymous_id"`
	SkuID        int64        `db:"sku_id"`
	BehaviorType BehaviorType `db:"behavior_type"`
	Source       string       `db:"source"`
	StayDuration int          `db:"stay_duration"`
	TimeBucket   int          `db:"time_bucket"`
	CreatedAt    time.Time    `db:"created_at"`
}

type ProductSimilarity struct {
	ID         int64     `db:"id"`
	SkuIDA     int64     `db:"sku_id_a"`
	SkuIDB     int64     `db:"sku_id_b"`
	Scene      string    `db:"scene"`
	Similarity float64   `db:"similarity"`
	Weight     int       `db:"weight"`
	UpdatedAt  time.Time `db:"updated_at"`
}

type UserCategoryPreference struct {
	ID         int64   `db:"id"`
	UserID     int64   `db:"user_id"`
	CategoryID int64   `db:"category_id"`
	Weight     float64 `db:"weight"`
	Source     string  `db:"source"`
}

type CategoryBestseller struct {
	ID         int64   `db:"id"`
	CategoryID int64   `db:"category_id"`
	SkuID      int64   `db:"sku_id"`
	SalesScore float64 `db:"sales_score"`
	Rank       int     `db:"rank"`
	Period     string  `db:"period"`
}

type GlobalBestseller struct {
	ID         int64   `db:"id"`
	SkuID      int64   `db:"sku_id"`
	SalesScore float64 `db:"sales_score"`
	Rank       int     `db:"rank"`
	Period     string  `db:"period"`
}

// RecItem - Recommended product item
type RecItem struct {
	SkuID      int64   `json:"sku_id"`
	Name       string  `json:"name"`
	Price      float64 `json:"price"`
	Image      string  `json:"image"`
	Similarity float64 `json:"similarity,omitempty"`
	Reason     string  `json:"reason,omitempty"`
}

// BehaviorReportRequest - Behavior report request
type BehaviorReportRequest struct {
	SkuID        int64  `json:"sku_id" form:"sku_id" binding:"required"`
	BehaviorType string `json:"behavior_type" form:"behavior_type" binding:"required,oneof=cart favorite purchase"`
	Source       string `json:"source" form:"source"`
	StayDuration int    `json:"stay_duration" form:"stay_duration"`
	AnonymousID  string `json:"anonymous_id" form:"anonymous_id"`
}

// SimilarProductsRequest - Similar products request
type SimilarProductsRequest struct {
	SkuID int64  `uri:"sku_id" binding:"required"`
	Scene string `form:"scene" binding:"omitempty,oneof=cart favorite purchase"`
	Limit int    `form:"limit"`
}

// HomeRecRequest - Home recommendation request
type HomeRecRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}