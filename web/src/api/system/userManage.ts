import { Alova } from '@/utils/http/alova/index';

// ==================== 类型定义 ====================

export interface UserItem {
  id: number;
  username: string;
  realName: string;
  email: string;
  type: string;
  status: number;
  createdAt: string;
}

export interface UserListResult {
  list: UserItem[];
  total: number;
  page: number;
  size: number;
}

export interface UserCreateData {
  username: string;
  password: string;
  realName?: string;
  email?: string;
  roleIds?: number[];
}

export interface UserUpdateData {
  realName?: string;
  email?: string;
  status?: number;
  roleIds?: number[];
}

// ==================== API ====================

export function getUserList(params?: { keyword?: string; page?: number; size?: number }) {
  return Alova.Get<UserListResult>('/v1/admin/users', { params });
}

export function createUser(data: UserCreateData) {
  return Alova.Post<{ id: number }>('/v1/admin/users', data);
}

export function updateUser(id: number, data: UserUpdateData) {
  return Alova.Put(`/v1/admin/users/${id}`, data);
}

export function resetUserPassword(id: number, newPassword: string) {
  return Alova.Put(`/v1/admin/users/${id}/reset-password`, { newPassword });
}

export function deleteUser(id: number) {
  return Alova.Delete(`/v1/admin/users/${id}`);
}
