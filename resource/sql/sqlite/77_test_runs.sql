-- 迁移 77：测试执行记录（pytest 深度支持 P1，#504）
-- test_runs = 一次批量上报（pytest session / CI job / 手工批次）
-- test_run_cases = 逐用例结果；external_key 存 pytest nodeid，test_case_id 可空（未映射）
-- test_cases 补 external_key 列支撑 --bcode-sync 按 nodeid 幂等自动建用例
-- （非唯一索引 + 业务层查重：空串会重复，唯一约束不可用）
CREATE TABLE IF NOT EXISTS `test_runs` (
    `id`           INTEGER PRIMARY KEY AUTOINCREMENT,
    `project_id`   INTEGER NOT NULL,
    `source`       TEXT NOT NULL DEFAULT 'pytest' CHECK(`source` IN ('manual','pytest','ci','junit')),
    `branch`       TEXT NOT NULL DEFAULT '',
    `git_sha`      TEXT NOT NULL DEFAULT '',
    `env`          TEXT NOT NULL DEFAULT '',
    `triggered_by` INTEGER NOT NULL DEFAULT 0,
    `total`        INTEGER NOT NULL DEFAULT 0,
    `passed`       INTEGER NOT NULL DEFAULT 0,
    `failed`       INTEGER NOT NULL DEFAULT 0,
    `skipped`      INTEGER NOT NULL DEFAULT 0,
    `errors`       INTEGER NOT NULL DEFAULT 0,
    `duration_ms`  INTEGER NOT NULL DEFAULT 0,
    `started_at`   TEXT NOT NULL DEFAULT '',
    `finished_at`  TEXT NOT NULL DEFAULT '',
    `created_at`   TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS `idx_test_runs_project` ON `test_runs` (`project_id`);
CREATE INDEX IF NOT EXISTS `idx_test_runs_created` ON `test_runs` (`created_at`);

CREATE TABLE IF NOT EXISTS `test_run_cases` (
    `id`           INTEGER PRIMARY KEY AUTOINCREMENT,
    `test_run_id`  INTEGER NOT NULL,
    `test_case_id` INTEGER NOT NULL DEFAULT 0,
    `external_key` TEXT NOT NULL DEFAULT '',
    `title`        TEXT NOT NULL DEFAULT '',
    `status`       TEXT NOT NULL DEFAULT 'pass' CHECK(`status` IN ('pass','fail','error','skip')),
    `duration_ms`  INTEGER NOT NULL DEFAULT 0,
    `message`      TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS `idx_trc_run` ON `test_run_cases` (`test_run_id`);
CREATE INDEX IF NOT EXISTS `idx_trc_case` ON `test_run_cases` (`test_case_id`);
CREATE INDEX IF NOT EXISTS `idx_trc_ext` ON `test_run_cases` (`external_key`);

ALTER TABLE `test_cases` ADD COLUMN `external_key` TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS `idx_test_cases_ext` ON `test_cases` (`project_id`, `external_key`);
