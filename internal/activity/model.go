package activity

import "time"

const (
	CouponStatusIssued = "ISSUED"
	CouponStatusUsed   = "USED"

	SeckillStatusActive  = "ACTIVE"
	SeckillStatusClosed  = "CLOSED"
	OrderStatusReserved  = "RESERVED"
	OrderStatusCompleted = "COMPLETED"
)

type Coupon struct {
	ID        int64     `db:"id" json:"-"`
	CouponID  string    `db:"coupon_id" json:"coupon_id"`
	UserID    string    `db:"user_id" json:"user_id"`
	Amount    float64   `db:"amount" json:"amount"`
	Status    string    `db:"status" json:"status"`
	IssuedAt  time.Time `db:"issued_at" json:"issued_at"`
	UsedAt    time.Time `db:"used_at" json:"used_at"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type Seckill struct {
	ID        int64     `db:"id" json:"-"`
	SkuID     string    `db:"sku_id" json:"sku_id"`
	Stock     int       `db:"stock" json:"stock"`
	Reserved  int       `db:"reserved" json:"reserved"`
	Status    string    `db:"status" json:"status"`
	StartTime time.Time `db:"start_time" json:"start_time"`
	EndTime   time.Time `db:"end_time" json:"end_time"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type SeckillOrder struct {
	ID        int64     `db:"id" json:"-"`
	SkuID     string    `db:"sku_id" json:"sku_id"`
	UserID    string    `db:"user_id" json:"user_id"`
	Quantity  int       `db:"quantity" json:"quantity"`
	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type CouponRequest struct {
	CouponID string  `json:"coupon_id"`
	UserID   string  `json:"user_id"`
	Amount   float64 `json:"amount"`
}

type SeckillRequest struct {
	SkuID    string `json:"sku_id"`
	UserID   string `json:"user_id"`
	Quantity int    `json:"quantity"`
}

type StatusQuery struct {
	CouponID string `form:"coupon_id" json:"coupon_id"`
	SkuID    string `form:"sku_id" json:"sku_id"`
}
