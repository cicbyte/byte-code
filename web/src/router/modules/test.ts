import { RouteRecordRaw } from 'vue-router';
import { Layout } from '@/router/constant';
import { BugOutlined } from '@vicons/antd';
import { renderIcon } from '@/utils/index';

const routeName = 'test';

const routes: Array<RouteRecordRaw> = [
  {
    path: '/test',
    name: routeName,
    redirect: '/test/cases',
    component: Layout,
    meta: {
      title: '测试管理',
      icon: renderIcon(BugOutlined),
      sort: 4,
    },
    children: [
      {
        path: 'cases',
        name: `${routeName}_cases`,
        meta: {
          title: '测试用例',
        },
        component: () => import('@/views/test/cases.vue'),
      },
      {
        path: 'plans',
        name: `${routeName}_plans`,
        meta: {
          title: '测试计划',
        },
        component: () => import('@/views/test/plans.vue'),
      },
    ],
  },
];

export default routes;
