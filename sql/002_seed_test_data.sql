-- ============================================================================
-- Recommendation Service Test Data
-- ============================================================================
-- Minimal test data for validating recommendation service functionality
-- ============================================================================

-- User behavior logs test data
INSERT INTO user_behavior_logs (user_id, sku_id, behavior_type, source, time_bucket) VALUES
(1, 1001, 'cart', 'detail', FLOOR(UNIX_TIMESTAMP(NOW()) / 300)),
(1, 1002, 'cart', 'detail', FLOOR(UNIX_TIMESTAMP(NOW()) / 300)),
(1, 1001, 'purchase', 'detail', FLOOR(UNIX_TIMESTAMP(NOW()) / 300)),
(2, 1001, 'favorite', 'detail', FLOOR(UNIX_TIMESTAMP(NOW()) / 300)),
(2, 1003, 'purchase', 'detail', FLOOR(UNIX_TIMESTAMP(NOW()) / 300)),
(3, 1002, 'cart', 'detail', FLOOR(UNIX_TIMESTAMP(NOW()) / 300)),
(3, 1001, 'purchase', 'cart', FLOOR(UNIX_TIMESTAMP(NOW()) / 300));