-- 任务截止日期（到期提醒依据；空=无截止）
ALTER TABLE `tasks` ADD COLUMN `due_date` TEXT NULL;
CREATE INDEX IF NOT EXISTS `idx_tasks_due` ON `tasks` (`due_date`);
