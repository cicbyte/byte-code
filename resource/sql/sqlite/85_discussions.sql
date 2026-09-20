-- 迁移 85：讨论区（Discussions）——想法/议题线程，论坛式回复，
-- 成熟后转任务（status=converted + converted_task_id 血缘回链）
CREATE TABLE IF NOT EXISTS `discussions` (
    `id`                INTEGER PRIMARY KEY AUTOINCREMENT,
    `project_id`        INTEGER NOT NULL,
    `title`             TEXT NOT NULL,
    `body`              TEXT NOT NULL DEFAULT '',
    `status`            TEXT NOT NULL DEFAULT 'open' CHECK(`status` IN ('open','converted','archived')),
    `author_id`         INTEGER NOT NULL DEFAULT 0,
    `converted_task_id` INTEGER,
    `created_at`        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at`        TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS `idx_discussions_project` ON `discussions` (`project_id`);

-- 讨论回复：结构与任务评论对齐（user_type 服务端按账号类型推导）
CREATE TABLE IF NOT EXISTS `discussion_replies` (
    `id`            INTEGER PRIMARY KEY AUTOINCREMENT,
    `discussion_id` INTEGER NOT NULL,
    `user_id`       INTEGER NOT NULL,
    `content`       TEXT NOT NULL,
    `user_type`     TEXT NOT NULL DEFAULT 'human' CHECK(`user_type` IN ('human', 'ai')),
    `created_at`    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS `idx_discussion_replies_thread` ON `discussion_replies` (`discussion_id`);
