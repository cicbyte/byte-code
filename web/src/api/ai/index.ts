import { Alova } from '@/utils/http/alova/index';

// ==================== 类型定义 ====================

/** AI 用户 */
export interface AiUserItem {
  id: number;
  username: string;
  realName: string;
  avatar: string;
  capabilities: string;
  status: number;
  ownerHumanId: number;
  createdAt: string;
}

export interface AiUserListResult {
  list: AiUserItem[];
  total: number;
}

export interface AiUserListParams {
  page?: number;
  size?: number;
}

export interface AiUserCreateData {
  username: string;
  realName?: string;
  capabilities?: string;
}

export interface AiUserUpdateData {
  realName?: string;
  capabilities?: string;
  status?: number;
}

export interface AiLoginData {
  apiKey: string;
}

// ==================== AI 用户 API ====================

/** AI 用户列表 */
export function getAiUsers(params?: AiUserListParams) {
  return Alova.Get<AiUserListResult>('/v1/ai-users', { params });
}

/** 创建 AI 用户 */
export function createAiUser(data: AiUserCreateData) {
  return Alova.Post<{ id: number; apiKey: string }>('/v1/ai-users', data);
}

/** 更新 AI 用户 */
export function updateAiUser(id: number, data: AiUserUpdateData) {
  return Alova.Put(`/v1/ai-users/${id}`, data);
}

/** 删除 AI 用户 */
export function deleteAiUser(id: number) {
  return Alova.Delete(`/v1/ai-users/${id}`);
}

/** 重置 API Key */
export function resetAiUserKey(id: number) {
  return Alova.Post<{ apiKey: string }>(`/v1/ai-users/${id}/reset-key`);
}

// ==================== AI 认证 API ====================

/** AI 用户登录 */
export function aiLogin(data: AiLoginData) {
  return Alova.Post<{ token: string }>('/auth/ai/login', data);
}
