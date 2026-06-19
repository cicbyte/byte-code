import { Alova } from '@/utils/http/alova/index';

// ==================== 类型定义 ====================

/** 产品 */
export interface ProductItem {
  id: number;
  name: string;
  description: string;
  ownerId: number;
  ownerName: string;
  status: string;
  createdAt: string;
  updatedAt: string;
}

export interface ProductListResult {
  list: ProductItem[];
  total: number;
  page: number;
  size: number;
}

export interface ProductListParams {
  status?: string;
  keyword?: string;
  page?: number;
  size?: number;
}

export interface ProductCreateData {
  name: string;
  description?: string;
}

export interface ProductUpdateData {
  name?: string;
  description?: string;
  status?: string;
}

/** 需求 */
export interface RequirementItem {
  id: number;
  productId: number;
  parentId: number;
  type: string;
  title: string;
  description: string;
  status: string;
  priority: number;
  assigneeId: number;
  assigneeName: string;
  creatorId: number;
  creatorName: string;
  milestoneId: number;
  sortOrder: number;
  acceptanceCriteria: string;
  source: string;
  children?: RequirementItem[];
  createdAt: string;
  updatedAt: string;
}

export interface RequirementListResult {
  list: RequirementItem[];
  total: number;
}

export interface RequirementListParams {
  type?: string;
  status?: string;
  parentId?: number;
  page?: number;
  size?: number;
}

export interface RequirementCreateData {
  parentId?: number;
  type?: string;
  title: string;
  description?: string;
  priority?: number;
  assigneeId?: number;
  milestoneId?: number;
  acceptanceCriteria?: string;
}

export interface RequirementUpdateData {
  title?: string;
  description?: string;
  status?: string;
  priority?: number;
  assigneeId?: number;
  milestoneId?: number;
  acceptanceCriteria?: string;
  sortOrder?: number;
}

/** 里程碑 */
export interface MilestoneItem {
  id: number;
  productId: number;
  name: string;
  description: string;
  targetDate: string;
  status: string;
  createdAt: string;
}

export interface MilestoneListResult {
  list: MilestoneItem[];
}

export interface MilestoneCreateData {
  name: string;
  description?: string;
  targetDate?: string;
}

// ==================== 产品 API ====================

/** 产品列表 */
export function getProducts(params?: ProductListParams) {
  return Alova.Get<ProductListResult>('/v1/products', { params });
}

/** 创建产品 */
export function createProduct(data: ProductCreateData) {
  return Alova.Post<{ id: number }>('/v1/products', data);
}

/** 产品详情 */
export function getProduct(id: number) {
  return Alova.Get<ProductItem>(`/v1/products/${id}`);
}

/** 更新产品 */
export function updateProduct(id: number, data: ProductUpdateData) {
  return Alova.Put(`/v1/products/${id}`, data);
}

/** 删除产品 */
export function deleteProduct(id: number) {
  return Alova.Delete(`/v1/products/${id}`);
}

// ==================== 需求 API ====================

/** 需求列表 */
export function getRequirements(productId: number, params?: RequirementListParams) {
  return Alova.Get<RequirementListResult>(`/v1/products/${productId}/requirements`, { params });
}

/** 创建需求 */
export function createRequirement(productId: number, data: RequirementCreateData) {
  return Alova.Post<{ id: number }>(`/v1/products/${productId}/requirements`, data);
}

/** 需求详情 */
export function getRequirement(id: number) {
  return Alova.Get<RequirementItem>(`/v1/requirements/${id}`);
}

/** 更新需求 */
export function updateRequirement(id: number, data: RequirementUpdateData) {
  return Alova.Put(`/v1/requirements/${id}`, data);
}

/** 删除需求 */
export function deleteRequirement(id: number) {
  return Alova.Delete(`/v1/requirements/${id}`);
}

// ==================== 里程碑 API ====================

/** 里程碑列表 */
export function getMilestones(productId: number) {
  return Alova.Get<MilestoneListResult>(`/v1/products/${productId}/milestones`);
}

/** 创建里程碑 */
export function createMilestone(productId: number, data: MilestoneCreateData) {
  return Alova.Post<{ id: number }>(`/v1/products/${productId}/milestones`, data);
}
