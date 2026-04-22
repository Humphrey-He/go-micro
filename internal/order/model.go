package order

import "time"

type Item struct {
	SkuID    string `json:"sku_id" db:"sku_id"`
	Quantity int    `json:"quantity" db:"quantity"`
	Price    int64  `json:"price" db:"price"`
}

type CreateOrderRequest struct {
	RequestID string `json:"request_id"`
	UserID    string `json:"user_id"`
	Items     []Item `json:"items"`
	Remark    string `json:"remark"`
}

type CreateOrderResponse struct {
	OrderID string `json:"order_id"`
	BizNo   string `json:"biz_no"`
	Status  string `json:"status"`
}

type Order struct {
	ID            int64     `db:"id" json:"-"`
	OrderID       string    `db:"order_id" json:"order_id"`
	BizNo         string    `db:"biz_no" json:"biz_no"`
	UserID        string    `db:"user_id" json:"user_id"`
	Status        string    `db:"status" json:"status"`
	TotalAmount   int64     `db:"total_amount" json:"total_amount"`
	IdempotentKey string    `db:"idempotent_key" json:"-"`
	ReservedID    string    `db:"reserved_id" json:"reserved_id"`
	Version       int64     `db:"version" json:"version"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
	Items         []Item    `json:"items"`
}

type ListOrdersRequest struct {
	Page      int32
	PageSize  int32
	OrderNo   string
	UserID    string
	Status    string
	StartTime int64
	EndTime   int64
	SortBy    string
	SortOrder string
}

type ListOrdersResponse struct {
	Orders   []OrderListItem `json:"orders"`
	Total    int32           `json:"total"`
	Page     int32           `json:"page"`
	PageSize int32           `json:"page_size"`
}

type OrderListItem struct {
	OrderID       string `json:"order_id"`
	BizNo         string `json:"biz_no"`
	UserID        string `json:"user_id"`
	Status        string `json:"status"`
	TotalAmount   int64  `json:"total_amount"`
	ItemCount     int    `json:"item_count"`
	PaymentStatus string `json:"payment_status"`
	CreatedAt     int64  `json:"created_at"`
	UpdatedAt     int64  `json:"updated_at"`
}
