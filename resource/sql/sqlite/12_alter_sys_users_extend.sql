-- 扩展 sys_users 表：支持 AI 用户
ALTER TABLE `sys_users` ADD COLUMN `type` TEXT NOT NULL DEFAULT 'human' CHECK(`type` IN ('human', 'ai'));
ALTER TABLE `sys_users` ADD COLUMN `capabilities` TEXT DEFAULT '';
ALTER TABLE `sys_users` ADD COLUMN `api_key` TEXT DEFAULT '';
ALTER TABLE `sys_users` ADD COLUMN `api_key_salt` TEXT DEFAULT '';
ALTER TABLE `sys_users` ADD COLUMN `owner_human_id` INTEGER DEFAULT 0;

CREATE INDEX IF NOT EXISTS `idx_sys_users_type` ON `sys_users` (`type`);
CREATE INDEX IF NOT EXISTS `idx_sys_users_api_key` ON `sys_users` (`api_key`);
