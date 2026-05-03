-- Sprint 迭代表
CREATE TABLE IF NOT EXISTS `sprints` (
    `id`         INTEGER PRIMARY KEY AUTOINCREMENT,
    `project_id` INTEGER NOT NULL,
    `name`       TEXT NOT NULL,
    `goal`       TEXT DEFAULT '',
    `start_date` TEXT NOT NULL,
    `end_date`   TEXT NOT NULL,
    `status`     TEXT NOT NULL DEFAULT 'planning' CHECK(`status` IN ('planning', 'active', 'completed')),
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`project_id`) REFERENCES `projects`(`id`)
);

CREATE INDEX IF NOT EXISTS `idx_sprints_project` ON `sprints` (`project_id`);
CREATE INDEX IF NOT EXISTS `idx_sprints_status` ON `sprints` (`status`);
