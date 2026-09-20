-- 迁移 79：test_run_cases 补 bug_task_id（失败→缺陷闭环挂钩；
-- 命名对齐 test_plan_cases 既有同名列）
ALTER TABLE `test_run_cases` ADD COLUMN `bug_task_id` INTEGER NOT NULL DEFAULT 0;
