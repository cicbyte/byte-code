import { RouteRecordRaw } from 'vue-router';
import { Layout } from '@/router/constant';
import { ControlOutlined } from '@vicons/antd';
import { renderIcon } from '@/utils/index';

const routeName = 'platform';

const routes: Array<RouteRecordRaw> = [
  {
    path: '/platform',
    name: routeName,
    redirect: '/platform/activities',
    component: Layout,
    meta: {
      menuKey: "platform_activities",
      title: '平台',
      icon: renderIcon(ControlOutlined),
      sort: 7,
    },
    children: [
      {
        path: 'activities',
        name: `${routeName}_activities`,
        meta: {
          title: '活动流',
        },
        component: () => import('@/views/platform/activities.vue'),
      },
      {
        path: 'tags',
        name: `${routeName}_tags`,
        meta: {
          title: '标签管理',
        },
        component: () => import('@/views/platform/tags.vue'),
      },
      {
        path: 'audit',
        name: `${routeName}_audit`,
        meta: {
          title: '审计日志',
        },
        component: () => import('@/views/platform/audit.vue'),
      },
    ],
  },
];

export default routes;
