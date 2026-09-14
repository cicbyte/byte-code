-- 迁移 71：任务关注者（watcher 订阅）
-- assignee/creator 之外的角色（如测试、协作方）可显式订阅任务动态：
-- 评论/认领/完成/阻塞/解除/重开/审核事件向 watcher 扇出通知。
-- (task_id, user_id) 唯一约束防重复关注；user_id 侧索引支撑"我关注的"反查
CREATE TABLE IF NOT EXISTS `task_watchers` (
    `id`         INTEGER PRIMARY KEY AUTOINCREMENT,
    `task_id`    INTEGER NOT NULL,
    `user_id`    INTEGER NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS `idx_task_watchers_pair` ON `task_watchers` (`task_id`, `user_id`);
CREATE INDEX IF NOT EXISTS `idx_task_watchers_user` ON `task_watchers` (`user_id`);
