-- 产品表
CREATE TABLE IF NOT EXISTS `products` (
    `id`          INTEGER PRIMARY KEY AUTOINCREMENT,
    `name`        TEXT NOT NULL,
    `description` TEXT DEFAULT '',
    `owner_id`    INTEGER NOT NULL DEFAULT 0,
    `status`      TEXT NOT NULL DEFAULT 'active' CHECK(`status` IN ('active', 'archived')),
    `created_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`owner_id`) REFERENCES `sys_users`(`id`)
);

CREATE INDEX IF NOT EXISTS `idx_products_owner` ON `products` (`owner_id`);
CREATE INDEX IF NOT EXISTS `idx_products_status` ON `products` (`status`);

CREATE TRIGGER IF NOT EXISTS `tr_products_updated_at`
AFTER UPDATE ON `products`
FOR EACH ROW
BEGIN
    UPDATE `products` SET `updated_at` = CURRENT_TIMESTAMP WHERE `id` = NEW.`id`;
END;
