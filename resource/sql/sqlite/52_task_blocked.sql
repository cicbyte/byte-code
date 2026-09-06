-- 迁移 52：任务状态新增 blocked（阻塞）
-- 语义：assignee 执行中遇到外部阻碍（等信息补充/环境问题/依赖未就绪）
-- 主动上报；阻塞任务豁免租约回收（等人不超时），信息补齐后回 in_progress。
-- SQLite 改 CHECK 需重建表：新表（以当前最终 schema 为准）→ 拷数据 → 替换 → 重建索引与触发器。
-- 注意：FK 对 "requirements_old" 的引用是 40 号迁移前的历史残留，原样保留（未开 PRAGMA foreign_keys，无实害）。

CREATE TABLE `tasks_new` (
    `id`                   INTEGER PRIMARY KEY AUTOINCREMENT,
    `project_id`           INTEGER NOT NULL,
    `requirement_id`       INTEGER DEFAULT 0,
    `sprint_id`            INTEGER DEFAULT 0,
    `title`                TEXT NOT NULL,
    `description`          TEXT DEFAULT '',
    `type`                 TEXT NOT NULL DEFAULT 'feature' CHECK(`type` IN ('feature', 'bug', 'chore', 'test')),
    `status`               TEXT NOT NULL DEFAULT 'open' CHECK(`status` IN ('open', 'in_progress', 'blocked', 'review', 'done', 'closed')),
    `priority`             INTEGER NOT NULL DEFAULT 3,
    `assignee_id`          INTEGER DEFAULT 0,
    `creator_id`           INTEGER NOT NULL DEFAULT 0,
    `parent_task_id`       INTEGER DEFAULT 0,
    `artifacts`            TEXT DEFAULT '',
    `requires_human_review` INTEGER NOT NULL DEFAULT 0,
    `human_review_status`  TEXT NOT NULL DEFAULT 'pending' CHECK(`human_review_status` IN ('pending', 'approved', 'rejected')),
    `sort_order`           INTEGER NOT NULL DEFAULT 0,
    `source`               TEXT NOT NULL DEFAULT 'human' CHECK(`source` IN ('human', 'ai', 'pm_import')),
    `created_at`           TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at`           TIMESTAMP DEFAULT CURRENT_TIMESTAMP, `completed_at` TEXT NOT NULL DEFAULT '', `ai_attempts` INTEGER NOT NULL DEFAULT 0, `ai_next_attempt_at` TEXT NOT NULL DEFAULT '', `due_date` TEXT NULL,
    FOREIGN KEY (`project_id`) REFERENCES `projects`(`id`),
    FOREIGN KEY (`requirement_id`) REFERENCES "requirements_old"(`id`),
    FOREIGN KEY (`assignee_id`) REFERENCES `sys_users`(`id`),
    FOREIGN KEY (`creator_id`) REFERENCES `sys_users`(`id`)
);

INSERT INTO `tasks_new` SELECT * FROM `tasks`;

DROP TRIGGER IF EXISTS `tr_tasks_updated_at`;
DROP TABLE `tasks`;
ALTER TABLE `tasks_new` RENAME TO `tasks`;

CREATE INDEX `idx_tasks_project` ON `tasks` (`project_id`);
CREATE INDEX `idx_tasks_status` ON `tasks` (`status`);
CREATE INDEX `idx_tasks_assignee` ON `tasks` (`assignee_id`);
CREATE INDEX `idx_tasks_sprint` ON `tasks` (`sprint_id`);
CREATE INDEX `idx_tasks_requirement` ON `tasks` (`requirement_id`);
CREATE INDEX `idx_tasks_parent` ON `tasks` (`parent_task_id`);
CREATE INDEX `idx_tasks_completed_at` ON `tasks` (`completed_at`);
CREATE INDEX `idx_tasks_ai_retry` ON `tasks` (`status`, `assignee_id`, `ai_next_attempt_at`);
CREATE INDEX `idx_tasks_due` ON `tasks` (`due_date`);

CREATE TRIGGER `tr_tasks_updated_at`
AFTER UPDATE ON `tasks`
FOR EACH ROW
BEGIN
    UPDATE `tasks` SET `updated_at` = CURRENT_TIMESTAMP WHERE `id` = NEW.`id`;
END;
