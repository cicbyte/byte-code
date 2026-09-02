-- 项目文档索引表（记忆中枢）：磁盘 vault 的纯缓存，无真相责任，TRUNCATE 后重扫即恢复
-- path 为 vault 内相对路径（/ 分隔）；tags/linked 为逗号分隔缓存，真相在文件 frontmatter
CREATE TABLE IF NOT EXISTS `project_document_index` (
    `project_id` INTEGER NOT NULL,
    `path`       TEXT NOT NULL,
    `space`      TEXT DEFAULT '',
    `title`      TEXT DEFAULT '',
    `type`       TEXT DEFAULT '',
    `status`     TEXT DEFAULT 'published',
    `tags`       TEXT DEFAULT '',
    `linked`     TEXT DEFAULT '',
    `ext`        TEXT DEFAULT '',
    `size`       INTEGER DEFAULT 0,
    `checksum`   TEXT DEFAULT '',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`project_id`, `path`)
);

CREATE INDEX IF NOT EXISTS `idx_doc_index_tags` ON `project_document_index` (`project_id`, `space`);
