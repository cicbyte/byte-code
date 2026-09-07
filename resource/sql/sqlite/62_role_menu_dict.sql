-- 迁移 62：权限字典与路由表对齐（只留真实生效的门槛项）
-- 背景：B 类化/项目工作台化后，大量 sys_menus 条目对应的页面已下线，或可见性
-- 从未由 sys_menus 控制（前端路由无 menuKey，勾选不产生任何效果）。
-- 保留的真实权限面（前端 menuKey ↔ sys_menus.name 一一对应）：
--   系统管理(10)：菜单管理(11, 超管直通凭据)、角色管理(12, 权限管理页)
--   设置(70)：系统设置(72)、全局记忆(73, 新增——页面调管理组接口)
--   平台(80)：Agent 账号(61, 迁移61迁入)、审计日志(83)
-- 删除的死项：1-2 仪表盘（路由无门槛）、20-22 产品管理（模块已删）、
--   30-33 项目管理（无门槛）、40-42 测试管理（并入项目工作台）、50-51 知识库（同）、
--   71 个人设置（无门槛）、81-82 活动流/标签管理（无门槛）
DELETE FROM `sys_role_menus` WHERE `menu_id` IN (1, 2, 20, 21, 22, 30, 31, 32, 33, 40, 41, 42, 50, 51, 71, 81, 82);
DELETE FROM `sys_menus` WHERE `id` IN (1, 2, 20, 21, 22, 30, 31, 32, 33, 40, 41, 42, 50, 51, 71, 81, 82);

INSERT OR IGNORE INTO `sys_menus` (`id`, `parent_id`, `name`, `path`, `component`, `redirect`, `title`, `icon`, `sort`, `status`, `hidden`, `type`, `auth`) VALUES
(73, 70, 'setting_global_memory', 'global-memory', '/setting/global-memory', '', '全局记忆', '', 80, 1, 0, 2, '');

-- 普通用户(2)不预置任何权限项：普通用户可见的全部页面（仪表盘/项目/我的任务/
-- 通知等）本就不走权限字典；兜底集仅保留给完全没有角色绑定的账号（auth 已区分）
INSERT OR IGNORE INTO `sys_role_menus` (`role_id`, `menu_id`)
SELECT 1, `id` FROM `sys_menus`;
