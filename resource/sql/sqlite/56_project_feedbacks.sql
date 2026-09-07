-- 迁移 56：跨项目反馈（线索投递 → 对方 agent 自主决策是否建任务）
-- A 关联 B 后，A 可向 B 投递反馈；B 的准入 agent 阅读分析后：
--   convert（建任务，回填 converted_task_id，可引用来源任务）或
--   dismiss（不建，必填理由）。
-- 反馈不进 B 的任务列表——任务池只放确认过的工作。

CREATE TABLE IF NOT EXISTS `project_feedbacks` (
    `id`                 INTEGER PRIMARY KEY AUTOINCREMENT,
    `project_id`         INTEGER NOT NULL,            -- 接收方（B）
    `source_project_id`  INTEGER NOT NULL,            -- 发起方（A）
    `source_task_id`     INTEGER DEFAULT 0,           -- 来源任务（可空，血缘可溯）
    `title`              TEXT NOT NULL,
    `content`            TEXT DEFAULT '',             -- markdown：现象/线索/怀疑点
    `status`             TEXT NOT NULL DEFAULT 'open' CHECK(`status` IN ('open', 'converted', 'dismissed')),
    `converted_task_id`  INTEGER DEFAULT 0,
    `handled_by`         INTEGER DEFAULT 0,
    `handled_at`         TEXT DEFAULT '',
    `dismiss_reason`     TEXT DEFAULT '',
    `created_by`         INTEGER NOT NULL DEFAULT 0,
    `created_at`         TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS `idx_feedbacks_project` ON `project_feedbacks` (`project_id`, `status`);
