-- 迁移 74：全局记忆提案队列——人人可发 + 管理员审核（独立于 project_memories，
-- 避开其 pending「未确认推测」语义；未审内容绝不进读者视野）
CREATE TABLE IF NOT EXISTS `global_memory_proposals` (
    `id`            INTEGER PRIMARY KEY AUTOINCREMENT,
    `key`           TEXT NOT NULL,
    `value`         TEXT NOT NULL DEFAULT '',
    `ttl`           TEXT NOT NULL DEFAULT '',
    `note`          TEXT NOT NULL DEFAULT '',
    `status`        TEXT NOT NULL DEFAULT 'submitted' CHECK(`status` IN ('submitted','approved','rejected')),
    `proposed_by`   INTEGER NOT NULL DEFAULT 0,
    `reviewed_by`   INTEGER NOT NULL DEFAULT 0,
    `review_reason` TEXT NOT NULL DEFAULT '',
    `created_at`    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `reviewed_at`   TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS `idx_gmp_status` ON `global_memory_proposals` (`status`);
