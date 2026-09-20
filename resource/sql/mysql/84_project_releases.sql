-- 同 sqlite/84：项目发布（Releases）+ 独立发布文件表
CREATE TABLE IF NOT EXISTS `project_releases` (
    `id`         INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `project_id` INT NOT NULL,
    `version`    VARCHAR(64) NOT NULL,
    `title`      VARCHAR(191) NOT NULL DEFAULT '',
    `notes`      TEXT NOT NULL,
    `channel`    VARCHAR(16) NOT NULL DEFAULT 'stable' CHECK(`channel` IN ('stable','beta','nightly')),
    `created_by` INT NOT NULL DEFAULT 0,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY `uk_releases_project_version` (`project_id`, `version`),
    KEY `idx_releases_project` (`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `release_files` (
    `id`               INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `release_id`       INT NOT NULL,
    `file_name`        VARCHAR(255) NOT NULL,
    `file_size`        BIGINT NOT NULL DEFAULT 0,
    `mime_type`        VARCHAR(128) NOT NULL DEFAULT '',
    `storage_key`      VARCHAR(512) NOT NULL,
    `uploader_id`      INT NOT NULL DEFAULT 0,
    `download_count`   INT NOT NULL DEFAULT 0,
    `share_token`      VARCHAR(64) NULL,
    `share_expires_at` VARCHAR(19) NOT NULL DEFAULT '',
    `created_at`       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY `uk_release_files_name` (`release_id`, `file_name`),
    KEY `idx_release_files_release` (`release_id`),
    UNIQUE KEY `idx_release_files_token` (`share_token`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
