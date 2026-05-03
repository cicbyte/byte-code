import { RouteRecordRaw } from 'vue-router';
import { Layout } from '@/router/constant';
import { BookOutlined } from '@vicons/antd';
import { renderIcon } from '@/utils/index';

const routeName = 'knowledge';

const routes: Array<RouteRecordRaw> = [
  {
    path: '/knowledge',
    name: routeName,
    redirect: '/knowledge/docs',
    component: Layout,
    meta: {
      title: '知识库',
      icon: renderIcon(BookOutlined),
      sort: 5,
    },
    children: [
      {
        path: 'docs',
        name: `${routeName}_docs`,
        meta: {
          title: '文档',
        },
        component: () => import('@/views/knowledge/docs.vue'),
      },
    ],
  },
];

export default routes;
