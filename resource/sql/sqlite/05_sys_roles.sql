-- 角色表
CREATE TABLE IF NOT EXISTS `sys_roles` (
    `id`          INTEGER PRIMARY KEY AUTOINCREMENT,
    `name`        TEXT    NOT NULL,
    `explain`     TEXT    DEFAULT '',
    `is_default`  INTEGER NOT NULL DEFAULT 0,
    `status`      TEXT    NOT NULL DEFAULT 'normal',
    `created_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER IF NOT EXISTS `tr_sys_roles_updated_at`
AFTER UPDATE ON `sys_roles`
FOR EACH ROW
BEGIN
    UPDATE `sys_roles` SET `updated_at` = CURRENT_TIMESTAMP WHERE `id` = NEW.`id`;
END;
