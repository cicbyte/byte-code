-- 迁移 59：项目短码（code）
-- 语义：稳定、可读、URL 安全的项目标识，供 CLI/会话/跨环境关联——
-- 自增 id 在不同部署间会漂移，且有枚举风险。code 由应用层创建时生成。
-- 存量回填：8 位小写十六进制，随机生成（碰撞概率可忽略，空值仅存量行才有）。
-- 注：dev 库曾手动补列后以仅含索引的缩水版本记录执行；此处恢复完整自洽
-- 内容——框架按文件名记账，已执行过的库不会重跑，仅影响新库首跑。

ALTER TABLE `projects` ADD COLUMN `code` TEXT NOT NULL DEFAULT '';
UPDATE `projects` SET `code` = lower(hex(randomblob(4))) WHERE `code` = '';
CREATE UNIQUE INDEX IF NOT EXISTS `idx_projects_code` ON `projects` (`code`);
