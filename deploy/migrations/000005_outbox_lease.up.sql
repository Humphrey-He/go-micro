-- R3: Outbox 投递深度——认领/租约/owner 三件套
ALTER TABLE order_outbox
    ADD COLUMN owner VARCHAR(64) NOT NULL DEFAULT '' COMMENT '当前认领实例',
    ADD COLUMN lease_until DATETIME(3) NULL COMMENT '认领租约到期时间',
    ADD INDEX idx_status_lease (status, lease_until);
