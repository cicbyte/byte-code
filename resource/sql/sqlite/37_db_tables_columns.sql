-- 数据库表定义
CREATE TABLE IF NOT EXISTS `db_tables` (
    `id`          INTEGER PRIMARY KEY AUTOINCREMENT,
    `project_id`  INTEGER NOT NULL,
    `name`        TEXT NOT NULL,
    `comment`     TEXT DEFAULT '',
    `engine`      TEXT DEFAULT '',
    `charset`     TEXT DEFAULT '',
    `sort_order`  INTEGER DEFAULT 0,
    `created_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`project_id`) REFERENCES `projects`(`id`),
    UNIQUE(`project_id`, `name`)
);

CREATE INDEX IF NOT EXISTS `idx_db_tables_project` ON `db_tables` (`project_id`);

CREATE TRIGGER IF NOT EXISTS `tr_db_tables_updated_at`
AFTER UPDATE ON `db_tables`
FOR EACH ROW
BEGIN
    UPDATE `db_tables` SET `updated_at` = CURRENT_TIMESTAMP WHERE `id` = NEW.`id`;
END;

-- 数据库列定义
CREATE TABLE IF NOT EXISTS `db_columns` (
    `id`                INTEGER PRIMARY KEY AUTOINCREMENT,
    `table_id`          INTEGER NOT NULL,
    `name`              TEXT NOT NULL,
    `type`              TEXT NOT NULL,
    `nullable`          INTEGER DEFAULT 1,
    `default_value`     TEXT DEFAULT '',
    `is_primary_key`    INTEGER DEFAULT 0,
    `is_auto_increment` INTEGER DEFAULT 0,
    `comment`           TEXT DEFAULT '',
    `sort_order`        INTEGER DEFAULT 0,
    `created_at`        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at`        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`table_id`) REFERENCES `db_tables`(`id`) ON DELETE CASCADE,
    UNIQUE(`table_id`, `name`)
);

CREATE INDEX IF NOT EXISTS `idx_db_columns_table` ON `db_columns` (`table_id`);

CREATE TRIGGER IF NOT EXISTS `tr_db_columns_updated_at`
AFTER UPDATE ON `db_columns`
FOR EACH ROW
BEGIN
    UPDATE `db_columns` SET `updated_at` = CURRENT_TIMESTAMP WHERE `id` = NEW.`id`;
END;
