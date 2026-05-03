-- ByteCode 新模块菜单种子数据
INSERT OR IGNORE INTO `sys_menus` (`id`, `parent_id`, `name`, `path`, `component`, `redirect`, `title`, `icon`, `sort`, `status`, `hidden`, `type`, `auth`) VALUES
-- 产品管理
(20, 0, 'Product',           '/product',          'LAYOUT', '/product/list',   '产品管理', 'AppstoreOutlined', 80, 1, 0, 1, ''),
(21, 20, 'product_list',     'list',              '/product/list',              '', '产品列表', '', 100, 1, 0, 2, ''),
(22, 20, 'product_reqs',     'requirements',      '/product/requirements',      '', '需求池',   '', 90, 1, 0, 2, ''),
-- 项目管理
(30, 0, 'Project',           '/project',          'LAYOUT', '/project/list',   '项目管理', 'FolderOutlined', 70, 1, 0, 1, ''),
(31, 30, 'project_list',     'list',              '/project/list',              '', '项目列表', '', 100, 1, 0, 2, ''),
(32, 30, 'project_detail',   'detail/:id',        '/project/detail',            '', '项目详情', '', 90, 1, 0, 2, ''),
(33, 30, 'project_sprints',  'sprints',           '/project/sprints',           '', 'Sprint管理', '', 80, 1, 0, 2, ''),
-- 测试管理
(40, 0, 'Test',              '/test',             'LAYOUT', '/test/cases',     '测试管理', 'ExperimentOutlined', 60, 1, 0, 1, ''),
(41, 40, 'test_cases',       'cases',             '/test/cases',                '', '测试用例', '', 100, 1, 0, 2, ''),
(42, 40, 'test_plans',       'plans',             '/test/plans',                '', '测试计划', '', 90, 1, 0, 2, ''),
-- 知识库
(50, 0, 'Knowledge',         '/knowledge',        'LAYOUT', '/knowledge/docs', '知识库',   'BookOutlined', 50, 1, 0, 1, ''),
(51, 50, 'knowledge_docs',   'docs',              '/knowledge/docs',            '', '文档管理', '', 100, 1, 0, 2, ''),
-- AI 管理
(60, 0, 'AiManage',          '/ai-manage',        'LAYOUT', '/ai-manage/users','AI管理',  'RobotOutlined', 40, 1, 0, 1, ''),
(61, 60, 'ai_users',         'users',             '/ai-manage/users',           '', 'AI用户',   '', 100, 1, 0, 2, ''),
-- 平台
(80, 0, 'Platform',          '/platform',         'LAYOUT', '/platform/activities','平台', 'TeamOutlined', 20, 1, 0, 1, ''),
(81, 80, 'platform_activities','activities',      '/platform/activities',       '', '活动流',   '', 100, 1, 0, 2, ''),
(82, 80, 'platform_tags',    'tags',              '/platform/tags',             '', '标签管理', '', 90, 1, 0, 2, ''),
(83, 80, 'platform_audit',   'audit',             '/platform/audit',            '', '审计日志', '', 80, 1, 0, 2, '');
