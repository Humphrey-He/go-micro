-- 演示栈首启初始化：按序执行全部迁移 + 渠道映射种子
-- （由 mysql 镜像的 docker-entrypoint-initdb.d 机制执行；挂载见 docker-compose.demo.yml）
CREATE DATABASE IF NOT EXISTS go_micro CHARACTER SET utf8mb4;
USE go_micro;

SOURCE /docker-entrypoint-initdb.d/migrations/000001_init_schema.up.sql;
SOURCE /docker-entrypoint-initdb.d/migrations/000002_state_audit.up.sql;
SOURCE /docker-entrypoint-initdb.d/migrations/000003_stock_journal_reconcile.up.sql;
SOURCE /docker-entrypoint-initdb.d/migrations/000004_consumed_messages.up.sql;
SOURCE /docker-entrypoint-initdb.d/migrations/000005_outbox_lease.up.sql;
SOURCE /docker-entrypoint-initdb.d/migrations/000006_channel_distribution.up.sql;
SOURCE /docker-entrypoint-initdb.d/seed.sql;
