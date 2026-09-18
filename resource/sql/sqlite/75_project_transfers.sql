-- 迁移 75：项目移交邀请队列表（邀请制转交：owner 发起，目标收通知接受/拒绝）
CREATE TABLE IF NOT EXISTS `project_transfers` (
    `id`             INTEGER PRIMARY KEY AUTOINCREMENT,
    `project_id`     INTEGER NOT NULL,
    `from_user_id`   INTEGER NOT NULL,
    `to_user_id`     INTEGER NOT NULL,
    `leave_project`  INTEGER NOT NULL DEFAULT 0,
    `status`         TEXT NOT NULL DEFAULT 'pending' CHECK(`status` IN ('pending','accepted','declined','cancelled')),
    `created_at`     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `resolved_at`    TEXT NOT NULL DEFAULT '',
    `resolved_by`    INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS `idx_pt_project_status` ON `project_transfers` (`project_id`, `status`);
CREATE INDEX IF NOT EXISTS `idx_pt_to_user_status` ON `project_transfers` (`to_user_id`, `status`);
