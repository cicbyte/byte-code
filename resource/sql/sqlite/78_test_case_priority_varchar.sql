-- 迁移 78：test_cases.priority 归一为 P0-P3 字符串（表重建）
-- 双重根因：①API 一直以 P0-P3 字符串校验/写入；②gdb 按列类型做值转换，
-- 'P1' 进 INTEGER 亲和列被 gconv 成 0（静默），进 MySQL INT 列严格模式报
-- Error 1366。SQLite 无法 ALTER 列型，走重建：新表 priority TEXT 亲和，
-- 存量 0-3 映射回 P 前缀。索引与 updated_at 触发器随重建恢复
CREATE TABLE `test_cases_rebuild78` (
    `id`              INTEGER PRIMARY KEY AUTOINCREMENT,
    `project_id`      INTEGER NOT NULL,
    `requirement_id`  INTEGER DEFAULT 0,
    `task_id`         INTEGER DEFAULT 0,
    `title`           TEXT NOT NULL,
    `preconditions`   TEXT DEFAULT '',
    `steps`           TEXT NOT NULL DEFAULT '',
    `expected_result` TEXT NOT NULL DEFAULT '',
    `category`        TEXT DEFAULT '',
    `module`          TEXT DEFAULT '',
    `priority`        TEXT NOT NULL DEFAULT 'P3',
    `source`          TEXT NOT NULL DEFAULT 'human' CHECK(`source` IN ('human', 'ai_generated')),
    `creator_id`      INTEGER NOT NULL DEFAULT 0,
    `status`          TEXT NOT NULL DEFAULT 'active' CHECK(`status` IN ('active', 'deprecated')),
    `external_key`    TEXT NOT NULL DEFAULT '',
    `created_at`      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at`      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO `test_cases_rebuild78`
    (`id`,`project_id`,`requirement_id`,`task_id`,`title`,`preconditions`,`steps`,`expected_result`,`category`,`module`,`priority`,`source`,`creator_id`,`status`,`external_key`,`created_at`,`updated_at`)
SELECT `id`,`project_id`,`requirement_id`,`task_id`,`title`,`preconditions`,`steps`,`expected_result`,`category`,`module`,
       CASE CAST(`priority` AS TEXT)
            WHEN '0' THEN 'P0' WHEN '1' THEN 'P1' WHEN '2' THEN 'P2' WHEN '3' THEN 'P3'
            ELSE CAST(`priority` AS TEXT) END,
       `source`,`creator_id`,`status`,`external_key`,`created_at`,`updated_at`
FROM `test_cases`;

DROP TABLE `test_cases`;
ALTER TABLE `test_cases_rebuild78` RENAME TO `test_cases`;

CREATE INDEX IF NOT EXISTS `idx_test_cases_project` ON `test_cases` (`project_id`);
CREATE INDEX IF NOT EXISTS `idx_test_cases_status` ON `test_cases` (`status`);
CREATE INDEX IF NOT EXISTS `idx_test_cases_ext` ON `test_cases` (`project_id`, `external_key`);

CREATE TRIGGER IF NOT EXISTS `tr_test_cases_updated_at`
AFTER UPDATE ON `test_cases`
FOR EACH ROW
BEGIN
    UPDATE `test_cases` SET `updated_at` = CURRENT_TIMESTAMP WHERE `id` = NEW.`id`;
END;
