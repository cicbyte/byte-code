-- 项目表
CREATE TABLE IF NOT EXISTS `projects` (
    `id`          INTEGER PRIMARY KEY AUTOINCREMENT,
    `product_id`  INTEGER DEFAULT 0,
    `name`        TEXT NOT NULL,
    `description` TEXT DEFAULT '',
    `created_by`  INTEGER NOT NULL DEFAULT 0,
    `status`      INTEGER NOT NULL DEFAULT 1,
    `created_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`product_id`) REFERENCES `products`(`id`),
    FOREIGN KEY (`created_by`) REFERENCES `sys_users`(`id`)
);

CREATE INDEX IF NOT EXISTS `idx_projects_status` ON `projects` (`status`);

-- 项目成员表
CREATE TABLE IF NOT EXISTS `project_members` (
    `id`         INTEGER PRIMARY KEY AUTOINCREMENT,
    `project_id` INTEGER NOT NULL,
    `user_id`    INTEGER NOT NULL,
    `role`       TEXT NOT NULL DEFAULT 'member',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`project_id`) REFERENCES `projects`(`id`),
    FOREIGN KEY (`user_id`) REFERENCES `sys_users`(`id`),
    UNIQUE(`project_id`, `user_id`)
);

CREATE INDEX IF NOT EXISTS `idx_project_members_project` ON `project_members` (`project_id`);
CREATE INDEX IF NOT EXISTS `idx_project_members_user` ON `project_members` (`user_id`);

CREATE TRIGGER IF NOT EXISTS `tr_projects_updated_at`
AFTER UPDATE ON `projects`
FOR EACH ROW
BEGIN
    UPDATE `projects` SET `updated_at` = CURRENT_TIMESTAMP WHERE `id` = NEW.`id`;
END;
