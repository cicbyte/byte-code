-- 需求表
CREATE TABLE IF NOT EXISTS `requirements` (
    `id`                  INTEGER PRIMARY KEY AUTOINCREMENT,
    `product_id`          INTEGER NOT NULL,
    `parent_id`           INTEGER DEFAULT 0,
    `type`                TEXT NOT NULL DEFAULT 'story' CHECK(`type` IN ('epic', 'story', 'task')),
    `title`               TEXT NOT NULL,
    `description`         TEXT DEFAULT '',
    `status`              TEXT NOT NULL DEFAULT 'draft' CHECK(`status` IN ('draft', 'confirmed', 'decomposed', 'implemented', 'closed')),
    `priority`            INTEGER NOT NULL DEFAULT 3,
    `assignee_id`         INTEGER DEFAULT 0,
    `creator_id`          INTEGER NOT NULL DEFAULT 0,
    `milestone_id`        INTEGER DEFAULT 0,
    `sort_order`          INTEGER NOT NULL DEFAULT 0,
    `acceptance_criteria` TEXT DEFAULT '',
    `source`              TEXT NOT NULL DEFAULT 'human' CHECK(`source` IN ('human', 'ai_decomposed')),
    `created_at`          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at`          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`product_id`) REFERENCES `products`(`id`),
    FOREIGN KEY (`parent_id`) REFERENCES `requirements`(`id`),
    FOREIGN KEY (`assignee_id`) REFERENCES `sys_users`(`id`),
    FOREIGN KEY (`creator_id`) REFERENCES `sys_users`(`id`)
);

CREATE INDEX IF NOT EXISTS `idx_requirements_product` ON `requirements` (`product_id`);
CREATE INDEX IF NOT EXISTS `idx_requirements_parent` ON `requirements` (`parent_id`);
CREATE INDEX IF NOT EXISTS `idx_requirements_status` ON `requirements` (`status`);
CREATE INDEX IF NOT EXISTS `idx_requirements_type` ON `requirements` (`type`);

CREATE TRIGGER IF NOT EXISTS `tr_requirements_updated_at`
AFTER UPDATE ON `requirements`
FOR EACH ROW
BEGIN
    UPDATE `requirements` SET `updated_at` = CURRENT_TIMESTAMP WHERE `id` = NEW.`id`;
END;
