-- 同 sqlite/73：附件路径键（文档页附件）
-- entity_key 用 VARCHAR 而非 TEXT：MySQL TEXT 列建索引必须给前缀长度（Error 1170），
-- 长度对齐 API max-length:256（基线 s3_key VARCHAR(512) UNIQUE 同理在 3072B 索引限内）
ALTER TABLE `attachments` ADD COLUMN `entity_key` VARCHAR(256) NOT NULL DEFAULT '';
CREATE INDEX `idx_attachments_entity_key` ON `attachments` (`entity_type`, `entity_key`);
