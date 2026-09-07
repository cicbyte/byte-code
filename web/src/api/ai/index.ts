import { Alova } from '@/utils/http/alova/index';

// ==================== 类型定义 ====================

/** Agent 账号 */
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

// ==================== Agent 账号 API ====================

/** Agent 账号列表 */
export function getAiUsers(params?: AiUserListParams) {
  return Alova.Get<AiUserListResult>('/v1/ai-users', { params });
}

/** 创建 Agent */
export function createAiUser(data: AiUserCreateData) {
  return Alova.Post<{ id: number; apiKey: string }>('/v1/ai-users', data);
}

/** 更新 Agent */
export function updateAiUser(id: number, data: AiUserUpdateData) {
  return Alova.Put(`/v1/ai-users/${id}`, data);
}

/** 删除 Agent */
export function deleteAiUser(id: number) {
  return Alova.Delete(`/v1/ai-users/${id}`);
}

/** 重置 API Key */
export function resetAiUserKey(id: number) {
  return Alova.Post<{ apiKey: string }>(`/v1/ai-users/${id}/reset-key`);
}

// ==================== Agent 认证 API ====================

