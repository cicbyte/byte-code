-- 迁移 68：P1 权限字典扩充（PRD design/permission-system-prd.md §3.1）
-- 标签管理、项目分组、创建项目三个权限面此前全员无门槛；P1-2 起
-- 后端按字典执行（perm.RequireMenuPerm），本迁移先落字典与角色预绑。
-- 挂载：platform_tags/platform_groups 入平台组(80)；project_create 入
-- 新根组 85「项目权限」（项目级治理权限的挂载点，后续可扩展归档等）。
INSERT IGNORE INTO `sys_menus` (`id`, `parent_id`, `name`, `path`, `component`, `redirect`, `title`, `icon`, `sort`, `status`, `hidden`, `type`, `auth`) VALUES
(85, 0, 'Project', '/project', '', '', '项目权限', '', 40, 1, 0, 1, ''),
(86, 80, 'platform_tags', 'tags', '', '', '标签管理', '', 60, 1, 0, 2, ''),
(87, 80, 'platform_groups', 'groups', '', '', '项目分组', '', 65, 1, 0, 2, ''),
(88, 85, 'project_create', '', '', '', '创建项目', '', 90, 1, 0, 2, '');

-- 超管(1)直通全部新权限项
INSERT IGNORE INTO `sys_role_menus` (`role_id`, `menu_id`)
SELECT 1, `id` FROM `sys_menus`;

-- 普通用户(2)预绑 project_create + platform_tags：维持「登录即可建项目/
-- 用标签」的现状，升级不断能力；platform_groups 不预绑（分组治理属
-- 管理侧授予，PRD §3.3）
INSERT IGNORE INTO `sys_role_menus` (`role_id`, `menu_id`)
SELECT 2, `id` FROM `sys_menus` WHERE `name` IN ('platform_tags', 'project_create');
