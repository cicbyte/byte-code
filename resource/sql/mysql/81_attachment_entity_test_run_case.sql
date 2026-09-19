-- 同 sqlite/81：entity_type 枚举扩 test_run_case（并补 project 与 Go 白名单对齐）
ALTER TABLE `attachments` MODIFY COLUMN `entity_type` VARCHAR(32) NOT NULL
    CHECK(`entity_type` IN ('task', 'doc', 'test_case', 'test_run_case', 'requirement', 'comment', 'project'));
