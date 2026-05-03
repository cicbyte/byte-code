-- GoFrame 项目 SQLite 初始化脚本
-- 数据库: app.db

-- 创建分类表
CREATE TABLE IF NOT EXISTS `categories` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `name` TEXT NOT NULL UNIQUE,
    `description` TEXT NOT NULL,
    `icon` TEXT,
    `sort` INTEGER NOT NULL DEFAULT 0,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS `idx_categories_sort` ON `categories` (`sort`);
CREATE INDEX IF NOT EXISTS `idx_categories_created_at` ON `categories` (`created_at`);

-- 创建触发器：自动更新 updated_at
CREATE TRIGGER IF NOT EXISTS `tr_categories_updated_at`
AFTER UPDATE ON `categories`
FOR EACH ROW
BEGIN
    UPDATE `categories` SET `updated_at` = CURRENT_TIMESTAMP WHERE `id` = NEW.`id`;
END;

-- 插入示例数据
INSERT INTO `categories` (`name`, `description`, `icon`, `sort`) VALUES
('技术', '技术相关文章分类', 'tech', 100),
('生活', '生活随笔分类', 'life', 90),
('学习', '学习笔记分类', 'study', 80);