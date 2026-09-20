-- 迁移 84：项目发布（Releases，#551）——本地打包上传、团队内下载安装包
-- version 项目内唯一；channel 稳定/内测/每日；文件复用附件通道（entity_type=release）
CREATE TABLE IF NOT EXISTS `project_releases` (
    `id`         INTEGER PRIMARY KEY AUTOINCREMENT,
    `project_id` INTEGER NOT NULL,
    `version`    TEXT NOT NULL,
    `title`      TEXT NOT NULL DEFAULT '',
    `notes`      TEXT NOT NULL DEFAULT '',
    `channel`    TEXT NOT NULL DEFAULT 'stable' CHECK(`channel` IN ('stable','beta','nightly')),
    `created_by` INTEGER NOT NULL DEFAULT 0,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (`project_id`, `version`)
);
CREATE INDEX IF NOT EXISTS `idx_releases_project` ON `project_releases` (`project_id`);

-- attachments.entity_type CHECK 扩 release（沿用 81 机制：SQLite 改 CHECK 须整表重建）
CREATE TABLE `attachments_rebuild84` (
    `id`             INTEGER PRIMARY KEY AUTOINCREMENT,
    `s3_key`         TEXT NOT NULL,
    `original_name`  TEXT NOT NULL,
    `file_size`      INTEGER NOT NULL DEFAULT 0,
    `mime_type`      TEXT NOT NULL DEFAULT '',
    `file_ext`       TEXT NOT NULL DEFAULT '',
    `entity_type`    TEXT NOT NULL CHECK(`entity_type` IN ('task', 'doc', 'test_case', 'test_run_case', 'requirement', 'comment', 'project', 'release')),
    `entity_id`      INTEGER NOT NULL DEFAULT 0,
    `entity_key`     TEXT NOT NULL DEFAULT '',
    `uploader_id`    INTEGER NOT NULL DEFAULT 0,
    `description`    TEXT NOT NULL DEFAULT '',
    `download_count` INTEGER NOT NULL DEFAULT 0,
    `created_at`     TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO `attachments_rebuild84` SELECT `id`,`s3_key`,`original_name`,`file_size`,`mime_type`,`file_ext`,`entity_type`,`entity_id`,`entity_key`,`uploader_id`,`description`,`download_count`,`created_at` FROM `attachments`;
DROP TABLE `attachments`;
ALTER TABLE `attachments_rebuild84` RENAME TO `attachments`;
CREATE INDEX IF NOT EXISTS `idx_attachments_entity` ON `attachments` (`entity_type`, `entity_id`);
CREATE INDEX IF NOT EXISTS `idx_attachments_s3_key` ON `attachments` (`s3_key`);
CREATE INDEX IF NOT EXISTS `idx_attachments_entity_key` ON `attachments` (`entity_type`, `entity_key`);
