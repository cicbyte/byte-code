-- 首次登录强制修改密码标记：默认口令未改过的账号必须先改密才能使用系统
ALTER TABLE `sys_users` ADD COLUMN `must_change_password` INTEGER NOT NULL DEFAULT 0;

-- 仅密码仍为默认值 admin123（对应 bcrypt 哈希）的账号需要强制改密，已改过的不受影响
UPDATE `sys_users` SET `must_change_password` = 1
WHERE `password` = '$2a$10$CaObBvqb0g4xPcIvNMOYX.RVgZVRGWgaTxlqKmwIosvsGtAyrhq8q';
