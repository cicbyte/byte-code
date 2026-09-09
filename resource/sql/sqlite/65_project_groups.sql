-- 迁移 65：项目分组（多对多）
-- 语义：同 group 的项目自动互为关联（反馈投递/跨项目引用免准入），
-- 替代手动 project_relations 的主要场景。一个项目可属多个 group。
CREATE TABLE IF NOT EXISTS `project_groups` (
    `id`          INTEGER PRIMARY KEY AUTOINCREMENT,
    `name`        TEXT    NOT NULL UNIQUE,
    `description` TEXT    NOT NULL DEFAULT '',
    `created_by`  INTEGER NOT NULL DEFAULT 0,
    `created_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS `project_group_members` (
    `id`         INTEGER PRIMARY KEY AUTOINCREMENT,
    `group_id`   INTEGER NOT NULL,
    `project_id` INTEGER NOT NULL,
    `added_by`   INTEGER NOT NULL DEFAULT 0,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(`group_id`, `project_id`)
);

CREATE INDEX IF NOT EXISTS `idx_pgm_group` ON `project_group_members` (`group_id`);
CREATE INDEX IF NOT EXISTS `idx_pgm_project` ON `project_group_members` (`project_id`);
