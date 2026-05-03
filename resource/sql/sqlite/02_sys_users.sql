-- 用户表
CREATE TABLE IF NOT EXISTS `sys_users` (
    `id`         INTEGER PRIMARY KEY AUTOINCREMENT,
    `username`   TEXT    NOT NULL UNIQUE,
    `password`   TEXT    NOT NULL,
    `real_name`  TEXT    NOT NULL DEFAULT '',
    `avatar`     TEXT    DEFAULT '',
    `email`      TEXT    DEFAULT '',
    `phone`      TEXT    DEFAULT '',
    `desc`       TEXT    DEFAULT '',
    `address`    TEXT    DEFAULT '',
    `status`     INTEGER NOT NULL DEFAULT 1,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS `idx_sys_users_username` ON `sys_users` (`username`);
CREATE INDEX IF NOT EXISTS `idx_sys_users_status` ON `sys_users` (`status`);

CREATE TRIGGER IF NOT EXISTS `tr_sys_users_updated_at`
AFTER UPDATE ON `sys_users`
FOR EACH ROW
BEGIN
    UPDATE `sys_users` SET `updated_at` = CURRENT_TIMESTAMP WHERE `id` = NEW.`id`;
END;
