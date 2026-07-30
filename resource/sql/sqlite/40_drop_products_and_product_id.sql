-- 产品层并入项目后的收尾清理（配合后端移除 product_id 的读写）：
-- 1) 清理 38 号迁移遗留的无项目归属孤儿需求/里程碑
-- 2) 重建 projects 表，去掉 product_id 列与对 products 的外键
--    （否则全新部署在 38 号迁移 PRAGMA 残留的连接上创建项目必违反外键）
-- 3) 删除废弃的 products 表与"产品管理"菜单
-- 注意：PRAGMA foreign_keys 在事务内无效，须置于事务外；结束时保持 off，
-- 与连接池其它连接的默认状态一致，同时中和 38 号迁移在同一条连接上的 PRAGMA 残留
PRAGMA foreign_keys = off;

BEGIN TRANSACTION;

DELETE FROM `requirements` WHERE `project_id` = 0 OR `project_id` NOT IN (SELECT `id` FROM `projects`);
DELETE FROM `milestones` WHERE `project_id` = 0 OR `project_id` NOT IN (SELECT `id` FROM `projects`);

CREATE TABLE `projects_rebuild` (
    `id`          INTEGER PRIMARY KEY AUTOINCREMENT,
    `name`        TEXT NOT NULL,
    `description` TEXT DEFAULT '',
    `created_by`  INTEGER NOT NULL DEFAULT 0,
    `status`      INTEGER NOT NULL DEFAULT 1,
    `created_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`created_by`) REFERENCES `sys_users`(`id`)
);

INSERT INTO `projects_rebuild` (`id`, `name`, `description`, `created_by`, `status`, `created_at`, `updated_at`)
SELECT `id`, `name`, `description`, `created_by`, `status`, `created_at`, `updated_at` FROM `projects`;

DROP TABLE `projects`;

ALTER TABLE `projects_rebuild` RENAME TO `projects`;

CREATE INDEX IF NOT EXISTS `idx_projects_status` ON `projects` (`status`);

CREATE TRIGGER IF NOT EXISTS `tr_projects_updated_at`
AFTER UPDATE ON `projects`
FOR EACH ROW
BEGIN
    UPDATE `projects` SET `updated_at` = CURRENT_TIMESTAMP WHERE `id` = NEW.`id`;
END;

DROP TABLE IF EXISTS `products`;

DELETE FROM `sys_role_menus` WHERE `menu_id` IN (20, 21, 22);
DELETE FROM `sys_menus` WHERE `id` IN (20, 21, 22);

COMMIT;
