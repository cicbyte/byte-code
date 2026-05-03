-- 角色种子数据
INSERT INTO `sys_roles` (`id`, `name`, `explain`, `is_default`, `status`) VALUES
(1, '超级管理员', '拥有全部权限', 1, 'normal'),
(2, '普通用户',   '基础访问权限', 0, 'normal');

-- 角色-菜单关联（超级管理员拥有所有菜单）
INSERT INTO `sys_role_menus` (`role_id`, `menu_id`)
SELECT 1, id FROM `sys_menus`;

-- 用户-角色关联（admin 用户 -> 超级管理员）
INSERT INTO `sys_user_roles` (`user_id`, `role_id`) VALUES (1, 1);
