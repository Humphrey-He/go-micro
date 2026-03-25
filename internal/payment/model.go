package payment

import "time"

type CreatePaymentRequest struct {
	OrderID   string `json:"order_id"`
	Amount    int64  `json:"amount"`
	RequestID string `json:"request_id"`
}

type Payment struct {
	ID        int64     `db:"id" json:"-"`
	PaymentID string    `db:"payment_id" json:"payment_id"`
	OrderID   string    `db:"order_id" json:"order_id"`
	Amount    int64     `db:"amount" json:"amount"`
	Status    string    `db:"status" json:"status"`
	RequestID string    `db:"request_id" json:"request_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
