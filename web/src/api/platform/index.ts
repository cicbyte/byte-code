import { Alova } from '@/utils/http/alova/index';

// ==================== 类型定义 ====================

/** 标签 */
export interface TagItem {
  id: number;
  name: string;
  color: string;
  creatorId: number;
  createdAt: string;
}

export interface TagListResult {
  list: TagItem[];
}

export interface TagCreateData {
  name: string;
  color: string;
}

export interface TagUpdateData {
  name?: string;
  color?: string;
}

export interface TagAttachData {
  entityType: 'task' | 'requirement' | 'test_case';
  entityId: number;
}

export interface TagEntitiesResult {
  tasks: number[];
  requirements: number[];
  testCases: number[];
}

/** 活动流 */
export interface ActivityItem {
  id: number;
  actorId: number;
  actorType: string;
  actorName: string;
  action: string;
  targetType: string;
  targetId: number;
  targetName: string;
  projectId: number;
  detail: string;
  createdAt: string;
}

export interface ActivityListResult {
  list: ActivityItem[];
  total: number;
}

export interface ActivityListParams {
  module?: string;
  actorId?: number;
  projectId?: number;
  page?: number;
  size?: number;
}

/** 通知 */
export interface NotificationItem {
  id: number;
  title: string;
  content: string;
  type: string;
  isRead: number;
  sourceType: string;
  sourceId: number;
  createdAt: string;
}

export interface NotificationListResult {
  list: NotificationItem[];
  total: number;
}

export interface NotificationListParams {
  unread?: number;
  page?: number;
  size?: number;
}

/** 全局搜索 */
export interface SearchResult {
  module: string;
  id: number;
  projectId: number;
  title: string;
  summary: string;
}

export interface SearchPlatformResult {
  list: SearchResult[];
  total: number;
}

export interface SearchParams {
  q: string;
  module?: string;
  page?: number;
  size?: number;
}

/** 仪表盘统计 */
export interface AiStatItem {
  aiName: string;
  taskCount: number;
}

export interface RecentTaskItem {
  id: number;
  title: string;
  status: string;
  updatedAt: string;
}

export interface DashboardStatsResult {
  totalRequirements: number;
  totalTasks: number;
  inProgressTasks: number;
  reviewTasks: number;
  testPassRate: number;
  aiStats: AiStatItem[];
  recentTasks: RecentTaskItem[];
}

/** 审计日志 */
export interface AuditLogItem {
  id: number;
  actorId: number;
  actorType: string;
  action: string;
  targetType: string;
  targetId: number;
  targetName: string;
  changes: string;
  ipAddress: string;
  projectId: number;
  createdAt: string;
}

export interface AuditLogListResult {
  list: AuditLogItem[];
  total: number;
}

export interface AuditLogListParams {
  targetType?: string;
  actorId?: number;
  action?: string;
  projectId?: number;
  page?: number;
  size?: number;
}

// ==================== 标签 API ====================

/** 标签列表 */
export function getTags() {
  return Alova.Get<TagListResult>('/v1/tags');
}

/** 创建标签 */
export function createTag(data: TagCreateData) {
  return Alova.Post<{ id: number }>('/v1/tags', data);
}

/** 更新标签 */
export function updateTag(id: number, data: TagUpdateData) {
  return Alova.Put(`/v1/tags/${id}`, data);
}

/** 删除标签 */
export function deleteTag(id: number) {
  return Alova.Delete(`/v1/tags/${id}`);
}

/** 给实体打标签 */
export function attachTag(id: number, data: TagAttachData) {
  return Alova.Post(`/v1/tags/${id}/attach`, data);
}

/** 移除实体标签（DELETE 第二参是请求体而非配置，query 需显式拼 URL） */
export function detachTag(id: number, entityType: string, entityId: number) {
  return Alova.Delete(
    `/v1/tags/${id}/detach?entityType=${entityType}&entityId=${entityId}`
  );
}

/** 按标签查询实体 */
export function getTagEntities(id: number) {
  return Alova.Get<TagEntitiesResult>(`/v1/tags/${id}/entities`);
}

// ==================== 活动流 API ====================

/** 活动流列表 */
export function getActivities(params?: ActivityListParams) {
  return Alova.Get<ActivityListResult>('/v1/activities', { params });
}

// ==================== 通知 API ====================

/** 通知列表 */
export function getNotifications(params?: NotificationListParams) {
  return Alova.Get<NotificationListResult>('/v1/notifications', { params });
}

/** 标记已读 */
export function readNotification(id: number) {
  return Alova.Put(`/v1/notifications/${id}/read`);
}

/** 全部已读 */
export function readAllNotifications() {
  return Alova.Put('/v1/notifications/read-all');
}

/** 未读数量 */
export function getUnreadCount() {
  return Alova.Get<{ count: number }>('/v1/notifications/unread-count');
}

// ==================== 搜索 API ====================

/** 全局搜索 */
export function search(params: SearchParams) {
  return Alova.Get<SearchPlatformResult>('/v1/search', { params });
}

// ==================== 仪表盘 API ====================

/** 全局仪表盘统计 */
export function getDashboardStats() {
  return Alova.Get<DashboardStatsResult>('/v1/stats/dashboard');
}

// ==================== 审计日志 API ====================

/** 审计日志列表 */
export function getAuditLogs(params?: AuditLogListParams) {
  return Alova.Get<AuditLogListResult>('/v1/admin/audit-logs', { params });
}
