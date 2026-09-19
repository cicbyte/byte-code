-- 迁移 81：attachments.entity_type CHECK 扩 test_run_case（执行用例失败截图，#527）
-- 与 Go 层白名单对齐补 project；SQLite 改 CHECK 须整表重建（列定义不变仅约束扩）
CREATE TABLE `attachments_rebuild81` (
    `id`             INTEGER PRIMARY KEY AUTOINCREMENT,
    `s3_key`         TEXT NOT NULL,
    `original_name`  TEXT NOT NULL,
    `file_size`      INTEGER NOT NULL DEFAULT 0,
    `mime_type`      TEXT NOT NULL DEFAULT '',
    `file_ext`       TEXT NOT NULL DEFAULT '',
    `entity_type`    TEXT NOT NULL CHECK(`entity_type` IN ('task', 'doc', 'test_case', 'test_run_case', 'requirement', 'comment', 'project')),
    `entity_id`      INTEGER NOT NULL DEFAULT 0,
    `entity_key`     TEXT NOT NULL DEFAULT '',
    `uploader_id`    INTEGER NOT NULL DEFAULT 0,
    `description`    TEXT NOT NULL DEFAULT '',
    `download_count` INTEGER NOT NULL DEFAULT 0,
    `created_at`     TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO `attachments_rebuild81` SELECT `id`,`s3_key`,`original_name`,`file_size`,`mime_type`,`file_ext`,`entity_type`,`entity_id`,`entity_key`,`uploader_id`,`description`,`download_count`,`created_at` FROM `attachments`;
DROP TABLE `attachments`;
ALTER TABLE `attachments_rebuild81` RENAME TO `attachments`;
CREATE INDEX IF NOT EXISTS `idx_attachments_entity` ON `attachments` (`entity_type`, `entity_id`);
CREATE INDEX IF NOT EXISTS `idx_attachments_s3_key` ON `attachments` (`s3_key`);
CREATE INDEX IF NOT EXISTS `idx_attachments_entity_key` ON `attachments` (`entity_type`, `entity_key`);
