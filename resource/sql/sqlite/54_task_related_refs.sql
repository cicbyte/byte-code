-- 迁移 54：任务跨项目引用（第一档：引用 + 被引方通知）
-- related_refs JSON 数组：[{"type":"task","projectId":2,"id":41,"title":"快照标题"}]
-- 引用方存快照 title（通知/活动流为快照语义，被引方改名不回填）。
-- 写入时 logic 层校验：结构合法 + 引用者对被引项目有读权限（CanAccessProject，
-- 防 P0-4 同款跨项目探测）。

ALTER TABLE `tasks` ADD COLUMN `related_refs` TEXT NOT NULL DEFAULT '[]';
