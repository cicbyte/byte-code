-- 审计日志表
CREATE TABLE IF NOT EXISTS `audit_logs` (
    `id`          INTEGER PRIMARY KEY AUTOINCREMENT,
    `actor_id`    INTEGER NOT NULL DEFAULT 0,
    `actor_type`  TEXT NOT NULL DEFAULT 'human' CHECK(`actor_type` IN ('human', 'ai', 'system')),
    `action`      TEXT NOT NULL DEFAULT '',
    `target_type` TEXT NOT NULL DEFAULT '',
    `target_id`   INTEGER NOT NULL DEFAULT 0,
    `target_name` TEXT DEFAULT '',
    `changes`     TEXT DEFAULT '',
    `ip_address`  TEXT DEFAULT '',
    `user_agent`  TEXT DEFAULT '',
    `project_id`  INTEGER DEFAULT 0,
    `created_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS `idx_audit_logs_target` ON `audit_logs` (`target_type`, `target_id`);
CREATE INDEX IF NOT EXISTS `idx_audit_logs_actor` ON `audit_logs` (`actor_id`);
CREATE INDEX IF NOT EXISTS `idx_audit_logs_created` ON `audit_logs` (`created_at`);
CREATE INDEX IF NOT EXISTS `idx_audit_logs_project` ON `audit_logs` (`project_id`);
