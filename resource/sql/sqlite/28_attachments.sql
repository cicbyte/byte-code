-- 附件表
CREATE TABLE IF NOT EXISTS `attachments` (
    `id`             INTEGER PRIMARY KEY AUTOINCREMENT,
    `s3_key`         TEXT NOT NULL UNIQUE,
    `original_name`  TEXT NOT NULL,
    `file_size`      INTEGER NOT NULL DEFAULT 0,
    `mime_type`      TEXT NOT NULL DEFAULT '',
    `file_ext`       TEXT DEFAULT '',
    `entity_type`    TEXT NOT NULL CHECK(`entity_type` IN ('task', 'doc', 'test_case', 'requirement', 'comment')),
    `entity_id`      INTEGER NOT NULL DEFAULT 0,
    `uploader_id`    INTEGER NOT NULL DEFAULT 0,
    `description`    TEXT DEFAULT '',
    `download_count` INTEGER NOT NULL DEFAULT 0,
    `created_at`     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`uploader_id`) REFERENCES `sys_users`(`id`)
);

CREATE INDEX IF NOT EXISTS `idx_attachments_entity` ON `attachments` (`entity_type`, `entity_id`);
CREATE INDEX IF NOT EXISTS `idx_attachments_s3_key` ON `attachments` (`s3_key`);
