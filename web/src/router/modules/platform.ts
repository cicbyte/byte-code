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
      // 通知中心/活动流均为个人可见数据（API 全员开放），平台组不设权限门槛
      title: '平台',
      icon: renderIcon(ControlOutlined),
      sort: 7,
    },
    children: [
      {
        path: 'notifications',
        name: `${routeName}_notifications`,
        meta: {
          title: '通知中心',
        },
        component: () => import('@/views/platform/notifications.vue'),
      },
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
          // 审计接口在管理组（MiddlewareAdminAuth），菜单同步设门槛
          menuKey: 'platform_audit',
        },
        component: () => import('@/views/platform/audit.vue'),
      },
      {
        // Agent 账号（bc_ Key 外部代理）：平台级管理入口，接口在管理组
        path: 'agents',
        name: `${routeName}_agents`,
        meta: {
          title: 'Agent 账号',
          menuKey: 'ai_users',
        },
        component: () => import('@/views/platform/agents.vue'),
      },
    ],
  },
  // 兼容旧地址 /ai-manage/users -> /platform/agents
  {
    path: '/ai-manage/:pathMatch(.*)',
    redirect: '/platform/agents',
    meta: { hideInMenu: true, title: 'Agent 账号' },
  },
];

export default routes;
