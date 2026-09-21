-- 同 sqlite/86：工作日志（Worklogs）+ 来源标记
CREATE TABLE IF NOT EXISTS `worklogs` (
    `id`         INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `project_id` INT NOT NULL,
    `author_id`  INT NOT NULL DEFAULT 0,
    `content`    TEXT NOT NULL,
    `source`     VARCHAR(16) NOT NULL DEFAULT 'manual' CHECK(`source` IN ('manual','tasks')),
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY `idx_worklogs_project` (`project_id`),
    KEY `idx_worklogs_created` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
