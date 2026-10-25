-- R1: 库存分层扣减配套表
-- stock_journal: Redis 预扣流水的 DB 真账（异步落库，唯一键保证幂等）
CREATE TABLE IF NOT EXISTS stock_journal (
    id          BIGINT AUTO_INCREMENT PRIMARY KEY,
    sku_id      VARCHAR(64) NOT NULL,
    request_id  VARCHAR(128) NOT NULL,
    quantity    INT NOT NULL,
    created_at  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_sku_request (sku_id, request_id),
    KEY idx_sku_created (sku_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存预扣流水真账';

-- reconcile_diff: 对账差异台账（状态机 PENDING -> RESOLVED/IGNORED）
CREATE TABLE IF NOT EXISTS reconcile_diff (
    id          BIGINT AUTO_INCREMENT PRIMARY KEY,
    biz_type    VARCHAR(32) NOT NULL DEFAULT 'STOCK',
    sku_id      VARCHAR(64) NOT NULL,
    request_id  VARCHAR(128) NOT NULL DEFAULT '',
    diff_type   VARCHAR(32) NOT NULL COMMENT 'REDIS_NOT_SYNCED / AVAIL_MISMATCH',
    detail      VARCHAR(512) NOT NULL DEFAULT '',
    status      VARCHAR(16) NOT NULL DEFAULT 'PENDING',
    created_at  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    resolved_at DATETIME(3) NULL,
    UNIQUE KEY uk_diff (biz_type, sku_id, request_id, diff_type),
    KEY idx_status_created (status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='对账差异台账';
