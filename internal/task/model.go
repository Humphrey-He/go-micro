package task

import "time"

type CreateTaskRequest struct {
	BizNo   string `json:"biz_no"`
	OrderID string `json:"order_id"`
	Type    string `json:"type"`
}

type Task struct {
	ID          int64     `db:"id" json:"-"`
	TaskID      string    `db:"task_id" json:"task_id"`
	BizNo       string    `db:"biz_no" json:"biz_no"`
	OrderID     string    `db:"order_id" json:"order_id"`
	Type        string    `db:"type" json:"type"`
	Status      string    `db:"status" json:"status"`
	RetryCount  int       `db:"retry_count" json:"retry_count"`
	NextRetryAt time.Time `db:"next_retry_at" json:"next_retry_at"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type Saga struct {
	ID        int64     `db:"id" json:"-"`
	SagaID    string    `db:"saga_id" json:"saga_id"`
	BizNo     string    `db:"biz_no" json:"biz_no"`
	Type      string    `db:"type" json:"type"`
	Status    string    `db:"status" json:"status"`
	Reason    string    `db:"reason" json:"reason"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
