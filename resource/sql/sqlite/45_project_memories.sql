-- 项目 KV 记忆表（记忆中枢）：agent 自治记忆，状态机 pending/active/stale/expired
-- 0=全局记忆；key 点分层级（build.cmd / conventions.naming）；正文不设 JSON 约束，格式由 agent 自管
CREATE TABLE IF NOT EXISTS `project_memories` (
    `id`               INTEGER PRIMARY KEY AUTOINCREMENT,
    `project_id`       INTEGER NOT NULL DEFAULT 0,
    `key`              TEXT NOT NULL,
    `value`            TEXT NOT NULL DEFAULT '',
    `status`           TEXT NOT NULL DEFAULT 'active' CHECK(`status` IN ('pending','active','stale','expired')),
    `expires_at`       TIMESTAMP NULL,
    `last_verified_at` TIMESTAMP NULL,
    `verified_by`      INTEGER DEFAULT 0,
    `updated_by`       INTEGER NOT NULL DEFAULT 0,
    `created_at`       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at`       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(`project_id`, `key`)
);

CREATE INDEX IF NOT EXISTS `idx_memories_status` ON `project_memories` (`project_id`, `status`);
