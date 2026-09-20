import { RouteRecordRaw } from 'vue-router';
import { Layout } from '@/router/constant';
import { ProjectOutlined } from '@vicons/antd';
import { renderIcon } from '@/utils/index';

const routeName = 'project';

const routes: Array<RouteRecordRaw> = [
  {
    path: '/project',
    name: routeName,
    redirect: '/project/list',
    component: Layout,
    meta: {
      title: '项目管理',
      icon: renderIcon(ProjectOutlined),
      sort: 3,
    },
    children: [
      {
        path: 'list',
        name: `${routeName}_list`,
        meta: {
          title: '项目列表',
        },
        component: () => import('@/views/project/list/index.vue'),
      },
      {
        // 分组私有化：归属创建者，任何登录用户可管理自己的分组
        //（入口不再挂权限字典；数据边界在后端按 created_by 过滤）
        path: 'groups',
        name: `${routeName}_groups`,
        meta: {
          title: '项目分组',
        },
        component: () => import('@/views/platform/groups/index.vue'),
      },
      {
        path: ':projectId',
        name: `${routeName}_workspace`,
        meta: {
          title: '项目详情',
          hideInMenu: true,
        },
        // 被权限过滤的静态子路由（如 /project/groups 无权限时）会跌进本
        // 参数路由形成空工作台；非数字 projectId 一律 404，不再静默空白
        beforeEnter: (to) =>
          /^\d+$/.test(String(to.params.projectId || '')) || { name: 'ErrorPageSon' },
        component: () => import('@/views/project/workspace/index.vue'),
        children: [
          {
            path: 'overview',
            name: `${routeName}_overview`,
            meta: {
              title: '项目概览',
              hideInMenu: true,
            },
            component: () => import('@/views/project/overview/index.vue'),
          },
          {
            path: 'board',
            name: `${routeName}_board`,
            meta: {
              title: '任务看板',
              group: '任务',
              hideInMenu: true,
            },
            component: () => import('@/views/project/board/index.vue'),
          },
          {
            path: 'tasks',
            name: `${routeName}_tasks`,
            meta: {
              title: '任务列表',
              group: '任务',
              hideInMenu: true,
            },
            component: () => import('@/views/project/tasks/index.vue'),
          },
          {
            // 审核工作台：集中验收（完成即待人审，CompleteTask 统一置位）
            path: 'reviews',
            name: `${routeName}_reviews`,
            meta: {
              title: '待审核',
              group: '任务',
            },
            component: () => import('@/views/project/reviews/index.vue'),
          },
          {
            // 专题（long-task）：长时间自动执行的工程，与日常任务分池
            path: 'topics',
            name: `${routeName}_topics`,
            meta: {
              title: '专题',
              group: '任务',
              hideInMenu: true,
            },
            component: () => import('@/views/project/topics/index.vue'),
          },
          {
            // 跨项目反馈收件箱：对方项目投递的线索，分析后转任务或忽略
            path: 'feedbacks',
            name: `${routeName}_feedbacks`,
            meta: {
              title: '反馈',
              group: '任务',
              hideInMenu: true,
            },
            component: () => import('@/views/project/feedbacks/index.vue'),
          },
          {
            path: 'requirements',
            name: `${routeName}_requirements`,
            meta: {
              title: '需求池',
              group: '规划',
              hideInMenu: true,
            },
            component: () => import('@/views/project/requirements/index.vue'),
          },
          {
            path: 'milestones',
            name: `${routeName}_milestones`,
            meta: {
              title: '里程碑',
              group: '规划',
              hideInMenu: true,
            },
            component: () => import('@/views/project/milestones/index.vue'),
          },
          {
            path: 'sprints',
            name: `${routeName}_sprints`,
            meta: {
              title: '迭代管理',
              group: '规划',
              hideInMenu: true,
            },
            component: () => import('@/views/project/sprints/index.vue'),
          },
          {
            // 记忆/文档中枢：知识库视图（vault 知识库空间 + 人审发布流）
            path: 'knowledge',
            name: `${routeName}_knowledge`,
            meta: {
              title: '知识库',
              group: '知识',
              hideInMenu: true,
              vaultSpace: 'knowledge',
            },
            component: () => import('@/views/project/docs/index.vue'),
          },
          {
            // 记忆/文档中枢：文档管理器（工作区空间）
            path: 'docs',
            name: `${routeName}_docs`,
            meta: {
              title: '文档',
              group: '知识',
              hideInMenu: true,
            },
            component: () => import('@/views/project/docs/index.vue'),
          },
          {
            // QA 库：常见问答沉淀（agent 维护 + 检索）
            path: 'qas',
            name: `${routeName}_qas`,
            meta: {
              title: 'QA 库',
              group: '知识',
              hideInMenu: true,
            },
            component: () => import('@/views/project/qas/index.vue'),
          },
          {
            // 记忆/文档中枢：项目 KV 记忆
            path: 'memories',
            name: `${routeName}_memories`,
            meta: {
              title: '项目记忆',
              group: '知识',
              hideInMenu: true,
            },
            component: () => import('@/views/project/memories/index.vue'),
          },
          {
            // 测试总览：趋势/Flaky/整体质量统计，从执行记录页迁出独立成页
            path: 'test-overview',
            name: `${routeName}_test_overview`,
            meta: {
              title: '测试总览',
              group: '测试',
              hideInMenu: true,
            },
            component: () => import('@/views/test/overview/index.vue'),
          },
          {
            path: 'test-cases',
            name: `${routeName}_test_cases`,
            meta: {
              title: '测试用例',
              group: '测试',
              hideInMenu: true,
            },
            component: () => import('@/views/test/cases/index.vue'),
          },
          {
            path: 'test-plans',
            name: `${routeName}_test_plans`,
            meta: {
              title: '测试计划',
              group: '测试',
              hideInMenu: true,
            },
            component: () => import('@/views/test/plans/index.vue'),
          },
          {
            // 执行记录：pytest/CI 批量上报的落点与逐用例结果
            path: 'test-runs',
            name: `${routeName}_test_runs`,
            meta: {
              title: '执行记录',
              group: '测试',
              hideInMenu: true,
            },
            component: () => import('@/views/test/runs/index.vue'),
          },
          {
            // 执行记录详情：该次执行整体情况，URL 可深链；
            // 无 meta.title——工作台菜单按 title 派生，详情页不进菜单
            path: 'test-runs/:runId',
            name: `${routeName}_test_run_detail`,
            meta: {
              group: '测试',
              hideInMenu: true,
            },
            component: () => import('@/views/test/runs/detail.vue'),
          },
          {
            // 讨论区：论坛式想法/议题线程，成员与 agent 均可参与，
            // 成熟后转任务；不进任务工作流
            path: 'discussions',
            name: `${routeName}_discussions`,
            meta: {
              title: '讨论区',
              group: '讨论',
              hideInMenu: true,
            },
            component: () => import('@/views/project/discussions/index.vue'),
          },
          {
            // 项目发布：本地打包上传、团队内下载安装包
            path: 'releases',
            name: `${routeName}_releases`,
            meta: {
              title: '项目发布',
              group: '管理',
              hideInMenu: true,
            },
            component: () => import('@/views/project/releases/index.vue'),
          },
          {
            // 关联项目 + 项目分组（从成员管理拆出）
            path: 'settings',
            name: `${routeName}_settings`,
            meta: {
              title: '项目设置',
              group: '管理',
              hideInMenu: true,
            },
            component: () => import('@/views/project/settings/index.vue'),
          },
          {
            path: 'members',
            name: `${routeName}_members`,
            meta: {
              title: '成员管理',
              group: '管理',
              hideInMenu: true,
            },
            component: () => import('@/views/project/members/index.vue'),
          },
        ],
      },
      // 兼容旧链接
      {
        path: 'detail/:id',
        redirect: (to) => ({ path: `/project/${to.params.id}/overview` }),
        meta: { hideInMenu: true },
      },
    ],
  },
];

export default routes;
