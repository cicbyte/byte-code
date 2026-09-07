-- 迁移 55：项目关联（项目级协作）
-- 语义：A 添加 B 为关联项目后，A 的新建任务入口可直接选 B 为目标项目
-- （任务实际落在 B，B 的准入 agent 通过既有任务池/广播直接认领，零协议改动）。
-- 单向关注起步（A 关联 B 不代表 B 关联 A）；添加时校验操作者对 B 可访问。

CREATE TABLE IF NOT EXISTS `project_relations` (
    `id`                 INTEGER PRIMARY KEY AUTOINCREMENT,
    `project_id`         INTEGER NOT NULL,
    `related_project_id` INTEGER NOT NULL,
    `created_by`         INTEGER NOT NULL DEFAULT 0,
    `created_at`         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(`project_id`, `related_project_id`)
);
CREATE INDEX IF NOT EXISTS `idx_relations_project` ON `project_relations` (`project_id`);
