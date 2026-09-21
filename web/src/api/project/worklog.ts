import { Alova } from '@/utils/http/alova/index';

/** ==================== 工作日志：项目演化叙事，日期分组时间轴 ==================== */

export interface WorklogItem {
  id: number;
  projectId: number;
  authorId: number;
  authorName?: string;
  authorType: 'human' | 'ai';
  content: string;
  source: 'manual' | 'tasks';
  createdAt: string;
  updatedAt: string;
}

export interface WorklogListResult {
  list: WorklogItem[];
  total: number;
  currentPage: number;
}

export interface WorklogDraftResult {
  content: string;
  taskCount: number;
  releaseCount: number;
}

export function getWorklogs(pid: number, params?: { pageNum?: number; pageSize?: number }) {
  return Alova.Get<WorklogListResult>(`/v1/projects/${pid}/worklogs`, { params });
}

export function getWorklogDetail(id: number) {
  return Alova.Get<WorklogItem>(`/v1/worklogs/${id}`);
}

export function createWorklog(pid: number, data: { content: string; source?: 'manual' | 'tasks' }) {
  return Alova.Post<{ id: number }>(`/v1/projects/${pid}/worklogs`, data);
}

export function updateWorklog(id: number, content: string) {
  return Alova.Put(`/v1/worklogs/${id}`, { content });
}

export function deleteWorklog(id: number) {
  return Alova.Delete(`/v1/worklogs/${id}`);
}

export function getWorklogDraft(pid: number, from: string, to: string) {
  return Alova.Get<WorklogDraftResult>(`/v1/projects/${pid}/worklogs/draft`, {
    params: { from, to },
  });
}
