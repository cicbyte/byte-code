import { computed } from 'vue';
import { useUserStore } from '@/store/modules/user';

/**
 * 权限判断统一入口（PRD design/permission-system-prd.md §3.4）。
 *
 * 数据源是后端 admin_info 下发的 permissions（sys_menus.name 权限字典），
 * 与路由过滤（asyncRoute）同一口径；页面上不要再手写 Set + has 判断。
 *
 * - isAdmin：管理口径 = 持有 system_menu 或 system_role（字典中的管理组项）
 * - has(key)：单项权限，如 has('platform_tags') / has('project_create')
 */
export function usePerm() {
  const userStore = useUserStore();
  const perms = computed<Set<string>>(
    () => new Set((userStore.permissions || []).map((p: any) => p?.value || p))
  );
  const isAdmin = computed(() => perms.value.has('system_menu') || perms.value.has('system_role'));
  const has = (key: string) => perms.value.has(key);
  return { perms, isAdmin, has };
}
