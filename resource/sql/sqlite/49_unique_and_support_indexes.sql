-- 项目名唯一约束：业务层 count-then-insert 查重存在竞态（并发同名双双通过），
-- DB 层兜底。存量重名先清洗（保留最早创建的，晚的加后缀）
UPDATE projects SET name = name || '-dup-' || id
WHERE id NOT IN (SELECT MIN(id) FROM projects GROUP BY name COLLATE NOCASE);

CREATE UNIQUE INDEX IF NOT EXISTS idx_projects_name ON projects (name COLLATE NOCASE);

-- 清理/物化支撑索引：dbclean 的 DELETE 与 MemMaterialize 的批量 UPDATE 均走全表扫
CREATE INDEX IF NOT EXISTS idx_notifications_clean ON notifications (is_read, created_at);
CREATE INDEX IF NOT EXISTS idx_tokens_expired ON sys_tokens (expired_at);
CREATE INDEX IF NOT EXISTS idx_memories_expires ON project_memories (status, expires_at);
CREATE INDEX IF NOT EXISTS idx_memories_verified ON project_memories (status, last_verified_at);
