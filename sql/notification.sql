-- 通知记录表
CREATE TABLE IF NOT EXISTS notifications (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(64) NOT NULL,
    type VARCHAR(32) NOT NULL,
    title VARCHAR(128) NOT NULL,
    content TEXT,
    is_read TINYINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_read (user_id, is_read),
    INDEX idx_created (created_at)
);

-- 通知订阅配置表
CREATE TABLE IF NOT EXISTS notification_configs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(64) NOT NULL,
    type VARCHAR(32) NOT NULL,
    email_enabled TINYINT DEFAULT 1,
    push_enabled TINYINT DEFAULT 1,
    threshold INT DEFAULT 0,
    UNIQUE INDEX idx_user_type (user_id, type)
);