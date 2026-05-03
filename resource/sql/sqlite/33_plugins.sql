-- 插件表
CREATE TABLE IF NOT EXISTS `plugins` (
    `id`          INTEGER PRIMARY KEY AUTOINCREMENT,
    `name`        TEXT NOT NULL UNIQUE,
    `version`     TEXT NOT NULL DEFAULT '1.0.0',
    `description` TEXT DEFAULT '',
    `author`      TEXT DEFAULT '',
    `enabled`     INTEGER NOT NULL DEFAULT 1,
    `config`      TEXT DEFAULT '{}',
    `installed_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
