package pricewatch

import "time"

type PriceWatch struct {
	ID            int64      `db:"id" json:"watch_id"`
	UserID       int64      `db:"user_id" json:"user_id"`
	SkuID        string     `db:"sku_id" json:"sku_id"`
	TargetPrice  float64    `db:"target_price" json:"target_price,omitempty"`
	CurrentPrice float64    `db:"current_price" json:"current_price"`
	NotifyEnabled bool      `db:"notify_enabled" json:"notify_enabled"`
	LastNotifyAt time.Time `db:"last_notify_at" json:"last_notify_at,omitempty"`
	NotifyCount  int       `db:"notify_count" json:"notify_count"`
	Status       string     `db:"status" json:"status"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
}

type PriceWatchNotification struct {
	ID              int64     `db:"id" json:"notification_id"`
	UserID         int64     `db:"user_id" json:"user_id"`
	WatchID        int64     `db:"watch_id" json:"watch_id"`
	SkuID          string    `db:"sku_id" json:"sku_id"`
	OldPrice       float64   `db:"old_price" json:"old_price"`
	NewPrice       float64   `db:"new_price" json:"new_price"`
	DiscountAmount float64   `db:"discount_amount" json:"discount_amount"`
	DiscountRate   float64   `db:"discount_rate" json:"discount_rate"`
	IsRead         bool      `db:"is_read" json:"is_read"`
	ReadAt         *time.Time `db:"read_at" json:"read_at,omitempty"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

// Request/Response types

type SetPriceWatchRequest struct {
	SkuID       string   `json:"sku_id" binding:"required"`
	TargetPrice *float64 `json:"target_price"`
}

type SetPriceWatchResponse struct {
	WatchID      int64    `json:"watch_id"`
	CurrentPrice float64  `json:"current_price"`
	TargetPrice  *float64 `json:"target_price,omitempty"`
}

type UpdatePriceWatchRequest struct {
	TargetPrice   *float64 `json:"target_price"`
	NotifyEnabled *bool    `json:"notify_enabled"`
}

type PriceWatchListResponse struct {
	Items    []PriceWatchItem `json:"items"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
	Total    int             `json:"total"`
}

type PriceWatchItem struct {
	WatchID        int64   `json:"watch_id"`
	SkuID         string  `json:"sku_id"`
	ProductName   string  `json:"product_name"`
	Image         string  `json:"image"`
	CurrentPrice  float64 `json:"current_price"`
	OriginalPrice float64 `json:"original_price"`
	TargetPrice   *float64 `json:"target_price,omitempty"`
	NotifyEnabled bool    `json:"notify_enabled"`
	LowestPrice  float64 `json:"lowest_price"`
	LowestPriceDate string `json:"lowest_price_date,omitempty"`
	CreatedAt    string  `json:"created_at"`
	PriceTrend   string  `json:"price_trend"` // down, stable, up
}

type PriceHistoryResponse struct {
	SkuID         string       `json:"sku_id"`
	CurrentPrice  float64      `json:"current_price"`
	LowestPrice   float64      `json:"lowest_price"`
	LowestDate    string       `json:"lowest_date,omitempty"`
	HighestPrice  float64      `json:"highest_price"`
	HighestDate   string       `json:"highest_date,omitempty"`
	AveragePrice  float64      `json:"average_price"`
	PricePoints   []PricePoint `json:"price_points"`
}

type PricePoint struct {
	Date  string  `json:"date"`
	Price float64 `json:"price"`
}

type NotificationListResponse struct {
	Items      []NotificationItem `json:"items"`
	UnreadCount int               `json:"unread_count"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	Total      int                `json:"total"`
}

type NotificationItem struct {
	NotificationID int64   `json:"notification_id"`
	SkuID         string  `json:"sku_id"`
	ProductName   string  `json:"product_name"`
	Image         string  `json:"image"`
	OldPrice      float64 `json:"old_price"`
	NewPrice      float64 `json:"new_price"`
	DiscountAmount float64 `json:"discount_amount"`
	DiscountRate  float64 `json:"discount_rate"`
	IsRead        bool    `json:"is_read"`
	CreatedAt     string  `json:"created_at"`
}
