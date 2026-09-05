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
              hideInMenu: true,
            },
            component: () => import('@/views/project/board.vue'),
          },
          {
            path: 'tasks',
            name: `${routeName}_tasks`,
            meta: {
              title: '任务列表',
              hideInMenu: true,
            },
            component: () => import('@/views/project/tasks.vue'),
          },
          {
            path: 'requirements',
            name: `${routeName}_requirements`,
            meta: {
              title: '需求池',
              hideInMenu: true,
            },
            component: () => import('@/views/project/requirements.vue'),
          },
          {
            path: 'milestones',
            name: `${routeName}_milestones`,
            meta: {
              title: '里程碑',
              hideInMenu: true,
            },
            component: () => import('@/views/project/milestones.vue'),
          },
          {
            path: 'sprints',
            name: `${routeName}_sprints`,
            meta: {
              title: '迭代管理',
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
              hideInMenu: true,
            },
            component: () => import('@/views/project/memories.vue'),
          },
          {
            path: 'test-cases',
            name: `${routeName}_test_cases`,
            meta: {
              title: '测试用例',
              hideInMenu: true,
            },
            component: () => import('@/views/test/cases.vue'),
          },
          {
            path: 'test-plans',
            name: `${routeName}_test_plans`,
            meta: {
              title: '测试计划',
              hideInMenu: true,
            },
            component: () => import('@/views/test/plans.vue'),
          },
          {
            path: 'members',
            name: `${routeName}_members`,
            meta: {
              title: '成员管理',
              hideInMenu: true,
            },
            component: () => import('@/views/project/members.vue'),
          },
          {
            path: 'database',
            name: `${routeName}_database`,
            meta: {
              title: '数据库模型',
              hideInMenu: true,
            },
            component: () => import('@/views/project/database.vue'),
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
