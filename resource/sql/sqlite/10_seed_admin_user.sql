-- 默认管理员账号
-- 用户名: admin  密码: admin123
-- bcrypt hash of "admin123" with cost 10
INSERT INTO `sys_users` (`id`, `username`, `password`, `real_name`, `avatar`, `email`, `phone`, `desc`, `status`) VALUES
(1, 'admin', '$2a$10$CaObBvqb0g4xPcIvNMOYX.RVgZVRGWgaTxlqKmwIosvsGtAyrhq8q', 'Admin', '', 'admin@byte-code.com', '', 'manager', 1);
