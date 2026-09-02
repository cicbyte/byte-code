import { toRaw } from 'vue';
import { defineStore } from 'pinia';
import { RouteRecordRaw } from 'vue-router';
import { store } from '@/store';
import { useUserStore } from '@/store/modules/user';
import { asyncRoutes, constantRouter } from '@/router/index';

export interface IAsyncRouteState {
  menus: RouteRecordRaw[];
  routers: any[];
  routersAdded: any[];
  keepAliveComponents: string[];
  isDynamicRouteAdded: boolean;
}

export const useAsyncRouteStore = defineStore({
  id: 'app-async-route',
  state: (): IAsyncRouteState => ({
    menus: [],
    routers: constantRouter,
    routersAdded: [],
    keepAliveComponents: [],
    isDynamicRouteAdded: false,
  }),
  getters: {
    getMenus(): RouteRecordRaw[] {
      return this.menus;
    },
    getIsDynamicRouteAdded(): boolean {
      return this.isDynamicRouteAdded;
    },
  },
  actions: {
    getRouters() {
      return toRaw(this.routersAdded);
    },
    setDynamicRouteAdded(added: boolean) {
      this.isDynamicRouteAdded = added;
    },
    setRouters(routers: RouteRecordRaw[]) {
      this.routersAdded = routers;
      this.routers = constantRouter.concat(routers);
    },
    setMenus(menus: RouteRecordRaw[]) {
      this.menus = menus;
    },
    setKeepAliveComponents(compNames: string[]) {
      this.keepAliveComponents = compNames;
    },
    async generateRoutes() {
      const userStore = useUserStore();
      const perms = new Set(
        (userStore.permissions || []).map((p: any) => p?.value || p)
      );
      // 超管（拥有 system_menu 或 system_role 权限）看全部；
      // 普通用户按 meta.menuKey 过滤；无 menuKey 的路由（项目工作台等）始终可见
      const isAdmin = perms.has('system_menu') || perms.has('system_role');
      const visible = isAdmin
        ? asyncRoutes
        : asyncRoutes.filter((route) => {
            const key = route.meta?.menuKey as string | undefined;
            return !key || perms.has(key);
          });
      this.setRouters(visible);
      this.setMenus(visible);
      return toRaw(visible);
    },
  },
});

export function useAsyncRoute() {
  return useAsyncRouteStore(store);
}
