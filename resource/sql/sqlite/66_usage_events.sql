-- 迁移 66：使用埋点事件表（CLI/平台行为分析地基）
-- 语义：服务端全量请求事件（读+写），供命令热度/失败率/延迟分位/漏斗分析。
-- 与 audit_logs 的边界：审计=合规留痕（只写操作、90 天不可清理），
-- usage_events=可过期统计数据（30 天清理），混表会打架。
-- endpoint 存模板化路径（/v1/tasks/{id}）防维度爆炸；params 存白名单抽取的
-- 参数 JSON（枚举值全记、自由文本记 _len、未知键记键名集合）。
CREATE TABLE IF NOT EXISTS `usage_events` (
    `id`          INTEGER PRIMARY KEY AUTOINCREMENT,
    `actor_id`    INTEGER NOT NULL DEFAULT 0,
    `actor_type`  TEXT    NOT NULL DEFAULT 'human',
    `client`      TEXT    NOT NULL DEFAULT 'web',
    `method`      TEXT    NOT NULL DEFAULT '',
    `endpoint`    TEXT    NOT NULL DEFAULT '',
    `status_code` INTEGER NOT NULL DEFAULT 0,
    `duration_ms` INTEGER NOT NULL DEFAULT 0,
    `project_id`  INTEGER NOT NULL DEFAULT 0,
    `session_id`  TEXT    NOT NULL DEFAULT '',
    `error_code`  INTEGER NOT NULL DEFAULT 0,
    `params`      TEXT    DEFAULT '',
    `created_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS `idx_usage_events_endpoint` ON `usage_events` (`endpoint`, `created_at`);
CREATE INDEX IF NOT EXISTS `idx_usage_events_actor` ON `usage_events` (`actor_id`, `created_at`);
CREATE INDEX IF NOT EXISTS `idx_usage_events_created` ON `usage_events` (`created_at`);
