package distribution

import "time"

// 渠道映射：商品 × 平台 × 站点 × 店铺
type ChannelMapping struct {
	ID           int64  `db:"id" json:"id"`
	SkuID        string `db:"sku_id" json:"sku_id"`
	Platform     string `db:"platform" json:"platform"`
	Site         string `db:"site" json:"site"`
	ShopID       string `db:"shop_id" json:"shop_id"`
	RateLimitQPS int    `db:"rate_limit_qps" json:"rate_limit_qps"`
	Status       string `db:"status" json:"status"`
}

// 分发任务状态机：
// QUEUED ──令牌──► PENDING_ACK ──回执──► CONFIRMED
//
//	▲                  │回执超时（diff job）
//	└──429/失败/补发────┴──► （记台账 RECEIPT_MISSING 后重新入队）
const (
	TaskQueued     = "QUEUED"
	TaskPendingAck = "PENDING_ACK"
	TaskConfirmed  = "CONFIRMED"
)

type DispatchTask struct {
	ID              int64      `db:"id" json:"id"`
	ClientMsgID     string     `db:"client_msg_id" json:"client_msg_id"`
	BatchID         string     `db:"batch_id" json:"batch_id"`
	SkuID           string     `db:"sku_id" json:"sku_id"`
	Platform        string     `db:"platform" json:"platform"`
	Site            string     `db:"site" json:"site"`
	ShopID          string     `db:"shop_id" json:"shop_id"`
	Status          string     `db:"status" json:"status"`
	PlatformTaskID  string     `db:"platform_task_id" json:"platform_task_id"`
	RedispatchCount int        `db:"redispatch_count" json:"redispatch_count"`
	LastError       string     `db:"last_error" json:"last_error"`
	SentAt          *time.Time `db:"sent_at" json:"sent_at,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// 与 R4 共用 reconcile_diff 台账：biz_type=DISTRIBUTION
const (
	LedgerBizType  = "DISTRIBUTION"
	LedgerDiffType = "RECEIPT_MISSING"
)
