ALTER TABLE order_outbox
    DROP INDEX idx_status_lease,
    DROP COLUMN owner,
    DROP COLUMN lease_until;
