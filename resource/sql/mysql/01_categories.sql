-- GoFrame 项目 MySQL 初始化脚本
-- 数据库: app

-- 创建分类表
CREATE TABLE IF NOT EXISTS `categories` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `name` VARCHAR(255) NOT NULL COMMENT '分类名称，唯一',
    `description` TEXT NOT NULL COMMENT '分类描述',
    `icon` VARCHAR(512) DEFAULT NULL COMMENT '分类图标URL或标识',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序，数字越大越靠前',
    `created_at` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_name` (`name`),
    KEY `idx_sort` (`sort`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分类表';

-- 插入示例数据
INSERT INTO `categories` (`name`, `description`, `icon`, `sort`) VALUES
('技术', '技术相关文章分类', 'tech', 100),
('生活', '生活随笔分类', 'life', 90),
('学习', '学习笔记分类', 'study', 80);