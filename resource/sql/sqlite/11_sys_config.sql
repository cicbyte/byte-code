-- 系统配置表
CREATE TABLE IF NOT EXISTS `sys_config` (
  `key`   VARCHAR(100) NOT NULL PRIMARY KEY,
  `value` TEXT NOT NULL DEFAULT ''
);

-- 默认系统配置
INSERT INTO `sys_config` (`key`, `value`) VALUES
('site_name', 'Byte Admin'),
('site_icp', ''),
('site_phone', ''),
('site_address', ''),
('login_captcha', '0'),
('site_open', '1'),
('site_close_text', '网站维护中，暂时无法访问！'),
('smtp_host', ''),
('smtp_port', '465'),
('smtp_user', ''),
('smtp_pass', ''),
('smtp_from', '');
