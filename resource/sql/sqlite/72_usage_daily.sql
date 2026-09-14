-- 迁移 72：使用分析 P2——CLI 版本列 + 日汇总物化表
-- client_version：埋点时从 User-Agent 解析（bcode/x.y.z），空=未识别
--（CLI 未带版本号的存量与 Go 默认 UA 都归未识别，不阻塞未来点亮）
ALTER TABLE `usage_events` ADD COLUMN `client_version` TEXT NOT NULL DEFAULT '';

-- usage_daily：按日×客户端的汇总物化。usage_events 明细 30 天清理，
-- 本表长期保留趋势；物化在每日维护链 dbclean 之前执行，报表查询时
-- 惰性补跑兜底重启缺口（幂等：重跑日先删后插）
CREATE TABLE IF NOT EXISTS `usage_daily` (
    `id`            INTEGER PRIMARY KEY AUTOINCREMENT,
    `day`           TEXT NOT NULL,
    `client`        TEXT NOT NULL,
    `calls`         INTEGER NOT NULL DEFAULT 0,
    `err_calls`     INTEGER NOT NULL DEFAULT 0,
    `active_actors` INTEGER NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS `idx_usage_daily_day_client` ON `usage_daily` (`day`, `client`);
