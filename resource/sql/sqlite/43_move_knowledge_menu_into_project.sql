-- 知识库并入项目工作台，移除全局知识库菜单入口
DELETE FROM `sys_role_menus` WHERE `menu_id` IN (50, 51);
DELETE FROM `sys_menus` WHERE `id` IN (50, 51);
