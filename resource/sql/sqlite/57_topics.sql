-- 迁移 57：专题（long-task）
-- 专注一个 topic 的长时间自动执行工程，独立于日常任务池。
-- 工作流：写 PRD（项目文档库）→ 创建 topic 关联文档 → 从 PRD 拆解阶段
-- → agent 循环推进（work：读 handoff→挑阶段→执行→留痕）→ 人终验收。
-- 阶段可转出日常任务（复用转任务链路，反向关联）。

CREATE TABLE IF NOT EXISTS `topics` (
    `id`           INTEGER PRIMARY KEY AUTOINCREMENT,
    `project_id`   INTEGER NOT NULL,
    `title`        TEXT NOT NULL,
    `goal`         TEXT DEFAULT '',              -- markdown：目标与范围
    `acceptance`   TEXT DEFAULT '',              -- 完成判据（创建时定义，人验收依据）
    `doc_path`     TEXT DEFAULT '',              -- 关联文档（通常是 PRD）：项目文档库内的 path
    `assignee_id`  INTEGER DEFAULT 0,            -- 执行 agent（一期单 agent）
    `status`       TEXT NOT NULL DEFAULT 'active' CHECK(`status` IN ('active', 'completed', 'abandoned')),
    `created_by`   INTEGER NOT NULL DEFAULT 0,
    `created_at`   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `completed_at` TEXT DEFAULT ''
);
CREATE INDEX IF NOT EXISTS `idx_topics_project` ON `topics` (`project_id`, `status`);

CREATE TABLE IF NOT EXISTS `topic_phases` (
    `id`         INTEGER PRIMARY KEY AUTOINCREMENT,
    `topic_id`   INTEGER NOT NULL,
    `title`      TEXT NOT NULL,
    `detail`     TEXT DEFAULT '',
    `sort_order` INTEGER NOT NULL DEFAULT 0,
    `status`     TEXT NOT NULL DEFAULT 'pending' CHECK(`status` IN ('pending', 'in_progress', 'done')),
    `task_id`    INTEGER DEFAULT 0,              -- 转出的日常任务（可空）
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS `idx_topic_phases` ON `topic_phases` (`topic_id`, `status`);

-- 专题工作日志：复用 ai_execution_logs，task_id=0 + topic_id 标识
-- （handoff 交接摘要也走这里，action=handoff）
ALTER TABLE `ai_execution_logs` ADD COLUMN `topic_id` INTEGER NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS `idx_ai_logs_topic` ON `ai_execution_logs` (`topic_id`);
