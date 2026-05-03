-- 用户-角色关联表
CREATE TABLE IF NOT EXISTS `sys_user_roles` (
    `id`      INTEGER PRIMARY KEY AUTOINCREMENT,
    `user_id` INTEGER NOT NULL,
    `role_id` INTEGER NOT NULL,
    UNIQUE(`user_id`, `role_id`)
);
