-- 任务表
CREATE TABLE IF NOT EXISTS `tasks` (
    `id`                   INTEGER PRIMARY KEY AUTOINCREMENT,
    `project_id`           INTEGER NOT NULL,
    `requirement_id`       INTEGER DEFAULT 0,
    `sprint_id`            INTEGER DEFAULT 0,
    `title`                TEXT NOT NULL,
    `description`          TEXT DEFAULT '',
    `type`                 TEXT NOT NULL DEFAULT 'feature' CHECK(`type` IN ('feature', 'bug', 'chore', 'test')),
    `status`               TEXT NOT NULL DEFAULT 'open' CHECK(`status` IN ('open', 'in_progress', 'review', 'done', 'closed')),
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
    `updated_at`           TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`project_id`) REFERENCES `projects`(`id`),
    FOREIGN KEY (`requirement_id`) REFERENCES `requirements`(`id`),
    FOREIGN KEY (`assignee_id`) REFERENCES `sys_users`(`id`),
    FOREIGN KEY (`creator_id`) REFERENCES `sys_users`(`id`)
);

CREATE INDEX IF NOT EXISTS `idx_tasks_project` ON `tasks` (`project_id`);
CREATE INDEX IF NOT EXISTS `idx_tasks_status` ON `tasks` (`status`);
CREATE INDEX IF NOT EXISTS `idx_tasks_assignee` ON `tasks` (`assignee_id`);
CREATE INDEX IF NOT EXISTS `idx_tasks_sprint` ON `tasks` (`sprint_id`);
CREATE INDEX IF NOT EXISTS `idx_tasks_requirement` ON `tasks` (`requirement_id`);

-- 评论表
CREATE TABLE IF NOT EXISTS `comments` (
    `id`         INTEGER PRIMARY KEY AUTOINCREMENT,
    `task_id`    INTEGER NOT NULL,
    `user_id`    INTEGER NOT NULL,
    `content`    TEXT NOT NULL,
    `user_type`  TEXT NOT NULL DEFAULT 'human' CHECK(`user_type` IN ('human', 'ai')),
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`task_id`) REFERENCES `tasks`(`id`),
    FOREIGN KEY (`user_id`) REFERENCES `sys_users`(`id`)
);

CREATE INDEX IF NOT EXISTS `idx_comments_task` ON `comments` (`task_id`);

-- AI 执行日志表
CREATE TABLE IF NOT EXISTS `ai_execution_logs` (
    `id`         INTEGER PRIMARY KEY AUTOINCREMENT,
    `task_id`    INTEGER NOT NULL,
    `ai_user_id` INTEGER NOT NULL,
    `action`     TEXT NOT NULL,
    `detail`     TEXT DEFAULT '',
    `status`     TEXT NOT NULL DEFAULT 'success' CHECK(`status` IN ('success', 'failed')),
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`task_id`) REFERENCES `tasks`(`id`),
    FOREIGN KEY (`ai_user_id`) REFERENCES `sys_users`(`id`)
);

CREATE INDEX IF NOT EXISTS `idx_ai_logs_task` ON `ai_execution_logs` (`task_id`);

CREATE TRIGGER IF NOT EXISTS `tr_tasks_updated_at`
AFTER UPDATE ON `tasks`
FOR EACH ROW
BEGIN
    UPDATE `tasks` SET `updated_at` = CURRENT_TIMESTAMP WHERE `id` = NEW.`id`;
END;
