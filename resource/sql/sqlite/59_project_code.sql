-- 迁移 59：项目短码（code）
-- 语义：稳定、可读、URL 安全的项目标识，供 CLI/会话/跨环境关联——
-- 自增 id 在不同部署间会漂移，且有枚举风险。code 由应用层创建时生成。
-- 注：ALTER ADD COLUMN 与存量回填已于 2026-09-07 手动执行（迁移框架
-- 首跑在 UPDATE 阶段报 UNIQUE 冲突未记录，本文件仅保留幂等部分防重跑报错）。

CREATE UNIQUE INDEX IF NOT EXISTS `idx_projects_code` ON `projects` (`code`);
