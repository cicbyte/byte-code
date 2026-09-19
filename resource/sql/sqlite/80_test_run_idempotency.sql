-- 迁移 80：test_runs 上报幂等键（#527，反馈 #16 T2-2）
-- 可空 + 唯一索引：未携带键的行存 NULL（SQLite/MySQL 唯一索引均允许多个 NULL，
-- 空串列则会互相冲突——不可用 DEFAULT ''）
ALTER TABLE `test_runs` ADD COLUMN `idempotency_key` TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS `idx_test_runs_idem` ON `test_runs` (`project_id`, `idempotency_key`);
