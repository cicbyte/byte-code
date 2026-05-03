-- 测试用例表
CREATE TABLE IF NOT EXISTS `test_cases` (
    `id`              INTEGER PRIMARY KEY AUTOINCREMENT,
    `project_id`      INTEGER NOT NULL,
    `requirement_id`  INTEGER DEFAULT 0,
    `task_id`         INTEGER DEFAULT 0,
    `title`           TEXT NOT NULL,
    `preconditions`   TEXT DEFAULT '',
    `steps`           TEXT NOT NULL DEFAULT '',
    `expected_result` TEXT NOT NULL DEFAULT '',
    `category`        TEXT DEFAULT '',
    `module`          TEXT DEFAULT '',
    `priority`        INTEGER NOT NULL DEFAULT 3,
    `source`          TEXT NOT NULL DEFAULT 'human' CHECK(`source` IN ('human', 'ai_generated')),
    `creator_id`      INTEGER NOT NULL DEFAULT 0,
    `status`          TEXT NOT NULL DEFAULT 'active' CHECK(`status` IN ('active', 'deprecated')),
    `created_at`      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at`      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`project_id`) REFERENCES `projects`(`id`),
    FOREIGN KEY (`creator_id`) REFERENCES `sys_users`(`id`)
);

CREATE INDEX IF NOT EXISTS `idx_test_cases_project` ON `test_cases` (`project_id`);
CREATE INDEX IF NOT EXISTS `idx_test_cases_status` ON `test_cases` (`status`);

-- 测试计划表
CREATE TABLE IF NOT EXISTS `test_plans` (
    `id`          INTEGER PRIMARY KEY AUTOINCREMENT,
    `project_id`  INTEGER NOT NULL,
    `name`        TEXT NOT NULL,
    `description` TEXT DEFAULT '',
    `milestone_id` INTEGER DEFAULT 0,
    `status`      TEXT NOT NULL DEFAULT 'draft' CHECK(`status` IN ('draft', 'running', 'completed')),
    `creator_id`  INTEGER NOT NULL DEFAULT 0,
    `created_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`project_id`) REFERENCES `projects`(`id`),
    FOREIGN KEY (`creator_id`) REFERENCES `sys_users`(`id`)
);

CREATE INDEX IF NOT EXISTS `idx_test_plans_project` ON `test_plans` (`project_id`);

-- 测试计划-用例关联表
CREATE TABLE IF NOT EXISTS `test_plan_cases` (
    `id`           INTEGER PRIMARY KEY AUTOINCREMENT,
    `test_plan_id` INTEGER NOT NULL,
    `test_case_id` INTEGER NOT NULL,
    `assignee_id`  INTEGER DEFAULT 0,
    `status`       TEXT NOT NULL DEFAULT 'pending' CHECK(`status` IN ('pending', 'pass', 'fail', 'blocked', 'skip')),
    `actual_result` TEXT DEFAULT '',
    `bug_task_id`  INTEGER DEFAULT 0,
    `executed_at`  TEXT DEFAULT '',
    FOREIGN KEY (`test_plan_id`) REFERENCES `test_plans`(`id`),
    FOREIGN KEY (`test_case_id`) REFERENCES `test_cases`(`id`),
    FOREIGN KEY (`assignee_id`) REFERENCES `sys_users`(`id`)
);

CREATE INDEX IF NOT EXISTS `idx_tpc_plan` ON `test_plan_cases` (`test_plan_id`);
CREATE INDEX IF NOT EXISTS `idx_tpc_case` ON `test_plan_cases` (`test_case_id`);

CREATE TRIGGER IF NOT EXISTS `tr_test_cases_updated_at`
AFTER UPDATE ON `test_cases`
FOR EACH ROW
BEGIN
    UPDATE `test_cases` SET `updated_at` = CURRENT_TIMESTAMP WHERE `id` = NEW.`id`;
END;
