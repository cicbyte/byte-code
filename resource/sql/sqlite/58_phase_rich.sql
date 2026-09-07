-- 迁移 58：阶段升级为迷你任务结构（专题自闭环，不再转出）
-- 复制任务表格的核心字段：描述（detail 已有）+ 负责人 + 产出；
-- 状态流转 pending→in_progress→done 专题内完成。
-- 转任务按钮随本迁移从前端移除（后端端点保留兼容既有血缘数据）。

ALTER TABLE `topic_phases` ADD COLUMN `assignee_id` INTEGER NOT NULL DEFAULT 0;
ALTER TABLE `topic_phases` ADD COLUMN `artifacts` TEXT NOT NULL DEFAULT '';
