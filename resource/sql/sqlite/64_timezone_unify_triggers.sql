-- 迁移 64：时区统一（SQLite 侧）——updated_at 触发器 UTC → 本地时间
-- 背景：应用层一直写本地时间字符串，而 11 个触发器写 CURRENT_TIMESTAMP
-- （SQLite 恒 UTC）——双轨制历史上已引发 8 小时偏移误释放事故（duedate
-- P0-2）。统一口径：全库时间列一律本地时间（YYYY-MM-DD HH:MM:SS）。
-- 注意：created_at 的 DEFAULT CURRENT_TIMESTAMP 改写与本迁移的存量数据
-- 修正无法在此执行（迁移框架单文件事务内 PRAGMA 是空操作），由 dbinit
-- 的启动后处理（autocommit 连接）完成——见 utility/dbinit/timezone.go
DROP TRIGGER IF EXISTS `tr_categories_updated_at`;
CREATE TRIGGER `tr_categories_updated_at`
AFTER UPDATE ON `categories`
FOR EACH ROW
BEGIN
    UPDATE `categories` SET `updated_at` = datetime('now', 'localtime') WHERE `id` = NEW.`id`;
END;

DROP TRIGGER IF EXISTS `tr_projects_updated_at`;
CREATE TRIGGER `tr_projects_updated_at`
AFTER UPDATE ON `projects`
FOR EACH ROW
BEGIN
    UPDATE `projects` SET `updated_at` = datetime('now', 'localtime') WHERE `id` = NEW.`id`;
END;

DROP TRIGGER IF EXISTS `tr_requirements_updated_at`;
CREATE TRIGGER `tr_requirements_updated_at`
AFTER UPDATE ON `requirements`
FOR EACH ROW
BEGIN
    UPDATE `requirements` SET `updated_at` = datetime('now', 'localtime') WHERE `id` = NEW.`id`;
END;

DROP TRIGGER IF EXISTS `tr_sys_menus_updated_at`;
CREATE TRIGGER `tr_sys_menus_updated_at`
AFTER UPDATE ON `sys_menus`
FOR EACH ROW
BEGIN
    UPDATE `sys_menus` SET `updated_at` = datetime('now', 'localtime') WHERE `id` = NEW.`id`;
END;

DROP TRIGGER IF EXISTS `tr_sys_roles_updated_at`;
CREATE TRIGGER `tr_sys_roles_updated_at`
AFTER UPDATE ON `sys_roles`
FOR EACH ROW
BEGIN
    UPDATE `sys_roles` SET `updated_at` = datetime('now', 'localtime') WHERE `id` = NEW.`id`;
END;

DROP TRIGGER IF EXISTS `tr_sys_users_updated_at`;
CREATE TRIGGER `tr_sys_users_updated_at`
AFTER UPDATE ON `sys_users`
FOR EACH ROW
BEGIN
    UPDATE `sys_users` SET `updated_at` = datetime('now', 'localtime') WHERE `id` = NEW.`id`;
END;

DROP TRIGGER IF EXISTS `tr_tasks_updated_at`;
CREATE TRIGGER `tr_tasks_updated_at`
AFTER UPDATE ON `tasks`
FOR EACH ROW
BEGIN
    UPDATE `tasks` SET `updated_at` = datetime('now', 'localtime') WHERE `id` = NEW.`id`;
END;

DROP TRIGGER IF EXISTS `tr_test_cases_updated_at`;
CREATE TRIGGER `tr_test_cases_updated_at`
AFTER UPDATE ON `test_cases`
FOR EACH ROW
BEGIN
    UPDATE `test_cases` SET `updated_at` = datetime('now', 'localtime') WHERE `id` = NEW.`id`;
END;

-- 遗留表（数据库模型功能已下线，表仍在）：一并统一，防将来复活踩坑
DROP TRIGGER IF EXISTS `tr_db_tables_updated_at`;
CREATE TRIGGER `tr_db_tables_updated_at`
AFTER UPDATE ON `db_tables`
FOR EACH ROW
BEGIN
    UPDATE `db_tables` SET `updated_at` = datetime('now', 'localtime') WHERE `id` = NEW.`id`;
END;

DROP TRIGGER IF EXISTS `tr_db_columns_updated_at`;
CREATE TRIGGER `tr_db_columns_updated_at`
AFTER UPDATE ON `db_columns`
FOR EACH ROW
BEGIN
    UPDATE `db_columns` SET `updated_at` = datetime('now', 'localtime') WHERE `id` = NEW.`id`;
END;

DROP TRIGGER IF EXISTS `tr_project_databases_updated_at`;
CREATE TRIGGER `tr_project_databases_updated_at`
AFTER UPDATE ON `project_databases`
FOR EACH ROW
BEGIN
    UPDATE `project_databases` SET `updated_at` = datetime('now', 'localtime') WHERE `id` = NEW.`id`;
END;
