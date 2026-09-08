-- 清理路径索引：ai_execution_logs（detail 存完整 LLM 输出，增长最快的表）
-- 的 created_at 清理此前全表扫；顺带补 agent 协议两张小表的有效期索引
--（dbclean 的 expires_at 清理同样全表扫，表小但顺手一起）
CREATE INDEX IF NOT EXISTS `idx_ai_logs_created` ON `ai_execution_logs` (`created_at`);
CREATE INDEX IF NOT EXISTS `idx_sessions_expires` ON `agent_sessions` (`expires_at`);
CREATE INDEX IF NOT EXISTS `idx_join_codes_expires` ON `agent_join_codes` (`expires_at`);
