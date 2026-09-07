-- Agent 账号菜单归属调整：独立顶级组「AI管理」并入「平台」组
-- （Agent 管理为平台级功能，与通知/活动/审计同域；接口本身在管理组）
DELETE FROM `sys_role_menus` WHERE `menu_id` = 60;
DELETE FROM `sys_menus` WHERE `id` = 60;
UPDATE `sys_menus`
SET `parent_id` = 80, `title` = 'Agent 账号', `path` = 'agents', `component` = '/platform/agents', `sort` = 70
WHERE `id` = 61;

-- 超级管理员补齐全量菜单：09 号种子执行时仅存在 1-12/70-72 号菜单，
-- 其后 35 号等新增的菜单从未关联到该角色（前端侧靠 system_menu/system_role
-- 超管直通掩盖了该缺口，但角色授权树与按角色分配权限会缺项）
INSERT OR IGNORE INTO `sys_role_menus` (`role_id`, `menu_id`)
SELECT 1, `id` FROM `sys_menus`;
