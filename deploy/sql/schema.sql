CREATE DATABASE IF NOT EXISTS go_micro DEFAULT CHARSET utf8mb4;
USE go_micro;

CREATE TABLE IF NOT EXISTS users (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  user_id VARCHAR(64) NOT NULL UNIQUE,
  username VARCHAR(64) NOT NULL UNIQUE,
  phone VARCHAR(32) UNIQUE COMMENT '手机号',
  password_hash VARCHAR(128) NOT NULL,
  role VARCHAR(32) NOT NULL DEFAULT 'user',
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS orders (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  order_id VARCHAR(64) NOT NULL UNIQUE,
  biz_no VARCHAR(64) NOT NULL UNIQUE,
  user_id VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL,
  total_amount BIGINT NOT NULL,
  idempotent_key VARCHAR(128) NOT NULL UNIQUE,
  reserved_id VARCHAR(64) NOT NULL DEFAULT '',
  version BIGINT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  INDEX idx_orders_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS order_items (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  order_id VARCHAR(64) NOT NULL,
  sku_id VARCHAR(64) NOT NULL,
  quantity INT NOT NULL,
  price BIGINT NOT NULL,
  INDEX idx_order_items_order_id (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS order_outbox (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  event_type VARCHAR(64) NOT NULL,
  payload JSON NOT NULL,
  status VARCHAR(32) NOT NULL,
  retry_count INT NOT NULL DEFAULT 0,
  last_error VARCHAR(512) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL,
  sent_at DATETIME NULL,
  INDEX idx_outbox_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS inventory (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  sku_id VARCHAR(64) NOT NULL UNIQUE,
  available INT NOT NULL,
  reserved INT NOT NULL DEFAULT 0,
  updated_at DATETIME NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS inventory_reserved (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  reserved_id VARCHAR(64) NOT NULL UNIQUE,
  order_id VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  INDEX idx_inventory_reserved_order_id (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS inventory_reserved_item (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  reserved_id VARCHAR(64) NOT NULL,
  sku_id VARCHAR(64) NOT NULL,
  quantity INT NOT NULL,
  INDEX idx_inventory_reserved_item_reserved_id (reserved_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS tasks (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  task_id VARCHAR(64) NOT NULL UNIQUE,
  biz_no VARCHAR(64) NOT NULL,
  order_id VARCHAR(64) NOT NULL,
  type VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL,
  retry_count INT NOT NULL DEFAULT 0,
  next_retry_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  INDEX idx_tasks_order_id (order_id),
  UNIQUE KEY uniq_tasks_order_type (order_id, type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS sagas (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  saga_id VARCHAR(64) NOT NULL UNIQUE,
  biz_no VARCHAR(64) NOT NULL,
  type VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL,
  reason VARCHAR(128) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  INDEX idx_sagas_biz_no (biz_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS saga_steps (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  saga_id VARCHAR(64) NOT NULL,
  step VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL,
  reason VARCHAR(128) NOT NULL DEFAULT '',
  payload JSON NOT NULL,
  next_step VARCHAR(64) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  UNIQUE KEY uniq_saga_step (saga_id, step),
  INDEX idx_saga_steps_saga_id (saga_id),
  INDEX idx_saga_steps_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS payments (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  payment_id VARCHAR(64) NOT NULL UNIQUE,
  order_id VARCHAR(64) NOT NULL,
  amount BIGINT NOT NULL,
  status VARCHAR(32) NOT NULL,
  request_id VARCHAR(128) NOT NULL UNIQUE,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  INDEX idx_payments_order_id (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS refunds (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  refund_id VARCHAR(64) NOT NULL,
  order_id VARCHAR(64) NOT NULL,
  refund_type VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL,
  retry_count INT NOT NULL DEFAULT 0,
  next_retry_time DATETIME DEFAULT NULL,
  last_error VARCHAR(255) NOT NULL DEFAULT '',
  reason VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  UNIQUE KEY uk_refund_id (refund_id),
  INDEX idx_refunds_order_id (order_id),
  INDEX idx_refunds_status (status),
  INDEX idx_refunds_retry (next_retry_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS activity_coupons (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  coupon_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  amount DECIMAL(10,2) NOT NULL,
  status VARCHAR(32) NOT NULL,
  issued_at DATETIME NOT NULL,
  used_at DATETIME DEFAULT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  UNIQUE KEY uk_coupon_id (coupon_id),
  INDEX idx_coupons_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS activity_seckill (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  sku_id VARCHAR(64) NOT NULL,
  stock INT NOT NULL,
  reserved INT NOT NULL DEFAULT 0,
  status VARCHAR(32) NOT NULL,
  start_time DATETIME NOT NULL,
  end_time DATETIME NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  UNIQUE KEY uk_seckill_sku (sku_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS seckill_orders (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  sku_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  quantity INT NOT NULL DEFAULT 1,
  status VARCHAR(32) NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  UNIQUE KEY uk_seckill_order (sku_id, user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS price_history (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  sku_id VARCHAR(64) NOT NULL,
  old_price DECIMAL(10,2) NOT NULL,
  new_price DECIMAL(10,2) NOT NULL,
  reason VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL,
  INDEX idx_price_sku (sku_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO inventory(sku_id, available, reserved, updated_at)
VALUES
('SKU-1001', 100, 0, NOW()),
('SKU-1002', 50, 0, NOW())
ON DUPLICATE KEY UPDATE
available = VALUES(available),
reserved = VALUES(reserved),
updated_at = VALUES(updated_at);
