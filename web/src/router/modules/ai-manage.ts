import { RouteRecordRaw } from 'vue-router';
import { Layout } from '@/router/constant';
import { AgentSparkleIcon } from '@/components/Icons/AgentSparkle';
import { renderIcon } from '@/utils/index';

const routeName = 'ai-manage';

const routes: Array<RouteRecordRaw> = [
  {
    path: '/ai-manage',
    name: routeName,
    redirect: '/ai-manage/users',
    component: Layout,
    meta: {
      menuKey: "ai_users",
      title: 'Agent 管理',
      icon: renderIcon(AgentSparkleIcon),
      sort: 6,
    },
    children: [
      {
        path: 'users',
        name: `${routeName}_users`,
        meta: {
          title: 'Agent 账号',
        },
        component: () => import('@/views/ai-manage/users.vue'),
      },
    ],
  },
];

export default routes;
