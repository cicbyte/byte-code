-- 同 sqlite/87：agent 自助注册来源 IP 留痕（审计）
ALTER TABLE `sys_users` ADD COLUMN `register_ip` VARCHAR(64) NOT NULL DEFAULT '';
