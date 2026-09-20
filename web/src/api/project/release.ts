import { Alova } from '@/utils/http/alova/index';

/** 项目发布条目（#551：本地打包上传、团队内下载） */
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

/** 删除发布（级联删附件记录） */
export function deleteRelease(id: number) {
  return Alova.Delete(`/v1/releases/${id}`);
}
