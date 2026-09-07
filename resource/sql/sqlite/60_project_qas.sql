-- 迁移 60：项目 QA 库
-- 常见问答沉淀：问题 → 答案（markdown），由 agent 在解决问题的过程中
-- 沉淀与更新（"遇到的坑 + 解法"），人或 agent 遇到问题时检索。
-- hits 计数用于开工包挑选高频 QA（新会话最可能用到的先给）。

CREATE TABLE IF NOT EXISTS `project_qas` (
    `id`         INTEGER PRIMARY KEY AUTOINCREMENT,
    `project_id` INTEGER NOT NULL,
    `question`   TEXT NOT NULL,
    `answer`     TEXT DEFAULT '',
    `tags`       TEXT DEFAULT '',
    `hits`       INTEGER NOT NULL DEFAULT 0,
    `status`     TEXT NOT NULL DEFAULT 'active' CHECK(`status` IN ('active', 'archived')),
    `created_by` INTEGER NOT NULL DEFAULT 0,
    `updated_by` INTEGER NOT NULL DEFAULT 0,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS `idx_qas_project` ON `project_qas` (`project_id`, `status`);
