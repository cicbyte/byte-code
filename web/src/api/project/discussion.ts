import { Alova } from '@/utils/http/alova/index';

/** ==================== 讨论区：想法/议题线程，可转任务 ==================== */

export interface DiscussionItem {
  id: number;
  projectId: number;
  title: string;
  body: string;
  status: 'open' | 'converted' | 'archived';
  authorId: number;
  authorName?: string;
  authorType: 'human' | 'ai';
  createdAt: string;
  updatedAt: string;
  replyCount: number;
  convertedTaskId: number;
}

export interface DiscussionReplyItem {
  id: number;
  discussionId: number;
  userId: number;
  userName?: string;
  userType: 'human' | 'ai';
  content: string;
  createdAt: string;
}

export interface DiscussionDetail extends Omit<DiscussionItem, 'replyCount'> {
  replies: DiscussionReplyItem[];
}

export interface DiscussionListResult {
  list: DiscussionItem[];
  total: number;
  currentPage: number;
}

export function getDiscussions(pid: number, params?: { pageNum?: number; pageSize?: number; status?: string }) {
  return Alova.Get<DiscussionListResult>(`/v1/projects/${pid}/discussions`, { params });
}

export function createDiscussion(pid: number, data: { title: string; body: string }) {
  return Alova.Post<{ id: number }>(`/v1/projects/${pid}/discussions`, data);
}

export function getDiscussionDetail(id: number) {
  return Alova.Get<DiscussionDetail>(`/v1/discussions/${id}`);
}

export function updateDiscussion(id: number, data: { title?: string; body?: string }) {
  return Alova.Put(`/v1/discussions/${id}`, data);
}

export function deleteDiscussion(id: number) {
  return Alova.Delete(`/v1/discussions/${id}`);
}

export function archiveDiscussion(id: number) {
  return Alova.Post<{ status: string }>(`/v1/discussions/${id}/archive`);
}

export function replyDiscussion(id: number, content: string) {
  return Alova.Post<{ id: number }>(`/v1/discussions/${id}/replies`, { content });
}

export function convertDiscussion(id: number, data: { title?: string; type?: string }) {
  return Alova.Post<{ taskId: number; discussionId: number }>(`/v1/discussions/${id}/convert`, data);
}
