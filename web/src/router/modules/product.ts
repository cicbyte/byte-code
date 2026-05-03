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
        path: ':productId',
        name: `${routeName}_workspace`,
        meta: {
          title: '产品详情',
          hideInMenu: true,
        },
        component: () => import('@/views/product/ProductWorkspace.vue'),
        children: [
          {
            path: 'overview',
            name: `${routeName}_overview`,
            meta: {
              title: '产品概览',
              hideInMenu: true,
            },
            component: () => import('@/views/product/overview.vue'),
          },
          {
            path: 'requirements',
            name: `${routeName}_requirements`,
            meta: {
              title: '需求池',
              hideInMenu: true,
            },
            component: () => import('@/views/product/requirements.vue'),
          },
          {
            path: 'milestones',
            name: `${routeName}_milestones`,
            meta: {
              title: '里程碑',
              hideInMenu: true,
            },
            component: () => import('@/views/product/milestones.vue'),
          },
        ],
      },
    ],
  },
];

export default routes;
