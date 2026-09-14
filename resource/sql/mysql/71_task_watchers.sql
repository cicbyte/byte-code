-- 同 sqlite/71：任务关注者（watcher 订阅）
CREATE TABLE IF NOT EXISTS `task_watchers` (
    `id`         INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `task_id`    INT NOT NULL,
    `user_id`    INT NOT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY `uk_task_watchers_pair` (`task_id`, `user_id`),
    KEY `idx_task_watchers_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
