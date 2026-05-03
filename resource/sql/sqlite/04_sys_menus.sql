-- 菜单表
CREATE TABLE IF NOT EXISTS `sys_menus` (
    `id`         INTEGER PRIMARY KEY AUTOINCREMENT,
    `parent_id`  INTEGER NOT NULL DEFAULT 0,
    `name`       TEXT    NOT NULL,
    `path`       TEXT    DEFAULT '',
    `component`  TEXT    DEFAULT '',
    `redirect`   TEXT    DEFAULT '',
    `title`      TEXT    NOT NULL,
    `icon`       TEXT    DEFAULT '',
    `sort`       INTEGER NOT NULL DEFAULT 0,
    `status`     INTEGER NOT NULL DEFAULT 1,
    `hidden`     INTEGER NOT NULL DEFAULT 0,
    `type`       INTEGER NOT NULL DEFAULT 1,
    `auth`       TEXT    DEFAULT '',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS `idx_sys_menus_parent_id` ON `sys_menus` (`parent_id`);
CREATE INDEX IF NOT EXISTS `idx_sys_menus_status` ON `sys_menus` (`status`);

CREATE TRIGGER IF NOT EXISTS `tr_sys_menus_updated_at`
AFTER UPDATE ON `sys_menus`
FOR EACH ROW
BEGIN
    UPDATE `sys_menus` SET `updated_at` = CURRENT_TIMESTAMP WHERE `id` = NEW.`id`;
END;
