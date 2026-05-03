-- 角色-菜单关联表
CREATE TABLE IF NOT EXISTS `sys_role_menus` (
    `id`      INTEGER PRIMARY KEY AUTOINCREMENT,
    `role_id` INTEGER NOT NULL,
    `menu_id` INTEGER NOT NULL,
    UNIQUE(`role_id`, `menu_id`)
);
