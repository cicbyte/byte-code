-- 评论系统增强：支持回复（parent_id）与编辑（updated_at）
ALTER TABLE `comments` ADD COLUMN `parent_id` INTEGER NOT NULL DEFAULT 0;
ALTER TABLE `comments` ADD COLUMN `updated_at` TEXT DEFAULT '';
CREATE INDEX IF NOT EXISTS `idx_comments_parent` ON `comments` (`parent_id`);
