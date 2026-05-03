-- JWT Token 存储表
CREATE TABLE IF NOT EXISTS `sys_tokens` (
    `id`         INTEGER PRIMARY KEY AUTOINCREMENT,
    `user_id`    INTEGER NOT NULL,
    `token`      TEXT    NOT NULL UNIQUE,
    `expired_at` TIMESTAMP NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS `idx_sys_tokens_token` ON `sys_tokens` (`token`);
CREATE INDEX IF NOT EXISTS `idx_sys_tokens_user_id` ON `sys_tokens` (`user_id`);
