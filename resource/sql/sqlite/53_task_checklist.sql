-- 迁移 53：任务步骤清单（checklist）
-- 长任务工作流：任务执行步骤的结构化清单（[{"text":"...","done":false}]），
-- 对应 long-running 工作流的 steps 概念——agent 按步执行、逐项打勾；
-- 打勾走 UpdateTask 触发 updated_at 变化，天然构成租约心跳（一石二鸟）。

ALTER TABLE `tasks` ADD COLUMN `checklist` TEXT NOT NULL DEFAULT '[]';
