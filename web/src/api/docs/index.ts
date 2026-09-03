import { Alova } from '@/utils/http/alova/index';

// ==================== 类型定义 ====================

/** 文件元数据（frontmatter） */
export interface FileMeta {
  title: string;
  space: string;
  type: string;
  status: string;
  tags: string[] | null;
  linked: string[] | null;
}

/** 目录树节点 */
export interface DocsTreeNode {
  name: string;
  path: string;
  isDir: boolean;
  size: number;
  modTime: string;
  meta?: FileMeta | null;
  children?: DocsTreeNode[];
}

export interface DocsTreeResult {
  tree: DocsTreeNode[];
}

/** 文件读取 */
export interface DocsFile {
  path: string;
  binary: boolean;
  content: string;
  meta?: FileMeta | null;
  size: number;
}

/** 写入结果 */
export interface DocsWriteResult {
  path: string;
  size: number;
}

/** 补丁操作 */
export interface DocsPatchOperation {
  type: 'replace' | 'append' | 'prepend';
  search?: string;
  replace?: string;
  content?: string;
}

export interface DocsPatchResult {
  applied: number;
  skipped: number;
  items: Array<{ index: number; type: string; applied: boolean; reason?: string }>;
}

/** 搜索/反查条目 */
export interface DocsSearchItem {
  path: string;
  title: string;
  space: string;
  type: string;
  tags: string[] | null;
}

/** 刷新结果 */
export interface DocsRefreshResult {
  changed: number;
  deleted: number;
}

/** 记忆条目 */
export interface MemoryItem {
  key: string;
  value: string;
  status: 'pending' | 'active' | 'stale' | 'expired';
  staleDays: number;
  lastVerifiedAt: string;
  expiresAt?: string;
  source: string;
  updatedAt: string;
}

export interface MemoryListResult {
  list: MemoryItem[];
}

// ==================== Vault API ====================

/** 目录树 */
export function getDocsTree(projectId: number, space?: string) {
  return Alova.Get<DocsTreeResult>(`/v1/projects/${projectId}/docs/tree`, {
    params: space ? { space } : {},
  });
}

/** 读文件 */
export function getDocsFile(projectId: number, path: string) {
  return Alova.Get<DocsFile>(`/v1/projects/${projectId}/docs/file`, {
    params: { path },
  });
}

/** 二进制原始下载地址 */
export function docsRawUrl(projectId: number, path: string) {
  return `/api/v1/projects/${projectId}/docs/raw?path=${encodeURIComponent(path)}`;
}

/** 写文件（整文件覆盖） */
export function writeDocsFile(projectId: number, path: string, content: string) {
  return Alova.Put<DocsWriteResult>(`/v1/projects/${projectId}/docs/file`, { path, content });
}

/** 补丁式修改 */
export function patchDocsFile(
  projectId: number,
  path: string,
  operations: DocsPatchOperation[]
) {
  return Alova.Post<DocsPatchResult>(`/v1/projects/${projectId}/docs/file/patch`, {
    path,
    operations,
  });
}

/** 创建目录 */
export function createDocsFolder(projectId: number, path: string) {
  return Alova.Post(`/v1/projects/${projectId}/docs/folder`, { path });
}

/** 上传文件 */
export function uploadDocsFile(projectId: number, path: string, file: File) {
  const form = new FormData();
  form.append('path', path);
  form.append('file', file);
  return Alova.Post<DocsWriteResult>(`/v1/projects/${projectId}/docs/upload`, form);
}

/** 移动/重命名 */
export function moveDocsPath(projectId: number, from: string, to: string) {
  return Alova.Post(`/v1/projects/${projectId}/docs/file/move`, { from, to });
}

/** 删除（DELETE 第二参是请求体而非配置，query 需显式拼 URL） */
export function deleteDocsPath(projectId: number, path: string) {
  return Alova.Delete(
    `/v1/projects/${projectId}/docs/file?path=${encodeURIComponent(path)}`
  );
}

/** 更新元数据（frontmatter） */
export function updateDocsMeta(
  projectId: number,
  path: string,
  meta: {
    title?: string;
    type?: string;
    status?: 'draft' | 'published';
    tags?: string[];
    linked?: string[];
  }
) {
  return Alova.Put(`/v1/projects/${projectId}/docs/meta`, { path, ...meta });
}

/** 搜索 */
export function searchDocsApi(projectId: number, q: string, space?: string) {
  return Alova.Get<{ items: DocsSearchItem[] | null }>(`/v1/projects/${projectId}/docs/search`, {
    params: { q, ...(space ? { space } : {}) },
  });
}

/** 反查关联文档 */
export function getLinkedDocs(projectId: number, target: string) {
  const [type, id] = target.split(':');
  const params: Record<string, number> = {};
  if (type === 'task') params.task = Number(id);
  if (type === 'req') params.req = Number(id);
  if (type === 'tc') params.tc = Number(id);
  return Alova.Get<{ items: DocsSearchItem[] | null }>(`/v1/projects/${projectId}/docs/linked`, {
    params,
  });
}

/** 手动触发增量重扫 */
export function refreshDocs(projectId: number) {
  return Alova.Post<DocsRefreshResult>(`/v1/projects/${projectId}/docs/refresh`);
}

// ==================== Memories API ====================

/** 记忆列表 */
export function getMemories(projectId: number, prefix?: string, include?: string) {
  return Alova.Get<MemoryListResult>(`/v1/projects/${projectId}/memories`, {
    params: { ...(prefix ? { prefix } : {}), ...(include ? { include } : {}) },
  });
}

/** 读单条 */
export function getMemory(projectId: number, key: string) {
  return Alova.Get<MemoryItem>(`/v1/projects/${projectId}/memories/${encodeURIComponent(key)}`);
}

/** 写入（upsert） */
export function setMemory(
  projectId: number,
  key: string,
  data: { value: string; ttl?: string; status?: 'pending' | 'active' }
) {
  return Alova.Put(`/v1/projects/${projectId}/memories/${encodeURIComponent(key)}`, data);
}

/** 验证保鲜 */
export function verifyMemory(projectId: number, key: string) {
  return Alova.Post(`/v1/projects/${projectId}/memories/${encodeURIComponent(key)}/verify`);
}

/** 主动废弃 */
export function expireMemory(projectId: number, key: string) {
  return Alova.Post(`/v1/projects/${projectId}/memories/${encodeURIComponent(key)}/expire`);
}

/** 删除 */
export function deleteMemory(projectId: number, key: string) {
  return Alova.Delete(`/v1/projects/${projectId}/memories/${encodeURIComponent(key)}`);
}
