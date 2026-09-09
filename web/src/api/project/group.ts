import { Alova } from '@/utils/http/alova/index';

// ==================== 项目分组（多对多） ====================

export interface GroupProjectBrief {
  id: number;
  code: string;
  name: string;
}

export interface GroupItem {
  id: number;
  name: string;
  description: string;
  createdBy: number;
  createdAt: string;
  projects: GroupProjectBrief[];
}

export function getGroups() {
  return Alova.Get<{ list: GroupItem[] }>('/v1/groups');
}

export function createGroup(data: { name: string; description?: string }) {
  return Alova.Post<{ id: number }>('/v1/groups', data);
}

export function updateGroup(id: number, data: { name?: string; description?: string }) {
  return Alova.Put(`/v1/groups/${id}`, data);
}

export function deleteGroup(id: number) {
  return Alova.Delete(`/v1/groups/${id}`);
}

export function addProjectToGroup(groupId: number, projectId: number) {
  return Alova.Post(`/v1/groups/${groupId}/projects`, { projectId });
}

export function removeProjectFromGroup(groupId: number, projectId: number) {
  return Alova.Delete(`/v1/groups/${groupId}/projects/${projectId}`);
}
