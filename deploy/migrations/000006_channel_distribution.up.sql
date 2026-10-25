-- R5: 渠道分发调度
-- channel_mappings: 商品 × 平台 × 站点 × 店铺（按店铺限频的映射基础）
CREATE TABLE IF NOT EXISTS channel_mappings (
    id            BIGINT AUTO_INCREMENT PRIMARY KEY,
    sku_id        VARCHAR(64) NOT NULL,
    platform      VARCHAR(32) NOT NULL,
    site          VARCHAR(32) NOT NULL,
    shop_id       VARCHAR(64) NOT NULL,
    rate_limit_qps INT NOT NULL DEFAULT 10,
    status        VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
    created_at    DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_sku_platform_site (sku_id, platform, site),
    KEY idx_shop (shop_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='渠道映射';

-- dispatch_tasks: 分发任务（client_msg_id 幂等键；补发复用同 ID）
CREATE TABLE IF NOT EXISTS dispatch_tasks (
    id               BIGINT AUTO_INCREMENT PRIMARY KEY,
    client_msg_id    VARCHAR(128) NOT NULL,
    batch_id         VARCHAR(64) NOT NULL,
    sku_id           VARCHAR(64) NOT NULL,
    platform         VARCHAR(32) NOT NULL,
    site             VARCHAR(32) NOT NULL,
    shop_id          VARCHAR(64) NOT NULL,
    status           VARCHAR(32) NOT NULL DEFAULT 'QUEUED',
    platform_task_id VARCHAR(64) NOT NULL DEFAULT '',
    redispatch_count INT NOT NULL DEFAULT 0,
    last_error       VARCHAR(255) NOT NULL DEFAULT '',
    sent_at          DATETIME(3) NULL,
    created_at       DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at       DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_client_msg (client_msg_id),
    KEY idx_status_id (status, id),
    KEY idx_status_sent (status, sent_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='多平台分发任务';
