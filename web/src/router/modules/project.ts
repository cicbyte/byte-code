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
            path: 'sprints',
            name: `${routeName}_sprints`,
            meta: {
              title: 'Sprint 管理',
              hideInMenu: true,
            },
            component: () => import('@/views/project/sprints.vue'),
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
