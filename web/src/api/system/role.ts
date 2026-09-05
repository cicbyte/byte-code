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

