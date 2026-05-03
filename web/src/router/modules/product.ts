import { RouteRecordRaw } from 'vue-router';
import { Layout } from '@/router/constant';
import { AppstoreOutlined } from '@vicons/antd';
import { renderIcon } from '@/utils/index';

const routeName = 'product';

const routes: Array<RouteRecordRaw> = [
  {
    path: '/product',
    name: routeName,
    redirect: '/product/list',
    component: Layout,
    meta: {
      title: '产品管理',
      icon: renderIcon(AppstoreOutlined),
      sort: 2,
    },
    children: [
      {
        path: 'list',
        name: `${routeName}_list`,
        meta: {
          title: '产品列表',
        },
        component: () => import('@/views/product/list.vue'),
      },
      {
        path: 'requirements',
        name: `${routeName}_requirements`,
        meta: {
          title: '需求池',
        },
        component: () => import('@/views/product/requirements.vue'),
      },
    ],
  },
];

export default routes;
