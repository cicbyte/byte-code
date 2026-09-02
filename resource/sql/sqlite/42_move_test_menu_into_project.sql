-- 测试管理并入项目工作台（前端项目菜单承载），移除全局测试管理菜单入口
DELETE FROM `sys_role_menus` WHERE `menu_id` IN (40, 41, 42);
DELETE FROM `sys_menus` WHERE `id` IN (40, 41, 42);
