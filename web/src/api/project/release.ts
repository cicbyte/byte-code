import { Alova } from '@/utils/http/alova/index';

/** 项目发布条目（#551/#552：本地打包上传、团队内下载） */
export interface ReleaseItem {
  id: number;
  projectId: number;
  version: string;
  title: string;
  notes: string;
  channel: string;
  createdBy: number;
  createdByName?: string;
  createdAt: string;
  updatedAt: string;
  fileCount: number;
  totalSizeBytes: number;
}

export interface ReleaseListResult {
  list: ReleaseItem[];
  total: number;
  currentPage: number;
}

export interface ReleaseListParams {
  channel?: string;
  pageNum?: number;
  pageSize?: number;
}

export interface ReleaseCreateData {
  version: string;
  title?: string;
  notes?: string;
  channel?: string;
}

export interface ReleaseUpdateData {
  title?: string;
  notes?: string;
  channel?: string;
}

/** 发布文件（#552 独立体系：按 项目/版本 组织、支持分享直链） */
export interface ReleaseFileItem {
  id: number;
  releaseId: number;
  fileName: string;
  fileSize: number;
  mimeType: string;
  uploaderId: number;
  uploaderName?: string;
  downloadCount: number;
  shared: boolean;
  shareExpiresAt?: string;
  createdAt: string;
}

export interface ReleaseShareResult {
  url: string;
  expiresAt: string;
}

/** 发布列表 */
export function getReleases(projectId: number, params?: ReleaseListParams) {
  return Alova.Get<ReleaseListResult>(`/v1/projects/${projectId}/releases`, { params });
}

/** 创建发布 */
export function createRelease(projectId: number, data: ReleaseCreateData) {
  return Alova.Post<{ id: number }>(`/v1/projects/${projectId}/releases`, data);
}

/** 更新发布（标题/说明/渠道） */
export function updateRelease(id: number, data: ReleaseUpdateData) {
  return Alova.Put(`/v1/releases/${id}`, data);
}

/** 删除发布（级联删文件） */
export function deleteRelease(id: number) {
  return Alova.Delete(`/v1/releases/${id}`);
}

/** 发布文件列表 */
export function getReleaseFiles(releaseId: number) {
  return Alova.Get<{ list: ReleaseFileItem[] }>(`/v1/releases/${releaseId}/files`);
}

/** 上传发布文件（multipart，单文件上限 2GB；fetch 无默认超时，大包可传） */
export function uploadReleaseFile(releaseId: number, file: File) {
  const formData = new FormData();
  formData.append('file', file);
  return Alova.Post<{ id: number }>(`/v1/releases/${releaseId}/files`, formData);
}

/** 删除发布文件（含存储对象） */
export function deleteReleaseFile(id: number) {
  return Alova.Delete(`/v1/release-files/${id}`);
}

/** 生成/获取分享直链（幂等；days>0 重新生成并设有效期，0=永久） */
export function shareReleaseFile(id: number, days = 0) {
  return Alova.Post<ReleaseShareResult>(`/v1/release-files/${id}/share`, { days });
}

/** 吊销分享直链 */
export function revokeReleaseFileShare(id: number) {
  return Alova.Delete(`/v1/release-files/${id}/share`);
}
