-- 菜单种子数据（精简版：仅保留核心管理页面）
INSERT OR IGNORE INTO `sys_menus` (`id`, `parent_id`, `name`, `path`, `component`, `redirect`, `title`, `icon`, `sort`, `status`, `hidden`, `type`, `auth`) VALUES
-- Dashboard
(1,  0, 'Dashboard',           '/dashboard', 'LAYOUT', '/dashboard/console', 'Dashboard', 'DashboardOutlined', 100, 1, 0, 1, ''),
(2,  1, 'dashboard_console',   'console',    '/dashboard/console/console', '', '仪表盘', '', 100, 1, 0, 2, 'dashboard_console'),
-- 系统管理
(10, 0, 'System',              '/system',    'LAYOUT', '/system/menu',       '系统管理', 'SettingOutlined', 90, 1, 0, 1, ''),
(11, 10, 'system_menu',        'menu',       '/system/menu/menu',           '', '菜单管理', '', 100, 1, 0, 2, 'system_menu'),
(12, 10, 'system_role',        'role',       '/system/role/role',           '', '角色管理', '', 90, 1, 0, 2, 'system_role'),
-- 设置
(70, 0, 'Setting',             '/setting',   'LAYOUT', '/setting/account',  '设置', 'SettingOutlined', 30, 1, 0, 1, ''),
(71, 70, 'setting_account',    'account',    '/setting/account/account',    '', '个人设置', '', 100, 1, 0, 2, ''),
(72, 70, 'setting_system',     'system',     '/setting/system/system',      '', '系统设置', '', 90, 1, 0, 2, '');
