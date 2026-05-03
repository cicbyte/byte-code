-- 通知表
CREATE TABLE IF NOT EXISTS `notifications` (
    `id`         INTEGER PRIMARY KEY AUTOINCREMENT,
    `user_id`    INTEGER NOT NULL DEFAULT 0,
    `title`      TEXT NOT NULL DEFAULT '',
    `content`    TEXT DEFAULT '',
    `type`       TEXT NOT NULL DEFAULT 'info' CHECK(`type` IN ('info', 'warning', 'success', 'error')),
    `is_read`    INTEGER NOT NULL DEFAULT 0,
    `source_type` TEXT DEFAULT '',
    `source_id`  INTEGER DEFAULT 0,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`user_id`) REFERENCES `sys_users`(`id`)
);

CREATE INDEX IF NOT EXISTS `idx_notifications_user` ON `notifications` (`user_id`);
CREATE INDEX IF NOT EXISTS `idx_notifications_read` ON `notifications` (`user_id`, `is_read`);
