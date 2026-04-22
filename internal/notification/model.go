package notification

import "time"

type Notification struct {
    ID        int64     `json:"id" db:"id"`
    UserID    string    `json:"user_id" db:"user_id"`
    Type      string    `json:"type" db:"type"`
    Title     string    `json:"title" db:"title"`
    Content   string    `json:"content" db:"content"`
    IsRead    bool      `json:"is_read" db:"is_read"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type NotificationConfig struct {
    ID           int64  `json:"id" db:"id"`
    UserID       string `json:"user_id" db:"user_id"`
    Type         string `json:"type" db:"type"`
    EmailEnabled bool   `json:"email_enabled" db:"email_enabled"`
    PushEnabled  bool   `json:"push_enabled" db:"push_enabled"`
    Threshold    int    `json:"threshold" db:"threshold"`
}

// NotificationType 常量
const (
    TypeRefundPending = "refund_pending"
    TypeLowStock      = "low_stock"
    TypePaymentFailed = "payment_failed"
    TypeDailyReport   = "daily_report"
    TypeWeeklyReport  = "weekly_report"
)