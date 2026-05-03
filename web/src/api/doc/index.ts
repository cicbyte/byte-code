import { Alova } from '@/utils/http/alova/index';

// ==================== 类型定义 ====================

/** 文档 */
export interface DocItem {
  id: number;
  projectId: number;
  parentId: number;
  title: string;
  content?: string;
  type: string;
  sortOrder: number;
  creatorId: number;
  lastEditorId: number;
  version: number;
  status: string;
  source: string;
  createdAt: string;
  updatedAt: string;
  creatorName?: string;
  editorName?: string;
}

export interface DocListResult {
  list: DocItem[];
  total: number;
  currentPage: number;
}

export interface DocListParams {
  projectId?: number;
  parentId?: number;
  type?: string;
  status?: string;
  keyword?: string;
  pageNum?: number;
  pageSize?: number;
}

export interface DocCreateData {
  projectId: number;
  parentId?: number;
  title: string;
  content?: string;
  type?: string;
  sortOrder?: number;
}

export interface DocUpdateData {
  parentId?: number;
  title?: string;
  content?: string;
  type?: string;
  sortOrder?: number;
  status?: string;
}

/** 文档树 */
export interface DocTreeNode {
  id: number;
  parentId: number;
  title: string;
  type: string;
  sortOrder: number;
  children?: DocTreeNode[];
}

export interface DocTreeResult {
  tree: DocTreeNode[];
}

/** 文档版本 */
export interface DocVersionItem {
  id: number;
  docId: number;
  version: number;
  content?: string;
  editorId: number;
  editorName?: string;
  changeSummary: string;
  createdAt: string;
}

export interface DocVersionListResult {
  list: DocVersionItem[];
  total: number;
  currentPage: number;
}

/** 文档关联 */
export interface DocRelationItem {
  id: number;
  docId: number;
  targetType: string;
  targetId: number;
  targetTitle?: string;
}

export interface DocRelationListResult {
  list: DocRelationItem[];
}

export interface DocRelationCreateData {
  targetType: 'task' | 'requirement' | 'test_case' | 'doc';
  targetId: number;
}

/** 全局搜索 */
export interface SearchResult {
  module: string;
  id: number;
  title: string;
  summary: string;
}

export interface SearchDocResult {
  list: SearchResult[];
  total: number;
  currentPage: number;
}

export interface SearchDocParams {
  keyword: string;
  pageNum?: number;
  pageSize?: number;
}

// ==================== 文档 CRUD API ====================

/** 文档列表 */
export function getDocs(params?: DocListParams) {
  return Alova.Get<DocListResult>('/v1/docs', { params });
}

/** 创建文档 */
export function createDoc(data: DocCreateData) {
  return Alova.Post<{ id: number }>('/v1/docs', data);
}

/** 文档详情 */
export function getDoc(id: number) {
  return Alova.Get<DocItem>(`/v1/docs/${id}`);
}

/** 更新文档 */
export function updateDoc(id: number, data: DocUpdateData) {
  return Alova.Put(`/v1/docs/${id}`, data);
}

/** 删除文档 */
export function deleteDoc(id: number) {
  return Alova.Delete(`/v1/docs/${id}`);
}

// ==================== 文档移动 API ====================

/** 移动文档 */
export function moveDoc(id: number, parentId: number) {
  return Alova.Put(`/v1/docs/${id}/move`, { parentId });
}

// ==================== 文档树 API ====================

/** 项目文档树 */
export function getDocTree(projectId: number) {
  return Alova.Get<DocTreeResult>(`/v1/projects/${projectId}/docs/tree`);
}

// ==================== 文档版本 API ====================

/** 版本历史 */
export function getDocVersions(id: number, params?: { pageNum?: number; pageSize?: number }) {
  return Alova.Get<DocVersionListResult>(`/v1/docs/${id}/versions`, { params });
}

/** 回退版本 */
export function revertDoc(id: number, data: { version: number; summary?: string }) {
  return Alova.Post(`/v1/docs/${id}/revert`, data);
}

// ==================== 文档关联 API ====================

/** 关联列表 */
export function getDocRelations(id: number) {
  return Alova.Get<DocRelationListResult>(`/v1/docs/${id}/relations`);
}

/** 创建关联 */
export function createDocRelation(id: number, data: DocRelationCreateData) {
  return Alova.Post<{ id: number }>(`/v1/docs/${id}/relations`, data);
}

/** 删除关联 */
export function deleteDocRelation(id: number, targetType: string, targetId: number) {
  return Alova.Delete(`/v1/docs/${id}/relations/${targetType}/${targetId}`);
}

// ==================== 搜索 API ====================

/** 全局搜索 */
export function searchDocs(params: SearchDocParams) {
  return Alova.Get<SearchDocResult>('/v1/search', { params });
}
