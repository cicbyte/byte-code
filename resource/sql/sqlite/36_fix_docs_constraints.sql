-- 修复 docs 表的类型和状态约束，与 API 定义对齐
-- SQLite 不支持 ALTER CHECK，需要重建表

-- 先删除依赖 docs 表的触发器
DROP TRIGGER IF EXISTS `tr_docs_fts_insert`;
DROP TRIGGER IF EXISTS `tr_docs_fts_update`;
DROP TRIGGER IF EXISTS `tr_docs_fts_delete`;
DROP TRIGGER IF EXISTS `tr_docs_updated_at`;

-- 删除 FTS 虚拟表（依赖 docs 表的 content 同步）
DROP TABLE IF EXISTS `docs_fts`;

-- 备份数据
CREATE TABLE IF NOT EXISTS `docs_backup` AS SELECT * FROM `docs`;

-- 删除旧表
DROP TABLE IF EXISTS `docs`;

-- 重建 docs 表（放宽约束）
CREATE TABLE IF NOT EXISTS `docs` (
    `id`            INTEGER PRIMARY KEY AUTOINCREMENT,
    `project_id`    INTEGER DEFAULT 0,
    `parent_id`     INTEGER DEFAULT 0,
    `title`         TEXT NOT NULL,
    `content`       TEXT DEFAULT '',
    `type`          TEXT NOT NULL DEFAULT 'doc',
    `sort_order`    INTEGER NOT NULL DEFAULT 0,
    `creator_id`    INTEGER NOT NULL DEFAULT 0,
    `last_editor_id` INTEGER DEFAULT 0,
    `version`       INTEGER NOT NULL DEFAULT 1,
    `status`        TEXT NOT NULL DEFAULT 'active',
    `source`        TEXT NOT NULL DEFAULT 'human',
    `created_at`    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at`    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`project_id`) REFERENCES `projects`(`id`),
    FOREIGN KEY (`creator_id`) REFERENCES `sys_users`(`id`)
);

-- 恢复数据
INSERT INTO `docs` SELECT * FROM `docs_backup`;
DROP TABLE IF EXISTS `docs_backup`;

-- 重建索引
CREATE INDEX IF NOT EXISTS `idx_docs_project` ON `docs` (`project_id`);
CREATE INDEX IF NOT EXISTS `idx_docs_parent` ON `docs` (`parent_id`);

-- 重建 FTS
CREATE VIRTUAL TABLE IF NOT EXISTS `docs_fts` USING fts5(
    `title`,
    `content`,
    content='docs',
    content_rowid='id'
);

-- 重建触发器
CREATE TRIGGER IF NOT EXISTS `tr_docs_fts_insert`
AFTER INSERT ON `docs`
FOR EACH ROW
BEGIN
    INSERT INTO `docs_fts` (`rowid`, `title`, `content`) VALUES (NEW.`id`, NEW.`title`, NEW.`content`);
END;

CREATE TRIGGER IF NOT EXISTS `tr_docs_fts_update`
AFTER UPDATE ON `docs`
FOR EACH ROW
BEGIN
    INSERT INTO `docs_fts` (`docs_fts`, `rowid`, `title`, `content`) VALUES ('delete', OLD.`id`, OLD.`title`, OLD.`content`);
    INSERT INTO `docs_fts` (`rowid`, `title`, `content`) VALUES (NEW.`id`, NEW.`title`, NEW.`content`);
END;

CREATE TRIGGER IF NOT EXISTS `tr_docs_fts_delete`
AFTER DELETE ON `docs`
FOR EACH ROW
BEGIN
    INSERT INTO `docs_fts` (`docs_fts`, `rowid`, `title`, `content`) VALUES ('delete', OLD.`id`, OLD.`title`, OLD.`content`);
END;

CREATE TRIGGER IF NOT EXISTS `tr_docs_updated_at`
AFTER UPDATE ON `docs`
FOR EACH ROW
BEGIN
    UPDATE `docs` SET `updated_at` = CURRENT_TIMESTAMP WHERE `id` = NEW.`id`;
END;

-- 修复 doc_relations 的 target_type CHECK（与 API 对齐）
CREATE TABLE IF NOT EXISTS `doc_relations_backup` AS SELECT * FROM `doc_relations`;
DROP TABLE IF EXISTS `doc_relations`;
CREATE TABLE IF NOT EXISTS `doc_relations` (
    `id`          INTEGER PRIMARY KEY AUTOINCREMENT,
    `doc_id`      INTEGER NOT NULL,
    `target_type` TEXT NOT NULL,
    `target_id`   INTEGER NOT NULL,
    FOREIGN KEY (`doc_id`) REFERENCES `docs`(`id`),
    UNIQUE(`doc_id`, `target_type`, `target_id`)
);
INSERT INTO `doc_relations` SELECT * FROM `doc_relations_backup`;
DROP TABLE IF EXISTS `doc_relations_backup`;
CREATE INDEX IF NOT EXISTS `idx_doc_relations_target` ON `doc_relations` (`target_type`, `target_id`);
