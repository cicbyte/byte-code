-- 迁移 82：test_run_cases 补 code 快照列（建议）
-- 失败用例源码内联展示（附件通道保留作截图/大文件补充）；服务端截断 64KB
ALTER TABLE `test_run_cases` ADD COLUMN `code` TEXT NOT NULL DEFAULT '';
