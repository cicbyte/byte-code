-- 项目数据库配置表
CREATE TABLE IF NOT EXISTS `project_databases` (
    `id`                INTEGER PRIMARY KEY AUTOINCREMENT,
    `project_id`        INTEGER NOT NULL UNIQUE,
    `db_type`           TEXT NOT NULL DEFAULT 'none' CHECK(`db_type` IN ('none', 'sqlite', 'mysql', 'postgresql')),
    `db_name`           TEXT DEFAULT '',
    `db_host`           TEXT DEFAULT '',
    `db_port`           INTEGER DEFAULT 0,
    `db_user`           TEXT DEFAULT '',
    `db_password`       TEXT DEFAULT '',
    `db_options`        TEXT DEFAULT '{}',
    `connection_status` TEXT NOT NULL DEFAULT 'disconnected' CHECK(`connection_status` IN ('disconnected', 'connected', 'error')),
    `last_tested_at`    TEXT DEFAULT '',
    `created_at`        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at`        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`project_id`) REFERENCES `projects`(`id`)
);

CREATE TRIGGER IF NOT EXISTS `tr_project_databases_updated_at`
AFTER UPDATE ON `project_databases`
FOR EACH ROW
BEGIN
    UPDATE `project_databases` SET `updated_at` = CURRENT_TIMESTAMP WHERE `id` = NEW.`id`;
END;

-- Schema 版本历史表
CREATE TABLE IF NOT EXISTS `schema_versions` (
    `id`                INTEGER PRIMARY KEY AUTOINCREMENT,
    `project_id`        INTEGER NOT NULL,
    `version`           INTEGER NOT NULL,
    `change_type`       TEXT NOT NULL CHECK(`change_type` IN ('create_table', 'alter_table', 'drop_table', 'rename')),
    `change_description` TEXT DEFAULT '',
    `table_name`        TEXT DEFAULT '',
    `sql_statement`     TEXT NOT NULL DEFAULT '',
    `schema_before`     TEXT DEFAULT '',
    `schema_after`      TEXT DEFAULT '',
    `operator_id`       INTEGER DEFAULT 0,
    `task_id`           INTEGER DEFAULT 0,
    `status`            TEXT NOT NULL DEFAULT 'applied' CHECK(`status` IN ('applied', 'rolled_back')),
    `created_at`        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`project_id`) REFERENCES `projects`(`id`),
    FOREIGN KEY (`operator_id`) REFERENCES `sys_users`(`id`)
);

CREATE INDEX IF NOT EXISTS `idx_schema_versions_project` ON `schema_versions` (`project_id`, `version`);
