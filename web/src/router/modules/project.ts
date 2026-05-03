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
        path: 'detail/:id',
        name: `${routeName}_detail`,
        meta: {
          title: '项目详情',
          hideInMenu: true,
        },
        component: () => import('@/views/project/detail.vue'),
      },
      {
        path: 'sprints',
        name: `${routeName}_sprints`,
        meta: {
          title: 'Sprint 管理',
        },
        component: () => import('@/views/project/sprints.vue'),
      },
    ],
  },
];

export default routes;
