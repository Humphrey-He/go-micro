-- Association Rules for Cart Add-on Recommendations
-- Stores mined association rules: when users buy A, they often buy B

-- Association rules table
CREATE TABLE IF NOT EXISTS association_rules (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sku_id_a BIGINT UNSIGNED NOT NULL COMMENT 'Premise product (A in A->B)',
    sku_id_b BIGINT UNSIGNED NOT NULL COMMENT 'Consequent product (B in A->B)',
    support DECIMAL(10,6) NOT NULL COMMENT 'Support: P(A∪B) - fraction of all transactions containing both',
    confidence DECIMAL(10,6) NOT NULL COMMENT 'Confidence: P(B|A) - fraction of A purchasers who also bought B',
    lift DECIMAL(10,6) NOT NULL COMMENT 'Lift: confidence/P(B) - how much A increases B probability',
    co_purchase_count INT NOT NULL DEFAULT 0 COMMENT 'Number of times A and B were purchased together',
    a_purchase_count INT NOT NULL DEFAULT 0 COMMENT 'Number of times A was purchased',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_rule (sku_id_a, sku_id_b),
    INDEX idx_sku_a (sku_id_a),
    INDEX idx_sku_b (sku_id_b),
    INDEX idx_lift (lift DESC),
    INDEX idx_confidence (confidence DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Association rules for cart add-on recommendations';

-- Frequent itemsets table (intermediate results from Apriori)
CREATE TABLE IF NOT EXISTS frequent_itemsets (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    item_set JSON NOT NULL COMMENT 'JSON array of sku_ids in the itemset',
    item_count INT NOT NULL COMMENT 'Number of items in the set',
    support DECIMAL(10,6) NOT NULL COMMENT 'Support: fraction of transactions containing this set',
    transaction_count INT NOT NULL DEFAULT 0 COMMENT 'Number of transactions containing this set',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_item_count (item_count),
    INDEX idx_support (support DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Frequent itemsets from Apriori algorithm';
