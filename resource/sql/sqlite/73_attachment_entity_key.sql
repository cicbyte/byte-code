-- 迁移 73：附件路径键（文档页附件）
-- attachments.entity_id 是整数，任务/需求/测试用例等整数 id 实体直接用；
-- 现行文档是路径型 vault 无整数 id——entity_type='doc' 改走 entity_key，
-- 格式 "{projectId}:{docPath}"（如 "6:/design/arch.md"），权限按解析出的
-- 项目 id 校验。entity_id 对 doc 恒为 0（旧 docs 表已废弃）
ALTER TABLE `attachments` ADD COLUMN `entity_key` TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS `idx_attachments_entity_key` ON `attachments` (`entity_type`, `entity_key`);
