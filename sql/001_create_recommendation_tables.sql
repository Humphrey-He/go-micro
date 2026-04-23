-- ============================================================================
-- Recommendation Service Database Tables
-- ============================================================================
-- This file contains all tables required for the recommendation service
-- including user behavior tracking, product similarity, and bestseller lists.
-- ============================================================================

-- 1. user_behavior_logs - User behavior log
-- Tracks user interactions with products across different sources
CREATE TABLE IF NOT EXISTS user_behavior_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'User ID',
    anonymous_id VARCHAR(64) COMMENT 'Anonymous user ID',
    sku_id BIGINT UNSIGNED NOT NULL COMMENT 'Product SKU ID',
    behavior_type ENUM('cart', 'favorite', 'purchase') NOT NULL COMMENT 'Behavior type',
    source VARCHAR(32) DEFAULT 'unknown' COMMENT 'Source: detail/cart/favorite/recommendation',
    stay_duration INT COMMENT 'Stay duration in seconds',
    time_bucket INT NOT NULL COMMENT '5-minute time bucket: FLOOR(UNIX_TIMESTAMP/300)',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_user_type_time (user_id, behavior_type, created_at),
    INDEX idx_sku_id (sku_id),
    INDEX idx_anonymous (anonymous_id),
    INDEX idx_created_at (created_at),
    UNIQUE KEY uk_dedup (user_id, anonymous_id, sku_id, behavior_type, time_bucket)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='User behavior log';

-- 2. product_similarity - Product similarity for Item-CF
-- Stores co-occurrence based product similarity scores
CREATE TABLE IF NOT EXISTS product_similarity (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sku_id_a BIGINT UNSIGNED NOT NULL COMMENT 'Product A',
    sku_id_b BIGINT UNSIGNED NOT NULL COMMENT 'Product B',
    scene ENUM('cart', 'favorite', 'purchase') NOT NULL COMMENT 'Scene',
    similarity DECIMAL(10,6) NOT NULL COMMENT 'Similarity score 0-1',
    weight INT DEFAULT 1 COMMENT 'Co-occurrence count',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (sku_id_a, sku_id_b, scene),
    INDEX idx_sku_a_scene (sku_id_a, scene),
    INDEX idx_similarity (similarity DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Product similarity';

-- 3. user_category_preference - User category preferences
-- Stores user preferences for product categories (explicit/implicit)
CREATE TABLE IF NOT EXISTS user_category_preference (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    category_id BIGINT UNSIGNED NOT NULL COMMENT 'Category ID',
    weight DECIMAL(10,4) DEFAULT 1.0 COMMENT 'Preference weight',
    source ENUM('explicit', 'implicit') DEFAULT 'implicit' COMMENT 'Explicit/implicit',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_user_category (user_id, category_id),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='User category preference';

-- 4. category_bestsellers - Category bestseller list
-- Stores bestseller rankings within each category
CREATE TABLE IF NOT EXISTS category_bestsellers (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    category_id BIGINT UNSIGNED NOT NULL,
    sku_id BIGINT UNSIGNED NOT NULL,
    sales_score DECIMAL(12,2) NOT NULL COMMENT 'Sales score',
    rank INT NOT NULL COMMENT 'Rank within category',
    period ENUM('7d', '30d') DEFAULT '30d' COMMENT 'Statistics period',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_category_sku_period (category_id, sku_id, period),
    INDEX idx_category_rank (category_id, period, rank)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Category bestsellers';

-- 5. global_bestsellers - Global bestseller list
-- Stores global bestseller rankings across all categories
CREATE TABLE IF NOT EXISTS global_bestsellers (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sku_id BIGINT UNSIGNED NOT NULL,
    sales_score DECIMAL(12,2) NOT NULL COMMENT 'Sales score',
    rank INT NOT NULL COMMENT 'Global rank',
    period ENUM('7d', '30d') DEFAULT '30d' COMMENT 'Statistics period',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_sku_period (sku_id, period),
    INDEX idx_rank_period (period, rank)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Global bestsellers';