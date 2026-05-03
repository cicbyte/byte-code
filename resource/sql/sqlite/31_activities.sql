-- 活动流表
CREATE TABLE IF NOT EXISTS `activities` (
    `id`          INTEGER PRIMARY KEY AUTOINCREMENT,
    `actor_id`    INTEGER NOT NULL DEFAULT 0,
    `actor_type`  TEXT NOT NULL DEFAULT 'human' CHECK(`actor_type` IN ('human', 'ai', 'system')),
    `actor_name`  TEXT DEFAULT '',
    `action`      TEXT NOT NULL DEFAULT '',
    `target_type` TEXT NOT NULL DEFAULT '',
    `target_id`   INTEGER NOT NULL DEFAULT 0,
    `target_name` TEXT DEFAULT '',
    `project_id`  INTEGER DEFAULT 0,
    `detail`      TEXT DEFAULT '',
    `created_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS `idx_activities_created` ON `activities` (`created_at`);
CREATE INDEX IF NOT EXISTS `idx_activities_project` ON `activities` (`project_id`);
CREATE INDEX IF NOT EXISTS `idx_activities_actor` ON `activities` (`actor_id`);
