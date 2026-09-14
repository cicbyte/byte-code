-- 同 sqlite/66：使用埋点事件表（CLI/平台行为分析地基）
CREATE TABLE IF NOT EXISTS `usage_events` (
    `id`          INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `actor_id`    INT NOT NULL DEFAULT 0,
    `actor_type`  VARCHAR(8) NOT NULL DEFAULT 'human',
    `client`      VARCHAR(8) NOT NULL DEFAULT 'web',
    `method`      VARCHAR(8) NOT NULL DEFAULT '',
    `endpoint`    VARCHAR(191) NOT NULL DEFAULT '',
    `status_code` INT NOT NULL DEFAULT 0,
    `duration_ms` INT NOT NULL DEFAULT 0,
    `project_id`  INT NOT NULL DEFAULT 0,
    `session_id`  VARCHAR(64) NOT NULL DEFAULT '',
    `error_code`  INT NOT NULL DEFAULT 0,
    `params`      TEXT,
    `created_at`  DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_usage_events_endpoint (`endpoint`, `created_at`),
    INDEX idx_usage_events_actor (`actor_id`, `created_at`),
    INDEX idx_usage_events_created (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
