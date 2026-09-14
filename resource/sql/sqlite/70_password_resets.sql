-- 迁移 70：忘记密码重置令牌
-- 令牌只存 SHA-256 摘要（拖库不泄露令牌本体）；30 分钟有效、单次使用；
-- used 置位而非删行（保留发起记录可审计）
CREATE TABLE IF NOT EXISTS `password_resets` (
    `id`         INTEGER PRIMARY KEY AUTOINCREMENT,
    `user_id`    INTEGER NOT NULL,
    `token_hash` TEXT NOT NULL,
    `expires_at` TEXT NOT NULL,
    `used`       INTEGER NOT NULL DEFAULT 0,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS `idx_pwd_resets_hash` ON `password_resets` (`token_hash`);
CREATE INDEX IF NOT EXISTS `idx_pwd_resets_user` ON `password_resets` (`user_id`);
