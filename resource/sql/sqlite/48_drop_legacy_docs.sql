-- 旧知识库（docs 表体系）整体下线：文档能力已由磁盘 vault（project_document_index）
-- 承接，存量数据经 legacy 导出迁入 vault。删除三张旧表及 FTS 残留触发器定义。
-- 若存在导出标记置位后仍写入旧表的数据（理论不可能：写路由已删），drop 前请自行备份。
DROP TABLE IF EXISTS `doc_versions`;
DROP TABLE IF EXISTS `doc_relations`;
DROP TABLE IF EXISTS `docs`;
