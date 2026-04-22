package inventory

import "time"

type Item struct {
	SkuID    string `json:"sku_id" db:"sku_id"`
	Quantity int    `json:"quantity" db:"quantity"`
}

type ReserveRequest struct {
	OrderID string `json:"order_id"`
	Items   []Item `json:"items"`
}

type ReserveResponse struct {
	ReservedID string `json:"reserved_id"`
}

type ReleaseRequest struct {
	ReservedID string `json:"reserved_id"`
}

type ReleaseByOrderRequest struct {
	OrderID string `json:"order_id"`
}

type ConfirmRequest struct {
	ReservedID string `json:"reserved_id"`
}

type Inventory struct {
	ID        int64     `db:"id" json:"-"`
	SkuID     string    `db:"sku_id" json:"sku_id"`
	Available int       `db:"available" json:"available"`
	Reserved  int       `db:"reserved" json:"reserved"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type InventoryItem struct {
	SkuID     string `json:"sku_id"`
	Available int    `json:"available"`
	Reserved  int    `json:"reserved"`
}

type Reservation struct {
	ID         int64     `db:"id" json:"-"`
	ReservedID string    `db:"reserved_id" json:"reserved_id"`
	OrderID    string    `db:"order_id" json:"order_id"`
	Status     string    `db:"status" json:"status"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}
