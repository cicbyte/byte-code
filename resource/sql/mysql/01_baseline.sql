-- ============================================================
-- ByteCode MySQL 基线（01）：全新库一次性建表 + 种子数据
-- 由 SQLite schema（迁移 01-62 终态）机械转换 + 人工审校生成。
-- 版本要求：MySQL >= 8.0.13 / MariaDB >= 10.2（TEXT 函数式默认值）。
-- 与 SQLite 的有意差异：
--   1) updated_at 触发器 → ON UPDATE CURRENT_TIMESTAMP 列属性
--   2) 外键声明不迁移：业务约定「0=无引用」与 FK 强约束冲突（SQLite 侧 FK 本就未开启）
--   3) FTS5 全文表不迁移：搜索实现走 LIKE，方言无关
--   4) 索引/唯一键涉及的 TEXT 列改为 VARCHAR；字符集 utf8mb4_unicode_ci（大小写不敏感，
--      与 SQLite 「name COLLATE NOCASE」唯一索引语义一致）
-- 后续增量迁移：resource/sql/mysql/ 下与 sqlite 同号同名双轨维护。
-- ============================================================
SET NAMES utf8mb4;

CREATE TABLE `categories` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(128) NOT NULL UNIQUE,
    `description` TEXT NOT NULL,
    `icon` TEXT ,
    `sort` INT NOT NULL DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `sys_users` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `username` VARCHAR(64) NOT NULL UNIQUE,
    `password` TEXT NOT NULL,
    `real_name` TEXT NOT NULL DEFAULT (''),
    `avatar` TEXT DEFAULT (''),
    `email` TEXT DEFAULT (''),
    `phone` TEXT DEFAULT (''),
    `desc` TEXT DEFAULT (''),
    `address` TEXT DEFAULT (''),
    `status` INT NOT NULL DEFAULT 1,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `type` VARCHAR(16) NOT NULL DEFAULT 'human' CHECK(`type` IN ('human', 'ai')),
    `capabilities` TEXT DEFAULT (''),
    `api_key` VARCHAR(255) DEFAULT '',
    `api_key_salt` TEXT DEFAULT (''),
    `owner_human_id` INT DEFAULT 0,
    `must_change_password` INT NOT NULL DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `sys_tokens` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `user_id` INT NOT NULL,
    `token` VARCHAR(512) NOT NULL UNIQUE,
    `expired_at` DATETIME NOT NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `sys_menus` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `parent_id` INT NOT NULL DEFAULT 0,
    `name` TEXT NOT NULL,
    `path` TEXT DEFAULT (''),
    `component` TEXT DEFAULT (''),
    `redirect` TEXT DEFAULT (''),
    `title` TEXT NOT NULL,
    `icon` TEXT DEFAULT (''),
    `sort` INT NOT NULL DEFAULT 0,
    `status` INT NOT NULL DEFAULT 1,
    `hidden` INT NOT NULL DEFAULT 0,
    `type` INT NOT NULL DEFAULT 1,
    `auth` TEXT DEFAULT (''),
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `sys_roles` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(191) NOT NULL,
    `explain` TEXT DEFAULT (''),
    `is_default` INT NOT NULL DEFAULT 0,
    `status` TEXT NOT NULL DEFAULT ('normal'),
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `sys_role_menus` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `role_id` INT NOT NULL,
    `menu_id` INT NOT NULL,
    UNIQUE(`role_id`, `menu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `sys_user_roles` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `user_id` INT NOT NULL,
    `role_id` INT NOT NULL,
    UNIQUE(`user_id`, `role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `sys_config` (
    `key` VARCHAR(100) NOT NULL PRIMARY KEY,
    `value` TEXT NOT NULL DEFAULT ('')
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `project_members` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `project_id` INT NOT NULL,
    `user_id` INT NOT NULL,
    `role` TEXT NOT NULL DEFAULT ('member'),
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(`project_id`, `user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `comments` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `task_id` INT NOT NULL,
    `user_id` INT NOT NULL,
    `content` TEXT NOT NULL,
    `user_type` TEXT NOT NULL DEFAULT ('human') CHECK(`user_type` IN ('human', 'ai')),
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `parent_id` INT NOT NULL DEFAULT 0,
    `updated_at` TEXT DEFAULT ('')
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `ai_execution_logs` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `task_id` INT NOT NULL,
    `ai_user_id` INT NOT NULL,
    `action` TEXT NOT NULL,
    `detail` TEXT DEFAULT (''),
    `status` TEXT NOT NULL DEFAULT ('success') CHECK(`status` IN ('success', 'failed')),
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `topic_id` INT NOT NULL DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `sprints` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `project_id` INT NOT NULL,
    `name` VARCHAR(191) NOT NULL,
    `goal` TEXT DEFAULT (''),
    `start_date` TEXT NOT NULL,
    `end_date` TEXT NOT NULL,
    `status` VARCHAR(32) NOT NULL DEFAULT 'planning' CHECK(`status` IN ('planning', 'active', 'completed')),
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `test_cases` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `project_id` INT NOT NULL,
    `requirement_id` INT DEFAULT 0,
    `task_id` INT DEFAULT 0,
    `title` TEXT NOT NULL,
    `preconditions` TEXT DEFAULT (''),
    `steps` TEXT NOT NULL DEFAULT (''),
    `expected_result` TEXT NOT NULL DEFAULT (''),
    `category` TEXT DEFAULT (''),
    `module` TEXT DEFAULT (''),
    `priority` INT NOT NULL DEFAULT 3,
    `source` TEXT NOT NULL DEFAULT ('human') CHECK(`source` IN ('human', 'ai_generated')),
    `creator_id` INT NOT NULL DEFAULT 0,
    `status` VARCHAR(32) NOT NULL DEFAULT 'active' CHECK(`status` IN ('active', 'deprecated')),
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `test_plans` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `project_id` INT NOT NULL,
    `name` TEXT NOT NULL,
    `description` TEXT DEFAULT (''),
    `milestone_id` INT DEFAULT 0,
    `status` TEXT NOT NULL DEFAULT ('draft') CHECK(`status` IN ('draft', 'running', 'completed')),
    `creator_id` INT NOT NULL DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `test_plan_cases` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `test_plan_id` INT NOT NULL,
    `test_case_id` INT NOT NULL,
    `assignee_id` INT DEFAULT 0,
    `status` TEXT NOT NULL DEFAULT ('pending') CHECK(`status` IN ('pending', 'pass', 'fail', 'blocked', 'skip')),
    `actual_result` TEXT DEFAULT (''),
    `bug_task_id` INT DEFAULT 0,
    `executed_at` TEXT DEFAULT ('')
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `attachments` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `s3_key` VARCHAR(512) NOT NULL UNIQUE,
    `original_name` TEXT NOT NULL,
    `file_size` INT NOT NULL DEFAULT 0,
    `mime_type` TEXT NOT NULL DEFAULT (''),
    `file_ext` TEXT DEFAULT (''),
    `entity_type` VARCHAR(32) NOT NULL CHECK(`entity_type` IN ('task', 'doc', 'test_case', 'requirement', 'comment')),
    `entity_id` INT NOT NULL DEFAULT 0,
    `uploader_id` INT NOT NULL DEFAULT 0,
    `description` TEXT DEFAULT (''),
    `download_count` INT NOT NULL DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `audit_logs` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `actor_id` INT NOT NULL DEFAULT 0,
    `actor_type` TEXT NOT NULL DEFAULT ('human') CHECK(`actor_type` IN ('human', 'ai', 'system')),
    `action` TEXT NOT NULL DEFAULT (''),
    `target_type` VARCHAR(32) NOT NULL DEFAULT '',
    `target_id` INT NOT NULL DEFAULT 0,
    `target_name` TEXT DEFAULT (''),
    `changes` TEXT DEFAULT (''),
    `ip_address` TEXT DEFAULT (''),
    `user_agent` TEXT DEFAULT (''),
    `project_id` INT DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `tags` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(128) NOT NULL UNIQUE,
    `color` TEXT NOT NULL DEFAULT ('#1890ff'),
    `creator_id` INT NOT NULL DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `entity_tags` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `tag_id` INT NOT NULL,
    `entity_type` VARCHAR(32) NOT NULL CHECK(`entity_type` IN ('task', 'requirement', 'test_case')),
    `entity_id` INT NOT NULL,
    UNIQUE(`tag_id`, `entity_type`, `entity_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `activities` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `actor_id` INT NOT NULL DEFAULT 0,
    `actor_type` TEXT NOT NULL DEFAULT ('human') CHECK(`actor_type` IN ('human', 'ai', 'system')),
    `actor_name` TEXT DEFAULT (''),
    `action` TEXT NOT NULL DEFAULT (''),
    `target_type` TEXT NOT NULL DEFAULT (''),
    `target_id` INT NOT NULL DEFAULT 0,
    `target_name` TEXT DEFAULT (''),
    `project_id` INT DEFAULT 0,
    `detail` TEXT DEFAULT (''),
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `notifications` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `user_id` INT NOT NULL DEFAULT 0,
    `title` TEXT NOT NULL DEFAULT (''),
    `content` TEXT DEFAULT (''),
    `type` TEXT NOT NULL DEFAULT ('info') CHECK(`type` IN ('info', 'warning', 'success', 'error')),
    `is_read` INT NOT NULL DEFAULT 0,
    `source_type` TEXT DEFAULT (''),
    `source_id` INT DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `plugins` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(191) NOT NULL UNIQUE,
    `version` TEXT NOT NULL DEFAULT ('1.0.0'),
    `description` TEXT DEFAULT (''),
    `author` TEXT DEFAULT (''),
    `enabled` INT NOT NULL DEFAULT 1,
    `config` TEXT DEFAULT ('{}'),
    `installed_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `project_databases` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `project_id` INT NOT NULL UNIQUE,
    `db_type` TEXT NOT NULL DEFAULT ('none') CHECK(`db_type` IN ('none', 'sqlite', 'mysql', 'postgresql')),
    `db_name` TEXT DEFAULT (''),
    `db_host` TEXT DEFAULT (''),
    `db_port` INT DEFAULT 0,
    `db_user` TEXT DEFAULT (''),
    `db_password` TEXT DEFAULT (''),
    `db_options` TEXT DEFAULT ('{}'),
    `connection_status` TEXT NOT NULL DEFAULT ('disconnected') CHECK(`connection_status` IN ('disconnected', 'connected', 'error')),
    `last_tested_at` TEXT DEFAULT (''),
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `requirements` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `project_id` INT NOT NULL DEFAULT 0,
    `parent_id` INT DEFAULT 0,
    `type` TEXT NOT NULL DEFAULT ('story') CHECK(`type` IN ('epic', 'story', 'task')),
    `title` TEXT NOT NULL,
    `description` TEXT DEFAULT (''),
    `status` TEXT NOT NULL DEFAULT ('draft') CHECK(`status` IN ('draft', 'confirmed', 'decomposed', 'implemented', 'closed')),
    `priority` INT NOT NULL DEFAULT 3,
    `assignee_id` INT DEFAULT 0,
    `creator_id` INT NOT NULL DEFAULT 0,
    `milestone_id` INT DEFAULT 0,
    `sort_order` INT NOT NULL DEFAULT 0,
    `acceptance_criteria` TEXT DEFAULT (''),
    `source` TEXT NOT NULL DEFAULT ('human') CHECK(`source` IN ('human', 'ai_decomposed')),
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `milestones` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `project_id` INT NOT NULL DEFAULT 0,
    `name` VARCHAR(191) NOT NULL,
    `description` TEXT DEFAULT (''),
    `target_date` TEXT DEFAULT (''),
    `status` TEXT NOT NULL DEFAULT ('planning') CHECK(`status` IN ('planning', 'in_progress', 'released')),
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `projects` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(191) NOT NULL,
    `description` TEXT DEFAULT (''),
    `created_by` INT NOT NULL DEFAULT 0,
    `status` INT NOT NULL DEFAULT 1,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `code` VARCHAR(16) NOT NULL DEFAULT ''
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `project_memories` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `project_id` INT NOT NULL DEFAULT 0,
    `key` VARCHAR(191) NOT NULL,
    `value` TEXT NOT NULL DEFAULT (''),
    `status` VARCHAR(32) NOT NULL DEFAULT 'active' CHECK(`status` IN ('pending','active','stale','expired')),
    `expires_at` DATETIME NULL,
    `last_verified_at` DATETIME NULL,
    `verified_by` INT DEFAULT 0,
    `updated_by` INT NOT NULL DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE(`project_id`, `key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `project_document_index` (
    `project_id` INT NOT NULL,
    `path` VARCHAR(512) NOT NULL,
    `space` VARCHAR(32) DEFAULT '',
    `title` TEXT DEFAULT (''),
    `type` TEXT DEFAULT (''),
    `status` TEXT DEFAULT ('published'),
    `tags` TEXT DEFAULT (''),
    `linked` TEXT DEFAULT (''),
    `ext` TEXT DEFAULT (''),
    `size` INT DEFAULT 0,
    `checksum` TEXT DEFAULT (''),
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`project_id`, `path`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `agent_join_codes` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `code` VARCHAR(64) NOT NULL UNIQUE,
    `project_id` INT NOT NULL,
    `role` TEXT NOT NULL DEFAULT ('member'),
    `created_by` INT NOT NULL DEFAULT 0,
    `expires_at` DATETIME NOT NULL,
    `used_at` DATETIME NULL,
    `used_by` INT NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `agent_project_bindings` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `agent_id` INT NOT NULL,
    `project_id` INT NOT NULL,
    `role` TEXT NOT NULL DEFAULT ('member'),
    `joined_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(`agent_id`, `project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `agent_sessions` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `session_id` VARCHAR(255) NOT NULL UNIQUE,
    `agent_id` INT NOT NULL,
    `project_id` INT NOT NULL,
    `expires_at` DATETIME NOT NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(`agent_id`, `project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `tasks` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `project_id` INT NOT NULL,
    `requirement_id` INT DEFAULT 0,
    `sprint_id` INT DEFAULT 0,
    `title` TEXT NOT NULL,
    `description` TEXT DEFAULT (''),
    `type` TEXT NOT NULL DEFAULT ('feature') CHECK(`type` IN ('feature', 'bug', 'chore', 'test')),
    `status` VARCHAR(32) NOT NULL DEFAULT 'open' CHECK(`status` IN ('open', 'in_progress', 'blocked', 'review', 'done', 'closed')),
    `priority` INT NOT NULL DEFAULT 3,
    `assignee_id` INT DEFAULT 0,
    `creator_id` INT NOT NULL DEFAULT 0,
    `parent_task_id` INT DEFAULT 0,
    `artifacts` TEXT DEFAULT (''),
    `requires_human_review` INT NOT NULL DEFAULT 0,
    `human_review_status` TEXT NOT NULL DEFAULT ('pending') CHECK(`human_review_status` IN ('pending', 'approved', 'rejected')),
    `sort_order` INT NOT NULL DEFAULT 0,
    `source` TEXT NOT NULL DEFAULT ('human') CHECK(`source` IN ('human', 'ai', 'pm_import')),
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `completed_at` VARCHAR(32) NOT NULL DEFAULT '',
    `ai_attempts` INT NOT NULL DEFAULT 0,
    `ai_next_attempt_at` VARCHAR(32) NOT NULL DEFAULT '',
    `due_date` VARCHAR(32) NULL,
    `checklist` TEXT NOT NULL DEFAULT ('[]'),
    `related_refs` TEXT NOT NULL DEFAULT ('[]')
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `project_relations` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `project_id` INT NOT NULL,
    `related_project_id` INT NOT NULL,
    `created_by` INT NOT NULL DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(`project_id`, `related_project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `project_feedbacks` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `project_id` INT NOT NULL,
    `source_project_id` INT NOT NULL,
    `source_task_id` INT DEFAULT 0,
    `title` TEXT NOT NULL,
    `content` TEXT DEFAULT (''),
    `status` VARCHAR(32) NOT NULL DEFAULT 'open' CHECK(`status` IN ('open', 'converted', 'dismissed')),
    `converted_task_id` INT DEFAULT 0,
    `handled_by` INT DEFAULT 0,
    `handled_at` TEXT DEFAULT (''),
    `dismiss_reason` TEXT DEFAULT (''),
    `created_by` INT NOT NULL DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `topics` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `project_id` INT NOT NULL,
    `title` TEXT NOT NULL,
    `goal` TEXT DEFAULT (''),
    `acceptance` TEXT DEFAULT (''),
    `doc_path` TEXT DEFAULT (''),
    `assignee_id` INT DEFAULT 0,
    `status` VARCHAR(32) NOT NULL DEFAULT 'active' CHECK(`status` IN ('active', 'completed', 'abandoned')),
    `created_by` INT NOT NULL DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `completed_at` TEXT DEFAULT ('')
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `topic_phases` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `topic_id` INT NOT NULL,
    `title` TEXT NOT NULL,
    `detail` TEXT DEFAULT (''),
    `sort_order` INT NOT NULL DEFAULT 0,
    `status` VARCHAR(32) NOT NULL DEFAULT 'pending' CHECK(`status` IN ('pending', 'in_progress', 'done')),
    `task_id` INT DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `assignee_id` INT NOT NULL DEFAULT 0,
    `artifacts` TEXT NOT NULL DEFAULT ('')
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `project_qas` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `project_id` INT NOT NULL,
    `question` TEXT NOT NULL,
    `answer` TEXT DEFAULT (''),
    `tags` TEXT DEFAULT (''),
    `hits` INT NOT NULL DEFAULT 0,
    `status` VARCHAR(32) NOT NULL DEFAULT 'active' CHECK(`status` IN ('active', 'archived')),
    `created_by` INT NOT NULL DEFAULT 0,
    `updated_by` INT NOT NULL DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE INDEX `idx_categories_sort` ON `categories` (`sort`);
CREATE INDEX `idx_categories_created_at` ON `categories` (`created_at`);
CREATE INDEX `idx_sys_users_username` ON `sys_users` (`username`);
CREATE INDEX `idx_sys_users_status` ON `sys_users` (`status`);
CREATE INDEX `idx_sys_tokens_token` ON `sys_tokens` (`token`);
CREATE INDEX `idx_sys_tokens_user_id` ON `sys_tokens` (`user_id`);
CREATE INDEX `idx_sys_menus_parent_id` ON `sys_menus` (`parent_id`);
CREATE INDEX `idx_sys_menus_status` ON `sys_menus` (`status`);
CREATE INDEX `idx_sys_users_type` ON `sys_users` (`type`);
CREATE INDEX `idx_sys_users_api_key` ON `sys_users` (`api_key`);
CREATE INDEX `idx_project_members_project` ON `project_members` (`project_id`);
CREATE INDEX `idx_project_members_user` ON `project_members` (`user_id`);
CREATE INDEX `idx_comments_task` ON `comments` (`task_id`);
CREATE INDEX `idx_ai_logs_task` ON `ai_execution_logs` (`task_id`);
CREATE INDEX `idx_sprints_project` ON `sprints` (`project_id`);
CREATE INDEX `idx_sprints_status` ON `sprints` (`status`);
CREATE INDEX `idx_test_cases_project` ON `test_cases` (`project_id`);
CREATE INDEX `idx_test_cases_status` ON `test_cases` (`status`);
CREATE INDEX `idx_test_plans_project` ON `test_plans` (`project_id`);
CREATE INDEX `idx_tpc_plan` ON `test_plan_cases` (`test_plan_id`);
CREATE INDEX `idx_tpc_case` ON `test_plan_cases` (`test_case_id`);
CREATE INDEX `idx_attachments_entity` ON `attachments` (`entity_type`, `entity_id`);
CREATE INDEX `idx_attachments_s3_key` ON `attachments` (`s3_key`);
CREATE INDEX `idx_audit_logs_target` ON `audit_logs` (`target_type`, `target_id`);
CREATE INDEX `idx_audit_logs_actor` ON `audit_logs` (`actor_id`);
CREATE INDEX `idx_audit_logs_created` ON `audit_logs` (`created_at`);
CREATE INDEX `idx_audit_logs_project` ON `audit_logs` (`project_id`);
CREATE INDEX `idx_entity_tags_tag` ON `entity_tags` (`tag_id`);
CREATE INDEX `idx_entity_tags_entity` ON `entity_tags` (`entity_type`, `entity_id`);
CREATE INDEX `idx_activities_created` ON `activities` (`created_at`);
CREATE INDEX `idx_activities_project` ON `activities` (`project_id`);
CREATE INDEX `idx_activities_actor` ON `activities` (`actor_id`);
CREATE INDEX `idx_notifications_user` ON `notifications` (`user_id`);
CREATE INDEX `idx_notifications_read` ON `notifications` (`user_id`, `is_read`);
CREATE INDEX `idx_requirements_project` ON `requirements` (`project_id`);
CREATE INDEX `idx_milestones_project` ON `milestones` (`project_id`);
CREATE INDEX `idx_projects_status` ON `projects` (`status`);
CREATE INDEX `idx_comments_user` ON `comments` (`user_id`);
CREATE INDEX `idx_ai_logs_ai_user` ON `ai_execution_logs` (`ai_user_id`);
CREATE INDEX `idx_requirements_milestone` ON `requirements` (`milestone_id`);
CREATE INDEX `idx_attachments_uploader` ON `attachments` (`uploader_id`);
CREATE INDEX `idx_tpc_assignee` ON `test_plan_cases` (`assignee_id`);
CREATE INDEX `idx_activities_project_created` ON `activities` (`project_id`, `created_at`);
CREATE INDEX `idx_audit_logs_project_created` ON `audit_logs` (`project_id`, `created_at`);
CREATE UNIQUE INDEX `uniq_sys_roles_name` ON `sys_roles` (`name`);
CREATE UNIQUE INDEX `uniq_sprints_project_name` ON `sprints` (`project_id`, `name`);
CREATE UNIQUE INDEX `uniq_milestones_project_name` ON `milestones` (`project_id`, `name`);
CREATE INDEX `idx_comments_parent` ON `comments` (`parent_id`);
CREATE INDEX `idx_memories_status` ON `project_memories` (`project_id`, `status`);
CREATE INDEX `idx_doc_index_tags` ON `project_document_index` (`project_id`, `space`);
CREATE UNIQUE INDEX `idx_projects_name` ON `projects` (name);
CREATE INDEX `idx_notifications_clean` ON `notifications` (is_read, created_at);
CREATE INDEX `idx_tokens_expired` ON `sys_tokens` (expired_at);
CREATE INDEX `idx_memories_expires` ON `project_memories` (status, expires_at);
CREATE INDEX `idx_memories_verified` ON `project_memories` (status, last_verified_at);
CREATE INDEX `idx_join_codes_code` ON `agent_join_codes` (`code`);
CREATE INDEX `idx_bindings_project` ON `agent_project_bindings` (`project_id`);
CREATE INDEX `idx_tasks_project` ON `tasks` (`project_id`);
CREATE INDEX `idx_tasks_status` ON `tasks` (`status`);
CREATE INDEX `idx_tasks_assignee` ON `tasks` (`assignee_id`);
CREATE INDEX `idx_tasks_sprint` ON `tasks` (`sprint_id`);
CREATE INDEX `idx_tasks_requirement` ON `tasks` (`requirement_id`);
CREATE INDEX `idx_tasks_parent` ON `tasks` (`parent_task_id`);
CREATE INDEX `idx_tasks_completed_at` ON `tasks` (`completed_at`);
CREATE INDEX `idx_tasks_ai_retry` ON `tasks` (`status`, `assignee_id`, `ai_next_attempt_at`);
CREATE INDEX `idx_tasks_due` ON `tasks` (`due_date`);
CREATE INDEX `idx_relations_project` ON `project_relations` (`project_id`);
CREATE INDEX `idx_feedbacks_project` ON `project_feedbacks` (`project_id`, `status`);
CREATE INDEX `idx_topics_project` ON `topics` (`project_id`, `status`);
CREATE INDEX `idx_topic_phases` ON `topic_phases` (`topic_id`, `status`);
CREATE INDEX `idx_ai_logs_topic` ON `ai_execution_logs` (`topic_id`);
CREATE UNIQUE INDEX `idx_projects_code` ON `projects` (`code`);
CREATE INDEX `idx_qas_project` ON `project_qas` (`project_id`, `status`);

-- ==================== 种子数据 ====================
-- 菜单（迁移 35/61/62 终态：权限字典 9 节点）
INSERT INTO `sys_menus` (`id`, `parent_id`, `name`, `path`, `component`, `redirect`, `title`, `icon`, `sort`, `status`, `hidden`, `type`, `auth`) VALUES
(10, 0, 'System', '/system', 'LAYOUT', '/system/menu', '系统管理', 'SettingOutlined', 90, 1, 0, 1, ''),
(11, 10, 'system_menu', 'menu', '/system/menu/menu', '', '菜单管理', '', 100, 1, 0, 2, 'system_menu'),
(12, 10, 'system_role', 'role', '/system/role/role', '', '角色管理', '', 90, 1, 0, 2, 'system_role'),
(61, 80, 'ai_users', 'agents', '/platform/agents', '', 'Agent 账号', '', 70, 1, 0, 2, ''),
(70, 0, 'Setting', '/setting', 'LAYOUT', '/setting/account', '设置', 'SettingOutlined', 30, 1, 0, 1, ''),
(72, 70, 'setting_system', 'system', '/setting/system/system', '', '系统设置', '', 90, 1, 0, 2, ''),
(73, 70, 'setting_global_memory', 'global-memory', '/setting/global-memory', '', '全局记忆', '', 80, 1, 0, 2, ''),
(80, 0, 'Platform', '/platform', 'LAYOUT', '/platform/activities', '平台', 'TeamOutlined', 20, 1, 0, 1, ''),
(83, 80, 'platform_audit', 'audit', '/platform/audit', '', '审计日志', '', 80, 1, 0, 2, '');

-- 角色（内置两条；普通用户不预置菜单，个人可见页面不走权限字典）
INSERT INTO `sys_roles` (`id`, `name`, `explain`, `is_default`, `status`) VALUES
(1, '超级管理员', '拥有全部权限', 1, 'normal'),
(2, '普通用户', '基础访问权限', 0, 'normal');

-- 超管角色绑定全量菜单
INSERT INTO `sys_role_menus` (`role_id`, `menu_id`) VALUES
(1, 10), (1, 11), (1, 12), (1, 61), (1, 70), (1, 72), (1, 73), (1, 80), (1, 83);

-- admin → 超级管理员
INSERT INTO `sys_user_roles` (`user_id`, `role_id`) VALUES (1, 1);

-- 默认管理员（admin/admin123，首登强制改密——与迁移 10+39 终态一致）
INSERT INTO `sys_users` (`id`, `username`, `password`, `real_name`, `avatar`, `email`, `phone`, `desc`, `address`, `status`, `type`, `capabilities`, `must_change_password`) VALUES
(1, 'admin', '$2a$10$CaObBvqb0g4xPcIvNMOYX.RVgZVRGWgaTxlqKmwIosvsGtAyrhq8q', 'Admin', '', 'admin@byte-code.com', '', 'manager', '', 1, 'human', '', 1);

-- 系统配置（迁移 11 默认值）
INSERT INTO `sys_config` (`key`, `value`) VALUES
('site_name', 'Byte Admin'),
('site_icp', ''),
('site_phone', ''),
('site_address', ''),
('login_captcha', '0'),
('site_open', '1'),
('site_close_text', '网站维护中，暂时无法访问！'),
('smtp_host', ''),
('smtp_port', '465'),
('smtp_user', ''),
('smtp_pass', ''),
('smtp_from', '');

-- 分类（迁移 01 默认值）
INSERT INTO `categories` (`id`, `name`, `description`, `icon`, `sort`) VALUES
(1, '技术', '技术相关文章分类', 'tech', 100),
(2, '生活', '生活随笔分类', 'life', 90),
(3, '学习', '学习笔记分类', 'study', 80);

