-- 迁移 69：Agent capabilities 生效（PRD §5）
-- agent_project_bindings 加 capabilities 列：逗号分隔能力 key（8 项字典见
-- perm.AgentCaps），空 = 全部能力——存量绑定零回填、行为不变。
-- 接入（join）不选能力；接入后由管理侧（owner/maintainer）按需收紧。
ALTER TABLE `agent_project_bindings` ADD COLUMN `capabilities` TEXT NOT NULL DEFAULT '';
