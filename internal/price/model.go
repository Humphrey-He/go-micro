package price

import (
	"time"

	"github.com/shopspring/decimal"
)

type History struct {
	ID        int64           `db:"id" json:"-"`
	SkuID     string          `db:"sku_id" json:"sku_id"`
	OldPrice  decimal.Decimal `db:"old_price" json:"old_price"`
	NewPrice  decimal.Decimal `db:"new_price" json:"new_price"`
	Reason    string          `db:"reason" json:"reason"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
}

type CalculateRequest struct {
	SkuID        string          `json:"sku_id"`
	BasePrice    decimal.Decimal `json:"base_price"`
	CouponAmount decimal.Decimal `json:"coupon_amount"`
	UserLevel    int             `json:"user_level"`
}

type CalculateResponse struct {
	SkuID          string          `json:"sku_id"`
	BasePrice      decimal.Decimal `json:"base_price"`
	CouponAmount   decimal.Decimal `json:"coupon_amount"`
	LevelDiscount  decimal.Decimal `json:"level_discount"`
	FinalPrice     decimal.Decimal `json:"final_price"`
	DiscountReason string          `json:"discount_reason"`
}