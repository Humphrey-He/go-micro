-- State Transition Audit Log Schema
-- Records all state transitions for auditing and debugging

CREATE TABLE IF NOT EXISTS state_audit_log (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    entity VARCHAR(50) NOT NULL COMMENT 'Entity type: order, payment, inventory, reservation, refund, task',
    entity_id VARCHAR(100) NOT NULL COMMENT 'Entity ID',
    from_state VARCHAR(50) NOT NULL COMMENT 'Previous state',
    to_state VARCHAR(50) NOT NULL COMMENT 'New state',
    event VARCHAR(50) NOT NULL COMMENT 'Triggering event',
    operator VARCHAR(100) COMMENT 'Operator who triggered the transition',
    reason VARCHAR(500) COMMENT 'Reason for the transition',
    request_id VARCHAR(100) COMMENT 'Request ID for tracing',
    metadata JSON COMMENT 'Additional metadata',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_entity (entity),
    INDEX idx_entity_id (entity_id),
    INDEX idx_request_id (request_id),
    INDEX idx_created_at (created_at),
    INDEX idx_entity_state (entity, to_state)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='State transition audit log';
