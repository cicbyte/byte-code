-- 同 sqlite/84：项目发布（Releases，#551）+ attachments 枚举扩 release
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

ALTER TABLE `attachments` MODIFY COLUMN `entity_type` VARCHAR(32) NOT NULL
    CHECK(`entity_type` IN ('task', 'doc', 'test_case', 'test_run_case', 'requirement', 'comment', 'project', 'release'));
