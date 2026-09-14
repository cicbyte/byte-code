-- 迁移 67：使用分析菜单项（管理员）
-- sys_menus.name ↔ 前端路由 menuKey 一一对应（62 迁移确立的权限字典口径）。
-- 接口在管理组（admin），菜单门槛同审计日志：仅管理员可见。
INSERT OR IGNORE INTO `sys_menus` (`id`, `parent_id`, `name`, `path`, `component`, `redirect`, `title`, `icon`, `sort`, `status`, `hidden`, `type`, `auth`) VALUES
(84, 80, 'platform_usage', 'usage', '/platform/usage', '', '使用分析', '', 70, 1, 0, 2, '');

-- 管理员(1)直通全部菜单项
INSERT OR IGNORE INTO `sys_role_menus` (`role_id`, `menu_id`) VALUES (1, 84);
