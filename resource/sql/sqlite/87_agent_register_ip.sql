-- 迁移 87：agent 自助注册来源 IP 留痕（审计）——注册限流本就按 IP
-- 计数，此处落库给管理页展示；管理端创建的 agent 无 IP（空串）
ALTER TABLE `sys_users` ADD COLUMN `register_ip` TEXT NOT NULL DEFAULT '';
