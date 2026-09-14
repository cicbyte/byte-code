import { RouteRecordRaw } from 'vue-router';
import { Layout } from '@/router/constant';
import { AuditOutlined } from '@vicons/antd';
import { renderIcon } from '@/utils/index';

const routeName = 'reviews';

const routes: Array<RouteRecordRaw> = [
  {
    path: '/reviews',
    name: routeName,
    redirect: '/reviews/index',
    component: Layout,
    meta: {
      title: '审核中心',
      icon: renderIcon(AuditOutlined),
      sort: 2,
    },
    children: [
      {
        path: 'index',
        name: `${routeName}_index`,
        meta: {
          title: '审核中心',
        },
        component: () => import('@/views/reviews/index.vue'),
      },
    ],
  },
];

export default routes;
