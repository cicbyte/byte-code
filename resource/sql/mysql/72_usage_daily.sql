-- 同 sqlite/72：使用分析 P2——CLI 版本列 + 日汇总物化表
ALTER TABLE `usage_events` ADD COLUMN `client_version` TEXT NOT NULL DEFAULT ('');

CREATE TABLE IF NOT EXISTS `usage_daily` (
    `id`            INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `day`           VARCHAR(10) NOT NULL,
    `client`        VARCHAR(8) NOT NULL,
    `calls`         INT NOT NULL DEFAULT 0,
    `err_calls`     INT NOT NULL DEFAULT 0,
    `active_actors` INT NOT NULL DEFAULT 0,
    UNIQUE KEY `uk_usage_daily_day_client` (`day`, `client`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
