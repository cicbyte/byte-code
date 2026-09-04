import { Alova } from '@/utils/http/alova/index';

// ==================== 类型定义 ====================

/** 附件 */
export interface AttachmentItem {
  id: number;
  s3Key: string;
  originalName: string;
  fileSize: number;
  mimeType: string;
  fileExt: string;
  entityType: string;
  entityId: number;
  uploaderId: number;
  description: string;
  downloadCount: number;
  createdAt: string;
  uploaderName?: string;
}

export interface AttachmentListResult {
  list: AttachmentItem[];
}

export interface AttachmentUploadData {
  entityType: 'task' | 'requirement' | 'doc' | 'test_case' | 'project';
  entityId: number;
  file: File;
}

/** 存储配置 */
export interface StorageConfig {
  endpoint: string;
  bucket: string;
  region: string;
  useSSL: boolean;
}

export interface StorageConfigUpdateData {
  endpoint: string;
  bucket: string;
  accessKey: string;
  secretKey: string;
  region?: string;
  useSSL?: boolean;
}

export interface StorageTestResult {
  ok: boolean;
  msg?: string;
}

// ==================== 附件 API ====================

/** 上传附件 */
export function uploadAttachment(data: AttachmentUploadData) {
  const formData = new FormData();
  formData.append('entityType', data.entityType);
  formData.append('entityId', String(data.entityId));
  formData.append('file', data.file);
  // 不显式设 Content-Type：浏览器需自动补 boundary，显式设置会导致后端解析失败
  return Alova.Post<{ id: number }>('/v1/attachments/upload', formData);
}

/** 获取附件信息 */
export function getAttachment(id: number) {
  return Alova.Get<AttachmentItem>(`/v1/attachments/${id}`);
}

/** 获取下载 URL */
export function downloadAttachment(id: number) {
  return Alova.Get<{ url: string }>(`/v1/attachments/${id}/download`);
}

/** 获取预览 URL */
export function previewAttachment(id: number) {
  return Alova.Get<{ url: string }>(`/v1/attachments/${id}/preview`);
}

/** 删除附件 */
export function deleteAttachment(id: number) {
  return Alova.Delete(`/v1/attachments/${id}`);
}

/** 实体附件列表 */
export function getAttachments(entityType: string, entityId: number) {
  return Alova.Get<AttachmentListResult>('/v1/attachments', {
    params: { entityType, entityId },
  });
}

/** 更新附件描述 */
export function updateAttachment(id: number, data: { description?: string }) {
  return Alova.Put(`/v1/attachments/${id}`, data);
}

// ==================== 存储配置 API ====================

/** 获取存储配置 */
export function getStorageConfig() {
  return Alova.Get<StorageConfig>('/v1/admin/storage/config');
}

/** 更新存储配置 */
export function updateStorageConfig(data: StorageConfigUpdateData) {
  return Alova.Put('/v1/admin/storage/config', data);
}

/** 测试 S3 连接 */
export function testStorage() {
  return Alova.Post<StorageTestResult>('/v1/admin/storage/test');
}
