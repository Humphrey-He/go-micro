-- 添加 phone 字段到 users 表
ALTER TABLE users ADD COLUMN phone VARCHAR(32) UNIQUE COMMENT '手机号';

-- 插入测试用户 (手机号 17673796081, 密码 260011)
-- bcrypt hash of "260011" with cost 10
INSERT INTO users (user_id, username, phone, password_hash, role, status, created_at, updated_at)
VALUES (
  'USER-17673796081',
  'user_17673796081',
  '17673796081',
  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
  'user',
  1,
  NOW(),
  NOW()
) ON DUPLICATE KEY UPDATE phone = VALUES(phone);