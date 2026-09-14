-- 同 sqlite/67：使用分析菜单项（管理员）
INSERT IGNORE INTO `sys_menus` (`id`, `parent_id`, `name`, `path`, `component`, `redirect`, `title`, `icon`, `sort`, `status`, `hidden`, `type`, `auth`) VALUES
(84, 80, 'platform_usage', 'usage', '/platform/usage', '', '使用分析', '', 70, 1, 0, 2, '');

INSERT IGNORE INTO `sys_role_menus` (`role_id`, `menu_id`) VALUES (1, 84);
