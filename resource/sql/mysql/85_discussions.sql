-- 同 sqlite/85：讨论区（Discussions）+ 回复表
CREATE TABLE IF NOT EXISTS `discussions` (
    `id`                INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `project_id`        INT NOT NULL,
    `title`             VARCHAR(191) NOT NULL,
    `body`              TEXT NOT NULL,
    `status`            VARCHAR(16) NOT NULL DEFAULT 'open' CHECK(`status` IN ('open','converted','archived')),
    `author_id`         INT NOT NULL DEFAULT 0,
    `converted_task_id` INT NULL,
    `created_at`        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY `idx_discussions_project` (`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `discussion_replies` (
    `id`            INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `discussion_id` INT NOT NULL,
    `user_id`       INT NOT NULL,
    `content`       TEXT NOT NULL,
    `user_type`     VARCHAR(8) NOT NULL DEFAULT 'human' CHECK(`user_type` IN ('human', 'ai')),
    `created_at`    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    KEY `idx_discussion_replies_thread` (`discussion_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
