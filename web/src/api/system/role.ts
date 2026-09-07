import { Alova } from '@/utils/http/alova/index';

// ==================== 类型定义 ====================

export interface RoleItem {
  id: string;
  name: string;
  explain: string;
  isDefault: boolean;
  /** 后端序列化为 snake_case，勿写成 menuIds */
  menu_ids?: number[];
  menu_keys?: string[];
  createDate?: string;
  create_date?: string;
  status: string;
}

export interface RoleListResult {
  page: number;
  pageSize: number;
  pageCount: number;
  list: RoleItem[];
}

export interface RoleCreateData {
  name: string;
  explain?: string;
  menuIds?: number[];
}

export interface RoleUpdateData {
  name?: string;
  explain?: string;
  status?: string;
  menuIds?: number[];
}

/** sys_menus 字典树节点（角色权限分配用，key 为菜单名、id 为存储主键） */
export interface MenuDictItem {
  id: number;
  label: string;
  key: string;
  children?: MenuDictItem[];
}

// ==================== API ====================

export function getRoleList(params?: { pageNum?: number; pageSize?: number; name?: string }) {
  return Alova.Get<RoleListResult>('/role/list', { params });
}

export function createRole(data: RoleCreateData) {
  return Alova.Post<{ id: number }>('/role', data);
}

export function updateRole(id: number, data: RoleUpdateData) {
  return Alova.Put(`/role/${id}`, data);
}

export function deleteRole(id: number) {
  return Alova.Delete(`/role/${id}`);
}

export function updateRoleMenus(id: number, menuIds: number[]) {
  return Alova.Put(`/role/${id}/menus`, { menuIds });
}

export function getMenuList() {
  return Alova.Get<{ list: MenuDictItem[] }>('/menu/list');
}
