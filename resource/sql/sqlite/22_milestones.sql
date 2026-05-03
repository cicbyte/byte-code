-- 版本（里程碑）表
CREATE TABLE IF NOT EXISTS `milestones` (
    `id`          INTEGER PRIMARY KEY AUTOINCREMENT,
    `product_id`  INTEGER NOT NULL,
    `name`        TEXT NOT NULL,
    `description` TEXT DEFAULT '',
    `target_date` TEXT DEFAULT '',
    `status`      TEXT NOT NULL DEFAULT 'planning' CHECK(`status` IN ('planning', 'in_progress', 'released')),
    `created_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`product_id`) REFERENCES `products`(`id`)
);

CREATE INDEX IF NOT EXISTS `idx_milestones_product` ON `milestones` (`product_id`);
