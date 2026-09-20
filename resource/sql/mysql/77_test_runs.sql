-- 同 sqlite/77：测试执行记录（pytest 深度支持 P1）
-- external_key 用 VARCHAR(512)（可索引；TEXT 建索引需前缀长度，迁移 73 同款教训）
-- errors 列名避开 MySQL 保留字 ERROR 的裸用风险
CREATE TABLE IF NOT EXISTS `test_runs` (
    `id`           INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `project_id`   INT NOT NULL,
    `source`       VARCHAR(16) NOT NULL DEFAULT 'pytest' CHECK(`source` IN ('manual','pytest','ci','junit')),
    `branch`       VARCHAR(128) NOT NULL DEFAULT '',
    `git_sha`      VARCHAR(64) NOT NULL DEFAULT '',
    `env`          VARCHAR(64) NOT NULL DEFAULT '',
    `triggered_by` INT NOT NULL DEFAULT 0,
    `total`        INT NOT NULL DEFAULT 0,
    `passed`       INT NOT NULL DEFAULT 0,
    `failed`       INT NOT NULL DEFAULT 0,
    `skipped`      INT NOT NULL DEFAULT 0,
    `errors`       INT NOT NULL DEFAULT 0,
    `duration_ms`  INT NOT NULL DEFAULT 0,
    `started_at`   VARCHAR(19) NOT NULL DEFAULT '',
    `finished_at`  VARCHAR(19) NOT NULL DEFAULT '',
    `created_at`   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    KEY `idx_test_runs_project` (`project_id`),
    KEY `idx_test_runs_created` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `test_run_cases` (
    `id`           INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `test_run_id`  INT NOT NULL,
    `test_case_id` INT NOT NULL DEFAULT 0,
    `external_key` VARCHAR(512) NOT NULL DEFAULT '',
    `title`        VARCHAR(512) NOT NULL DEFAULT '',
    `status`       VARCHAR(8) NOT NULL DEFAULT 'pass' CHECK(`status` IN ('pass','fail','error','skip')),
    `duration_ms`  INT NOT NULL DEFAULT 0,
    `message`      TEXT NOT NULL,
    KEY `idx_trc_run` (`test_run_id`),
    KEY `idx_trc_case` (`test_case_id`),
    KEY `idx_trc_ext` (`external_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE `test_cases` ADD COLUMN `external_key` VARCHAR(512) NOT NULL DEFAULT '', ADD INDEX `idx_test_cases_ext` (`project_id`, `external_key`);
