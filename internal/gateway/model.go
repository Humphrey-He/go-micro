package gateway

type Item struct {
	SkuID    string `json:"skuId"`
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
	OrderNo           string     `json:"order_no"`
	ViewStatus        string     `json:"view_status"`
	OrderStatus       string     `json:"order_status"`
	TaskStatus        string     `json:"task_status"`
	ReservationStatus string     `json:"reservation_status"`
	CancelReason      string     `json:"cancel_reason"`
}

type OrderDetailData struct {
	OrderID        string `json:"orderId"`
	BizNo          string `json:"bizNo"`
	UserID         string `json:"userId"`
	Status         string `json:"status"`
	TotalAmount    int64  `json:"totalAmount"`
	Items          []Item `json:"items"`
	PaymentStatus  string `json:"paymentStatus"`
	ViewStatus     string `json:"viewStatus"`
}

type ListOrdersRequest struct {
	Page      int32  `json:"page"`
	PageSize  int32  `json:"page_size"`
	OrderNo   string `json:"order_no"`
	UserID    string `json:"user_id"`
	Status    string `json:"status"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
}
