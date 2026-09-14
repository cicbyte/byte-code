-- 同 sqlite/70：忘记密码重置令牌（只存 SHA-256 摘要，30min 单次）
CREATE TABLE IF NOT EXISTS `password_resets` (
    `id`         INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `user_id`    INT NOT NULL,
    `token_hash` VARCHAR(64) NOT NULL,
    `expires_at` DATETIME NOT NULL,
    `used`       TINYINT NOT NULL DEFAULT 0,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE INDEX `idx_pwd_resets_hash` ON `password_resets` (`token_hash`);
CREATE INDEX `idx_pwd_resets_user` ON `password_resets` (`user_id`);
