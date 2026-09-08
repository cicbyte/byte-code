-- 同 sqlite/63：清理路径索引（MySQL 基线之后的同号增量）
CREATE INDEX `idx_ai_logs_created` ON `ai_execution_logs` (`created_at`);
CREATE INDEX `idx_sessions_expires` ON `agent_sessions` (`expires_at`);
CREATE INDEX `idx_join_codes_expires` ON `agent_join_codes` (`expires_at`);
