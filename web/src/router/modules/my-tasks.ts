import { RouteRecordRaw } from 'vue-router';
import { Layout } from '@/router/constant';
import { UnorderedListOutlined } from '@vicons/antd';
import { renderIcon } from '@/utils/index';

const routeName = 'my-tasks';

const routes: Array<RouteRecordRaw> = [
  {
    path: '/my-tasks',
    name: routeName,
    redirect: '/my-tasks/index',
    component: Layout,
    meta: {
      title: '我的任务',
      icon: renderIcon(UnorderedListOutlined),
      sort: 1,
    },
    children: [
      {
        path: 'index',
        name: `${routeName}_index`,
        meta: {
          title: '我的任务',
        },
        component: () => import('@/views/my-tasks/index.vue'),
      },
    ],
  },
];

export default routes;
