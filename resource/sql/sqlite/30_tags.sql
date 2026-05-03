-- 标签表
CREATE TABLE IF NOT EXISTS `tags` (
    `id`         INTEGER PRIMARY KEY AUTOINCREMENT,
    `name`       TEXT NOT NULL UNIQUE,
    `color`      TEXT NOT NULL DEFAULT '#1890ff',
    `creator_id` INTEGER NOT NULL DEFAULT 0,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`creator_id`) REFERENCES `sys_users`(`id`)
);

-- 实体标签关联表
CREATE TABLE IF NOT EXISTS `entity_tags` (
    `id`          INTEGER PRIMARY KEY AUTOINCREMENT,
    `tag_id`      INTEGER NOT NULL,
    `entity_type` TEXT NOT NULL CHECK(`entity_type` IN ('task', 'requirement', 'test_case')),
    `entity_id`   INTEGER NOT NULL,
    FOREIGN KEY (`tag_id`) REFERENCES `tags`(`id`),
    UNIQUE(`tag_id`, `entity_type`, `entity_id`)
);

CREATE INDEX IF NOT EXISTS `idx_entity_tags_tag` ON `entity_tags` (`tag_id`);
CREATE INDEX IF NOT EXISTS `idx_entity_tags_entity` ON `entity_tags` (`entity_type`, `entity_id`);
