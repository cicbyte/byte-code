-- 数据层 P2 加固：高频查询索引、低风险唯一约束、任务完成时间列
-- 注意：projects.name 的唯一性由业务层查重保证（存量重名清理代价高，不在 DB 层强约束）

-- ============ 高频查询索引 ============
CREATE INDEX IF NOT EXISTS `idx_comments_user` ON `comments` (`user_id`);
CREATE INDEX IF NOT EXISTS `idx_ai_logs_ai_user` ON `ai_execution_logs` (`ai_user_id`);
CREATE INDEX IF NOT EXISTS `idx_requirements_milestone` ON `requirements` (`milestone_id`);
CREATE INDEX IF NOT EXISTS `idx_tasks_parent` ON `tasks` (`parent_task_id`);
CREATE INDEX IF NOT EXISTS `idx_doc_versions_editor` ON `doc_versions` (`editor_id`);
CREATE INDEX IF NOT EXISTS `idx_attachments_uploader` ON `attachments` (`uploader_id`);
CREATE INDEX IF NOT EXISTS `idx_tpc_assignee` ON `test_plan_cases` (`assignee_id`);
CREATE INDEX IF NOT EXISTS `idx_activities_project_created` ON `activities` (`project_id`, `created_at`);
CREATE INDEX IF NOT EXISTS `idx_audit_logs_project_created` ON `audit_logs` (`project_id`, `created_at`);

-- ============ 唯一约束（先清重名保留最小 id，并解除引用）============
-- 角色重名：先清角色-菜单/角色-用户关联
DELETE FROM `sys_role_menus` WHERE `role_id` IN (
    SELECT `id` FROM `sys_roles` WHERE `id` NOT IN (SELECT MIN(`id`) FROM `sys_roles` GROUP BY `name`));
DELETE FROM `sys_user_roles` WHERE `role_id` IN (
    SELECT `id` FROM `sys_roles` WHERE `id` NOT IN (SELECT MIN(`id`) FROM `sys_roles` GROUP BY `name`));
DELETE FROM `sys_roles` WHERE `id` NOT IN (SELECT MIN(`id`) FROM `sys_roles` GROUP BY `name`);
CREATE UNIQUE INDEX IF NOT EXISTS `uniq_sys_roles_name` ON `sys_roles` (`name`);

-- 同项目内 Sprint 重名：先解除任务引用
UPDATE `tasks` SET `sprint_id` = 0 WHERE `sprint_id` IN (
    SELECT `id` FROM `sprints` WHERE `id` NOT IN (SELECT MIN(`id`) FROM `sprints` GROUP BY `project_id`, `name`));
DELETE FROM `sprints` WHERE `id` NOT IN (SELECT MIN(`id`) FROM `sprints` GROUP BY `project_id`, `name`);
CREATE UNIQUE INDEX IF NOT EXISTS `uniq_sprints_project_name` ON `sprints` (`project_id`, `name`);

-- 同项目内里程碑重名：先解除需求引用
UPDATE `requirements` SET `milestone_id` = 0 WHERE `milestone_id` IN (
    SELECT `id` FROM `milestones` WHERE `id` NOT IN (SELECT MIN(`id`) FROM `milestones` GROUP BY `project_id`, `name`));
DELETE FROM `milestones` WHERE `id` NOT IN (SELECT MIN(`id`) FROM `milestones` GROUP BY `project_id`, `name`);
CREATE UNIQUE INDEX IF NOT EXISTS `uniq_milestones_project_name` ON `milestones` (`project_id`, `name`);

-- ============ 任务完成时间（燃尽图数据质量：不再被后续编辑重写）============
ALTER TABLE `tasks` ADD COLUMN `completed_at` TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS `idx_tasks_completed_at` ON `tasks` (`completed_at`);
