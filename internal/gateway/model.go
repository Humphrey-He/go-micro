package gateway

type Item struct {
	SkuID    string `json:"sku_id"`
	Quantity int    `json:"quantity"`
	Price    int64  `json:"price"`
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

type OrderViewResponse struct {
	OrderNo           string `json:"order_no"`
	ViewStatus        string `json:"view_status"`
	OrderStatus       string `json:"order_status"`
	TaskStatus        string `json:"task_status"`
	ReservationStatus string `json:"reservation_status"`
	CancelReason      string `json:"cancel_reason"`
}
