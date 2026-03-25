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
	OrderNo         string `json:"order_no"`
	Status          string `json:"status"`
	InventoryStatus string `json:"inventory_status"`
	TaskStatus      string `json:"task_status"`
}
