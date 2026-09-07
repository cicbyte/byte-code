import { RouteRecordRaw } from 'vue-router';
import { Layout } from '@/router/constant';
import { SettingOutlined } from '@vicons/antd';
import { renderIcon } from '@/utils/index';

const routes: Array<RouteRecordRaw> = [
  {
    path: '/setting',
    name: 'Setting',
    redirect: '/setting/account',
    component: Layout,
    meta: {
      title: '设置页面',
      icon: renderIcon(SettingOutlined),
      sort: 5,
    },
    children: [
      {
        path: 'account',
        name: 'setting-account',
        meta: {
          title: '个人设置',
        },
        component: () => import('@/views/setting/account/account.vue'),
      },
      {
        path: 'system',
        name: 'setting-system',
        meta: {
          title: '系统设置',
          // 系统配置接口在管理组（MiddlewareAdminAuth），菜单同步设门槛
          menuKey: 'setting_system',
        },
        component: () => import('@/views/setting/system/system.vue'),
      },
      {
        path: 'global-memory',
        name: 'setting-global-memory',
        meta: {
          title: '全局记忆',
          // 记忆写操作接口在管理组，读全员；页面含管理动作，按管理入口设门槛
          menuKey: 'setting_global_memory',
        },
        component: () => import('@/views/setting/memory/global-memory.vue'),
      },
    ],
  },
];

export default routes;
