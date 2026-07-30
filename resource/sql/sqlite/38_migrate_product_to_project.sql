-- B 类彻底化：需求/里程碑从产品级迁移到项目级
-- 原 requirements.product_id / milestones.product_id 改为 project_id
--
-- SQLite 不支持直接 DROP 带外键的列，用重建表方式：
--   1. 旧表改名
--   2. 新建结构正确的表（project_id 替代 product_id，去掉对 products 的外键）
--   3. 从旧表复制数据，project_id 通过 projects.product_id 反查（取 MIN(id)）
--   4. DROP 旧表
--   5. 重建索引和 trigger
--
-- 注意：事务由迁移执行器（dbinit）统一包裹，本文件不再包含
-- BEGIN/COMMIT/PRAGMA（PRAGMA 在事务内是空操作，写在这里只会造成误导）

-- ============ requirements ============

ALTER TABLE `requirements` RENAME TO `requirements_old`;

CREATE TABLE IF NOT EXISTS `requirements` (
    `id`                  INTEGER PRIMARY KEY AUTOINCREMENT,
    `project_id`          INTEGER NOT NULL DEFAULT 0,
    `parent_id`           INTEGER DEFAULT 0,
    `type`                TEXT NOT NULL DEFAULT 'story' CHECK(`type` IN ('epic', 'story', 'task')),
    `title`               TEXT NOT NULL,
    `description`         TEXT DEFAULT '',
    `status`              TEXT NOT NULL DEFAULT 'draft' CHECK(`status` IN ('draft', 'confirmed', 'decomposed', 'implemented', 'closed')),
    `priority`            INTEGER NOT NULL DEFAULT 3,
    `assignee_id`         INTEGER DEFAULT 0,
    `creator_id`          INTEGER NOT NULL DEFAULT 0,
    `milestone_id`        INTEGER DEFAULT 0,
    `sort_order`          INTEGER NOT NULL DEFAULT 0,
    `acceptance_criteria` TEXT DEFAULT '',
    `source`              TEXT NOT NULL DEFAULT 'human' CHECK(`source` IN ('human', 'ai_decomposed')),
    `created_at`          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at`          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`project_id`) REFERENCES `projects`(`id`),
    FOREIGN KEY (`parent_id`) REFERENCES `requirements`(`id`),
    FOREIGN KEY (`assignee_id`) REFERENCES `sys_users`(`id`),
    FOREIGN KEY (`creator_id`) REFERENCES `sys_users`(`id`)
);

CREATE INDEX IF NOT EXISTS `idx_requirements_project` ON `requirements` (`project_id`);
CREATE INDEX IF NOT EXISTS `idx_requirements_parent` ON `requirements` (`parent_id`);
CREATE INDEX IF NOT EXISTS `idx_requirements_status` ON `requirements` (`status`);
CREATE INDEX IF NOT EXISTS `idx_requirements_type` ON `requirements` (`type`);

INSERT INTO `requirements` (
    `id`, `project_id`, `parent_id`, `type`, `title`, `description`, `status`, `priority`,
    `assignee_id`, `creator_id`, `milestone_id`, `sort_order`, `acceptance_criteria`,
    `source`, `created_at`, `updated_at`
)
SELECT
    `id`,
    COALESCE((SELECT MIN(p.`id`) FROM `projects` p WHERE p.`product_id` = `requirements_old`.`product_id`), 0),
    `parent_id`, `type`, `title`, `description`, `status`, `priority`,
    `assignee_id`, `creator_id`, `milestone_id`, `sort_order`, `acceptance_criteria`,
    `source`, `created_at`, `updated_at`
FROM `requirements_old`;

DROP TABLE `requirements_old`;

CREATE TRIGGER IF NOT EXISTS `tr_requirements_updated_at`
AFTER UPDATE ON `requirements`
FOR EACH ROW
BEGIN
    UPDATE `requirements` SET `updated_at` = CURRENT_TIMESTAMP WHERE `id` = NEW.`id`;
END;

-- ============ milestones ============

ALTER TABLE `milestones` RENAME TO `milestones_old`;

CREATE TABLE IF NOT EXISTS `milestones` (
    `id`          INTEGER PRIMARY KEY AUTOINCREMENT,
    `project_id`  INTEGER NOT NULL DEFAULT 0,
    `name`        TEXT NOT NULL,
    `description` TEXT DEFAULT '',
    `target_date` TEXT DEFAULT '',
    `status`      TEXT NOT NULL DEFAULT 'planning' CHECK(`status` IN ('planning', 'in_progress', 'released')),
    `created_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`project_id`) REFERENCES `projects`(`id`)
);

CREATE INDEX IF NOT EXISTS `idx_milestones_project` ON `milestones` (`project_id`);

INSERT INTO `milestones` (`id`, `project_id`, `name`, `description`, `target_date`, `status`, `created_at`)
SELECT
    `id`,
    COALESCE((SELECT MIN(p.`id`) FROM `projects` p WHERE p.`product_id` = `milestones_old`.`product_id`), 0),
    `name`, `description`, `target_date`, `status`, `created_at`
FROM `milestones_old`;

DROP TABLE `milestones_old`;

-- 注：projects.product_id 字段保留，前端/后端不再使用，留作回滚备份。
