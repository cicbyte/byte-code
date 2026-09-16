-- 同 sqlite/74：全局记忆提案队列（人人可发 + 管理员审核）。
-- 索引内联进建表语句（单条 DDL，规避 MySQL 多语句迁移的 DDL 原子性问题，见 README）
CREATE TABLE IF NOT EXISTS `global_memory_proposals` (
    `id`            INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `key`           VARCHAR(191) NOT NULL,
    `value`         TEXT NOT NULL DEFAULT (''),
    `ttl`           VARCHAR(8) NOT NULL DEFAULT '',
    `note`          TEXT NOT NULL DEFAULT (''),
    `status`        VARCHAR(16) NOT NULL DEFAULT 'submitted' CHECK(`status` IN ('submitted','approved','rejected')),
    `proposed_by`   INT NOT NULL DEFAULT 0,
    `reviewed_by`   INT NOT NULL DEFAULT 0,
    `review_reason` TEXT NOT NULL DEFAULT (''),
    `created_at`    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `reviewed_at`   VARCHAR(19) NOT NULL DEFAULT '',
    KEY `idx_gmp_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
