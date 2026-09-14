-- 同 sqlite/73：附件路径键（文档页附件，#430）
ALTER TABLE `attachments` ADD COLUMN `entity_key` TEXT NULL;
CREATE INDEX `idx_attachments_entity_key` ON `attachments` (`entity_type`, `entity_key`);
