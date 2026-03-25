package price

import "time"

type History struct {
	ID        int64     `db:"id" json:"-"`
	SkuID     string    `db:"sku_id" json:"sku_id"`
	OldPrice  float64   `db:"old_price" json:"old_price"`
	NewPrice  float64   `db:"new_price" json:"new_price"`
	Reason    string    `db:"reason" json:"reason"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type CalculateRequest struct {
	SkuID        string  `json:"sku_id"`
	BasePrice    float64 `json:"base_price"`
	CouponAmount float64 `json:"coupon_amount"`
	UserLevel    int     `json:"user_level"`
}

type CalculateResponse struct {
	SkuID          string  `json:"sku_id"`
	BasePrice      float64 `json:"base_price"`
	CouponAmount   float64 `json:"coupon_amount"`
	LevelDiscount  float64 `json:"level_discount"`
	FinalPrice     float64 `json:"final_price"`
	DiscountReason string  `json:"discount_reason"`
}
