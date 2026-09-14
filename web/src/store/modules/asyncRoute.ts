import { toRaw } from 'vue';
import { defineStore } from 'pinia';
import { RouteRecordRaw } from 'vue-router';
import { store } from '@/store';
import { usePerm } from '@/composables/usePerm';
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
      const { perms, isAdmin } = usePerm();
      // 超管（拥有 system_menu 或 system_role 权限）看全部；
      // 普通用户按 meta.menuKey 过滤（顶级与子级同规则）；无 menuKey 的路由始终可见
      const hasPerm = (meta: any) => {
        const key = meta?.menuKey as string | undefined;
        return !key || perms.value.has(key);
      };
      const visible = isAdmin.value
        ? asyncRoutes
        : asyncRoutes
            .filter((route) => hasPerm(route.meta))
            .map((route) => {
              if (!route.children?.length) return route;
              const children = route.children.filter((c) => hasPerm(c.meta));
              return { ...route, children };
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
