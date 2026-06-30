-- Price Watch Tables
-- User price watch list
CREATE TABLE IF NOT EXISTS user_price_watches (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id         BIGINT UNSIGNED NOT NULL COMMENT 'User ID',
    sku_id          VARCHAR(64) NOT NULL COMMENT 'Product SKU ID',
    target_price    DECIMAL(12,2) COMMENT 'Target price, NULL means any drop',
    current_price   DECIMAL(12,2) COMMENT 'Price when watch was set',
    notify_enabled  TINYINT DEFAULT 1 COMMENT 'Notification switch: 1 on, 0 off',
    last_notify_at  DATETIME COMMENT 'Last notification time',
    notify_count    INT DEFAULT 0 COMMENT 'Total notification count',
    status          ENUM('active', 'stopped', 'expired') DEFAULT 'active' COMMENT 'Watch status',
    expire_at       DATE COMMENT 'Expiration date, NULL means never expires',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_user_sku (user_id, sku_id),
    INDEX idx_user_id (user_id),
    INDEX idx_sku_id (sku_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='User price watch list';

-- Price watch notifications
CREATE TABLE IF NOT EXISTS price_watch_notifications (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id         BIGINT UNSIGNED NOT NULL COMMENT 'User ID',
    watch_id        BIGINT UNSIGNED NOT NULL COMMENT 'Watch ID',
    sku_id          VARCHAR(64) NOT NULL COMMENT 'Product SKU ID',
    old_price       DECIMAL(12,2) NOT NULL COMMENT 'Price before drop',
    new_price       DECIMAL(12,2) NOT NULL COMMENT 'Price after drop',
    discount_amount DECIMAL(12,2) NOT NULL COMMENT 'Discount amount',
    discount_rate   DECIMAL(5,2) COMMENT 'Discount rate (percentage)',
    is_read         TINYINT DEFAULT 0 COMMENT '0 unread, 1 read',
    read_at         DATETIME COMMENT 'Read time',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_user_id (user_id),
    INDEX idx_watch_id (watch_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Price watch notifications';

-- Product price snapshots (for price history)
CREATE TABLE IF NOT EXISTS product_price_snapshots (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sku_id          VARCHAR(64) NOT NULL COMMENT 'Product SKU ID',
    price           DECIMAL(12,2) NOT NULL COMMENT 'Price at snapshot time',
    original_price  DECIMAL(12,2) COMMENT 'Original price',
    snapshot_date   DATE NOT NULL COMMENT 'Snapshot date',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uk_sku_date (sku_id, snapshot_date),
    INDEX idx_sku_id (sku_id),
    INDEX idx_snapshot_date (snapshot_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Product price snapshots';
