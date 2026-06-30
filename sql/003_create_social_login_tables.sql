-- Social Login Tables
-- Add social login fields to users table
ALTER TABLE users ADD COLUMN social_provider ENUM('none', 'wechat', 'google', 'apple') DEFAULT 'none' COMMENT '主要社交登录平台';
ALTER TABLE users ADD COLUMN social_uid VARCHAR(128) COMMENT '主要社交平台用户ID';
ALTER TABLE users ADD COLUMN is_phone_bound TINYINT DEFAULT 0 COMMENT '是否已绑定手机号';
ALTER TABLE users ADD COLUMN created_source ENUM('phone', 'wechat', 'google', 'apple') COMMENT '注册来源';

-- User social bindings table
CREATE TABLE IF NOT EXISTS user_social_bindings (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id         BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    provider        ENUM('wechat', 'google', 'apple') NOT NULL COMMENT '登录平台',
    provider_user_id VARCHAR(128) NOT NULL COMMENT '第三方平台用户ID',
    access_token    TEXT COMMENT '第三方 access_token (微信专有)',
    refresh_token   TEXT COMMENT '第三方 refresh_token (微信专有，7天有效)',
    token_expires_at DATETIME COMMENT 'access_token 过期时间',
    union_id        VARCHAR(128) COMMENT '微信 UnionID (跨应用识别)',
    nickname        VARCHAR(64) COMMENT '第三方昵称(首次同步)',
    avatar          VARCHAR(512) COMMENT '第三方头像URL',
    bind_status     TINYINT DEFAULT 1 COMMENT '绑定状态: 1正常, 0解绑',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_provider_user (provider, provider_user_id),
    KEY idx_user_id (user_id),
    KEY idx_union_id (union_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户社交账号绑定表';
