-- 迁移 78：test_cases.priority INT → VARCHAR(8)
-- API 层一直以 P0-P3 字符串校验/写入（SQLite 动态类型无感），MySQL INT 列在
-- 严格模式下插入 'P2' 直接 Error 1366——#505 用例同步依赖该列可写，先行归一
UPDATE `test_cases` SET `priority` = CONCAT('P', `priority`) WHERE `priority` REGEXP '^[0-3]$';
ALTER TABLE `test_cases` MODIFY COLUMN `priority` VARCHAR(8) NOT NULL DEFAULT 'P3';
