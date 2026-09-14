-- 同 sqlite/69：Agent capabilities 列（空 = 全部能力）。
-- MySQL 的 TEXT 不支持 DEFAULT 字面量，取 NULL；读取侧按空串处理。
ALTER TABLE `agent_project_bindings` ADD COLUMN `capabilities` TEXT NULL;
