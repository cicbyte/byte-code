import { RouteRecordRaw } from 'vue-router';
import { Layout } from '@/router/constant';
import { ShieldCheckmarkOutline } from '@vicons/ionicons5';
import { renderIcon } from '@/utils/index';

const routeName = 'permission';

const routes: Array<RouteRecordRaw> = [
  {
    path: '/permission',
    name: routeName,
    redirect: '/permission/role',
    component: Layout,
    meta: {
      menuKey: "system_role",
      title: '权限管理',
      icon: renderIcon(ShieldCheckmarkOutline),
      sort: 8,
    },
    children: [
      {
        path: 'role',
        name: `${routeName}_role`,
        meta: {
          title: '角色权限',
        },
        component: () => import('@/views/system/role/role.vue'),
      },
    ],
  },
];

export default routes;
