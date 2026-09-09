-- 同 sqlite/65：项目分组（多对多）
CREATE TABLE IF NOT EXISTS `project_groups` (
    `id`          INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `name`        VARCHAR(191) NOT NULL UNIQUE,
    `description` TEXT NOT NULL DEFAULT (''),
    `created_by`  INT NOT NULL DEFAULT 0,
    `created_at`  DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `project_group_members` (
    `id`         INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `group_id`   INT NOT NULL,
    `project_id` INT NOT NULL,
    `added_by`   INT NOT NULL DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(`group_id`, `project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX `idx_pgm_group` ON `project_group_members` (`group_id`);
CREATE INDEX `idx_pgm_project` ON `project_group_members` (`project_id`);
