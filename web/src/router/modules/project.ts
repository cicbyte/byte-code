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
        component: () => import('@/views/project/list.vue'),
      },
      {
        path: ':projectId',
        name: `${routeName}_workspace`,
        meta: {
          title: '项目详情',
          hideInMenu: true,
        },
        component: () => import('@/views/project/ProjectWorkspace.vue'),
        children: [
          {
            path: 'overview',
            name: `${routeName}_overview`,
            meta: {
              title: '项目概览',
              hideInMenu: true,
            },
            component: () => import('@/views/project/overview.vue'),
          },
          {
            path: 'board',
            name: `${routeName}_board`,
            meta: {
              title: '任务看板',
              group: '任务',
              hideInMenu: true,
            },
            component: () => import('@/views/project/board.vue'),
          },
          {
            path: 'tasks',
            name: `${routeName}_tasks`,
            meta: {
              title: '任务列表',
              group: '任务',
              hideInMenu: true,
            },
            component: () => import('@/views/project/tasks.vue'),
          },
          {
            // 审核工作台：集中验收（完成即待人审，CompleteTask 统一置位）
            path: 'reviews',
            name: `${routeName}_reviews`,
            meta: {
              title: '待审核',
              group: '任务',
            },
            component: () => import('@/views/project/reviews.vue'),
          },
          {
            path: 'requirements',
            name: `${routeName}_requirements`,
            meta: {
              title: '需求池',
              group: '规划',
              hideInMenu: true,
            },
            component: () => import('@/views/project/requirements.vue'),
          },
          {
            path: 'milestones',
            name: `${routeName}_milestones`,
            meta: {
              title: '里程碑',
              group: '规划',
              hideInMenu: true,
            },
            component: () => import('@/views/project/milestones.vue'),
          },
          {
            path: 'sprints',
            name: `${routeName}_sprints`,
            meta: {
              title: '迭代管理',
              group: '规划',
              hideInMenu: true,
            },
            component: () => import('@/views/project/sprints.vue'),
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
            component: () => import('@/views/project/docs.vue'),
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
            component: () => import('@/views/project/docs.vue'),
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
            component: () => import('@/views/project/memories.vue'),
          },
          {
            path: 'test-cases',
            name: `${routeName}_test_cases`,
            meta: {
              title: '测试用例',
              group: '测试',
              hideInMenu: true,
            },
            component: () => import('@/views/test/cases.vue'),
          },
          {
            path: 'test-plans',
            name: `${routeName}_test_plans`,
            meta: {
              title: '测试计划',
              group: '测试',
              hideInMenu: true,
            },
            component: () => import('@/views/test/plans.vue'),
          },
          {
            path: 'members',
            name: `${routeName}_members`,
            meta: {
              title: '成员管理',
              group: '管理',
              hideInMenu: true,
            },
            component: () => import('@/views/project/members.vue'),
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
