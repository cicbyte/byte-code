-- 同 sqlite/80：test_runs 上报幂等键（可空列；项目内唯一，NULL 互不冲突）
ALTER TABLE `test_runs` ADD COLUMN `idempotency_key` VARCHAR(128) NULL;
CREATE UNIQUE INDEX `idx_test_runs_idem` ON `test_runs` (`project_id`, `idempotency_key`);
