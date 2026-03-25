package refund

import "time"

const (
	StatusPending = "PENDING"
	StatusSuccess = "SUCCESS"
	StatusFailed  = "FAILED"
)

type Refund struct {
	ID         int64     `db:"id" json:"-"`
	RefundID   string    `db:"refund_id" json:"refund_id"`
	OrderID    string    `db:"order_id" json:"order_id"`
	RefundType string    `db:"refund_type" json:"refund_type"`
	Status     string    `db:"status" json:"status"`
	RetryCount int       `db:"retry_count" json:"retry_count"`
	NextRetry  time.Time `db:"next_retry_time" json:"next_retry_time"`
	LastError  string    `db:"last_error" json:"last_error"`
	Reason     string    `db:"reason" json:"reason"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

type InitiateRequest struct {
	RefundID   string `json:"refund_id"`
	OrderID    string `json:"order_id"`
	RefundType string `json:"refund_type"`
	Reason     string `json:"reason"`
}

type StatusRequest struct {
	RefundID string `json:"refund_id"`
}

type RollbackRequest struct {
	RefundID string `json:"refund_id"`
	OrderID  string `json:"order_id"`
	Reason   string `json:"reason"`
}
