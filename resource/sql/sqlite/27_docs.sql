-- 文档表
CREATE TABLE IF NOT EXISTS `docs` (
    `id`            INTEGER PRIMARY KEY AUTOINCREMENT,
    `project_id`    INTEGER DEFAULT 0,
    `parent_id`     INTEGER DEFAULT 0,
    `title`         TEXT NOT NULL,
    `content`       TEXT DEFAULT '',
    `type`          TEXT NOT NULL DEFAULT 'doc' CHECK(`type` IN ('doc', 'folder', 'template')),
    `sort_order`    INTEGER NOT NULL DEFAULT 0,
    `creator_id`    INTEGER NOT NULL DEFAULT 0,
    `last_editor_id` INTEGER DEFAULT 0,
    `version`       INTEGER NOT NULL DEFAULT 1,
    `status`        TEXT NOT NULL DEFAULT 'active' CHECK(`status` IN ('active', 'archived')),
    `source`        TEXT NOT NULL DEFAULT 'human' CHECK(`source` IN ('human', 'ai_generated', 'ai_updated')),
    `created_at`    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at`    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`project_id`) REFERENCES `projects`(`id`),
    FOREIGN KEY (`creator_id`) REFERENCES `sys_users`(`id`)
);

CREATE INDEX IF NOT EXISTS `idx_docs_project` ON `docs` (`project_id`);
CREATE INDEX IF NOT EXISTS `idx_docs_parent` ON `docs` (`parent_id`);

-- 文档关联表
CREATE TABLE IF NOT EXISTS `doc_relations` (
    `id`          INTEGER PRIMARY KEY AUTOINCREMENT,
    `doc_id`      INTEGER NOT NULL,
    `target_type` TEXT NOT NULL CHECK(`target_type` IN ('project', 'requirement', 'task', 'test_case')),
    `target_id`   INTEGER NOT NULL,
    FOREIGN KEY (`doc_id`) REFERENCES `docs`(`id`),
    UNIQUE(`doc_id`, `target_type`, `target_id`)
);

CREATE INDEX IF NOT EXISTS `idx_doc_relations_target` ON `doc_relations` (`target_type`, `target_id`);

-- 文档版本表
CREATE TABLE IF NOT EXISTS `doc_versions` (
    `id`             INTEGER PRIMARY KEY AUTOINCREMENT,
    `doc_id`         INTEGER NOT NULL,
    `version`        INTEGER NOT NULL,
    `content`        TEXT NOT NULL DEFAULT '',
    `editor_id`      INTEGER DEFAULT 0,
    `change_summary` TEXT DEFAULT '',
    `created_at`     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`doc_id`) REFERENCES `docs`(`id`)
);

CREATE INDEX IF NOT EXISTS `idx_doc_versions_doc` ON `doc_versions` (`doc_id`);

-- FTS5 全文搜索虚拟表
CREATE VIRTUAL TABLE IF NOT EXISTS `docs_fts` USING fts5(
    `title`,
    `content`,
    content='docs',
    content_rowid='id'
);

-- 触发器：自动同步 FTS
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
