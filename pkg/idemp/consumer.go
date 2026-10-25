package idemp

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// ConsumeRecorder 基于 consumed_messages 表的消费端幂等：
// 消息 ID + 消费组唯一约束，INSERT IGNORE 撞键即判定为已消费。
// 配合 MQ 的 at-least-once 投递，构成「业务恰好一次」的消费侧半边
// （生产侧半边是本地消息表/Outbox）。
type ConsumeRecorder struct {
	db *sqlx.DB
}

func NewConsumeRecorder(dbx *sqlx.DB) *ConsumeRecorder {
	return &ConsumeRecorder{db: dbx}
}

// TryRecord 登记一次消费。返回 true 表示首次消费（应执行业务逻辑），
// false 表示该消息已被本消费组处理过（应直接 ACK 跳过）。
func (r *ConsumeRecorder) TryRecord(ctx context.Context, consumerGroup, messageID string) (bool, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT IGNORE INTO consumed_messages (consumer_group, message_id) VALUES (?, ?)`,
		consumerGroup, messageID)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}
