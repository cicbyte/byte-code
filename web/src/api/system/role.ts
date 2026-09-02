import { Alova } from '@/utils/http/alova/index';

/**
 * @description: 角色列表
 */
export function getRoleList(params) {
  return Alova.Get('/role/list', { params });
}


export interface RoleCreateData {
  name: string;
  explain?: string;
  menuIds: number[];
}

export interface RoleUpdateData {
  name?: string;
  explain?: string;
  status?: string;
  menuIds?: number[];
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
