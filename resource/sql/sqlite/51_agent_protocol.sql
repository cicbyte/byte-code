-- 外部 Agent 接入协议（dev-docs/agent-protocol.md）：准入码 / 项目绑定 / 工作会话

-- 项目接入码：owner 生成，Agent 自助 join（一次性/24h/绑定项目）
CREATE TABLE IF NOT EXISTS `agent_join_codes` (
    `id`         INTEGER PRIMARY KEY AUTOINCREMENT,
    `code`       TEXT NOT NULL UNIQUE,
    `project_id` INTEGER NOT NULL,
    `role`       TEXT NOT NULL DEFAULT 'member',
    `created_by` INTEGER NOT NULL DEFAULT 0,
    `expires_at` TIMESTAMP NOT NULL,
    `used_at`    TIMESTAMP NULL,
    `used_by`    INTEGER NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS `idx_join_codes_code` ON `agent_join_codes` (`code`);

-- Agent ⇄ 项目 多对多准入（Agent= sys_users type='ai'）
CREATE TABLE IF NOT EXISTS `agent_project_bindings` (
    `id`         INTEGER PRIMARY KEY AUTOINCREMENT,
    `agent_id`   INTEGER NOT NULL,
    `project_id` INTEGER NOT NULL,
    `role`       TEXT NOT NULL DEFAULT 'member',
    `joined_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(`agent_id`, `project_id`)
);
CREATE INDEX IF NOT EXISTS `idx_bindings_project` ON `agent_project_bindings` (`project_id`);

-- 工作会话：key(agent)+project 派生，键为 (agent_id, project_id)
CREATE TABLE IF NOT EXISTS `agent_sessions` (
    `id`         INTEGER PRIMARY KEY AUTOINCREMENT,
    `session_id` TEXT NOT NULL UNIQUE,
    `agent_id`   INTEGER NOT NULL,
    `project_id` INTEGER NOT NULL,
    `expires_at` TIMESTAMP NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(`agent_id`, `project_id`)
);
