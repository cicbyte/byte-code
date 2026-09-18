-- 同 sqlite/79：test_run_cases 补 bug_task_id
ALTER TABLE `test_run_cases` ADD COLUMN `bug_task_id` INT NOT NULL DEFAULT 0;
