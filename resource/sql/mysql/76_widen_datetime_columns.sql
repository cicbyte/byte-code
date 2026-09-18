-- 同 sqlite/76（SQLite 侧 TEXT 不受限，注释占位）：放宽 #74/#75 引入的
-- VARCHAR(19) 时间列——gdb 驱动把 gtime 格式化带微秒（26 字符）时报
-- Error 1406 Data too long（生产实测）。代码侧已改写定长字符串，此处
-- 再放宽列宽作双保险，兼容任何带微秒的存量写入尝试
ALTER TABLE `project_transfers` MODIFY COLUMN `resolved_at` VARCHAR(32) NOT NULL DEFAULT '';
ALTER TABLE `global_memory_proposals` MODIFY COLUMN `reviewed_at` VARCHAR(32) NOT NULL DEFAULT '';
