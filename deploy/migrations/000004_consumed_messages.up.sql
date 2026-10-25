-- R2: 消费端幂等表（消息 ID + 消费组唯一约束）
CREATE TABLE IF NOT EXISTS consumed_messages (
    id             BIGINT AUTO_INCREMENT PRIMARY KEY,
    consumer_group VARCHAR(64) NOT NULL,
    message_id     VARCHAR(128) NOT NULL,
    consumed_at    DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_group_message (consumer_group, message_id),
    KEY idx_consumed_at (consumed_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消费端幂等记录';
