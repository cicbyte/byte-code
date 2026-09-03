-- AI 任务执行退避：失败计数与下次可尝试时间。
-- 此前失败任务直接回退 open，2 分钟后无限重试（持续烧 API 费用），
-- 且 oldest-first 排序下 3 个毒任务会饿死全部新任务
ALTER TABLE `tasks` ADD COLUMN `ai_attempts` INTEGER NOT NULL DEFAULT 0;
ALTER TABLE `tasks` ADD COLUMN `ai_next_attempt_at` TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS `idx_tasks_ai_retry` ON `tasks` (`status`, `assignee_id`, `ai_next_attempt_at`);
