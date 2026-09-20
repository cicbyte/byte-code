-- 迁移 84：项目发布（Releases，#551/#552）——本地打包上传、团队内下载安装包
-- version 项目内唯一；channel 稳定/内测/每日
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

-- 发布文件独立体系（#552：不走附件通道——安装包动辄几百 MB，且要按
-- 项目/版本组织、支持公开令牌直链分享）
-- storage_key = release/{projectId}/{version}/{文件名}，版本内同名拒传
-- share_token 非空即可公开下载（/api/release-files/public/{token}），
-- share_expires_at 空=永久；吊销=清空 token
CREATE TABLE IF NOT EXISTS `release_files` (
    `id`               INTEGER PRIMARY KEY AUTOINCREMENT,
    `release_id`       INTEGER NOT NULL,
    `file_name`        TEXT NOT NULL,
    `file_size`        INTEGER NOT NULL DEFAULT 0,
    `mime_type`        TEXT NOT NULL DEFAULT '',
    `storage_key`      TEXT NOT NULL,
    `uploader_id`      INTEGER NOT NULL DEFAULT 0,
    `download_count`   INTEGER NOT NULL DEFAULT 0,
    `share_token`      TEXT,
    `share_expires_at` TEXT NOT NULL DEFAULT '',
    `created_at`       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (`release_id`, `file_name`)
);
CREATE INDEX IF NOT EXISTS `idx_release_files_release` ON `release_files` (`release_id`);
CREATE UNIQUE INDEX IF NOT EXISTS `idx_release_files_token` ON `release_files` (`share_token`);
