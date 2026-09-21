-- 迁移 86：工作日志（Worklogs）——项目演化叙事，日期分组时间轴呈现；
-- 成员与 agent（worklog 能力位）撰写，可从已完成任务/发布半自动生成草稿
-- （source=manual 手写 | tasks 草稿生成后发布）
CREATE TABLE IF NOT EXISTS `worklogs` (
    `id`         INTEGER PRIMARY KEY AUTOINCREMENT,
    `project_id` INTEGER NOT NULL,
    `author_id`  INTEGER NOT NULL DEFAULT 0,
    `content`    TEXT NOT NULL,
    `source`     TEXT NOT NULL DEFAULT 'manual' CHECK(`source` IN ('manual','tasks')),
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS `idx_worklogs_project` ON `worklogs` (`project_id`);
CREATE INDEX IF NOT EXISTS `idx_worklogs_created` ON `worklogs` (`created_at`);
