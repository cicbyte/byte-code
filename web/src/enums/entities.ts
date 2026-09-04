/**
 * 其它实体字典：需求/里程碑/Sprint/记忆/项目状态的 label 与 tag 配色统一出口。
 * 与任务字典（task.ts）分域；各状态值与后端 logic 层口径一致。
 */
import type { TagType } from './task';

export interface EntityStatusDict {
  value: string;
  label: string;
  tagType: TagType;
}

function buildDict(rows: EntityStatusDict[]) {
  const map = new Map(rows.map((r) => [r.value, r]));
  const options = rows.map((r) => ({ label: r.label, value: r.value }));
  return {
    rows,
    options,
    label: (v: string) => map.get(v)?.label || v,
    tagType: (v: string): TagType => map.get(v)?.tagType || 'default',
  };
}

/** 需求状态 */
export const REQ_STATUS = buildDict([
  { value: 'draft', label: '草稿', tagType: 'default' },
  { value: 'active', label: '进行中', tagType: 'info' },
  { value: 'completed', label: '已完成', tagType: 'success' },
  { value: 'archived', label: '已归档', tagType: 'default' },
]);

/** 需求类型（与 requirements 表口径一致） */
export const REQ_TYPES = buildDict([
  { value: 'epic', label: 'Epic', tagType: 'success' },
  { value: 'story', label: 'Story', tagType: 'info' },
  { value: 'task', label: 'Task', tagType: 'default' },
]);

/** 里程碑状态 */
export const MILESTONE_STATUS = buildDict([
  { value: 'planning', label: '规划中', tagType: 'default' },
  { value: 'in_progress', label: '进行中', tagType: 'info' },
  { value: 'released', label: '已发布', tagType: 'success' },
]);

/** Sprint 状态（测试计划复用同一套 planning/active/completed） */
export const SPRINT_STATUS = buildDict([
  { value: 'planning', label: '规划中', tagType: 'default' },
  { value: 'active', label: '进行中', tagType: 'info' },
  { value: 'completed', label: '已完成', tagType: 'success' },
]);

/** 项目 KV 记忆状态（memories / global-memory 共用） */
export const MEMORY_STATUS = buildDict([
  { value: 'active', label: '有效', tagType: 'success' },
  { value: 'pending', label: '待验证', tagType: 'warning' },
  { value: 'stale', label: '已腐化', tagType: 'error' },
  { value: 'expired', label: '已过期', tagType: 'default' },
]);

/** 项目状态（数值型：1=进行中 2=已结束） */
const PROJECT_ROWS: Array<{ value: number; label: string; tagType: TagType }> = [
  { value: 1, label: '进行中', tagType: 'success' },
  { value: 2, label: '已结束', tagType: 'default' },
];
const projectMap = new Map(PROJECT_ROWS.map((r) => [r.value, r]));
export const PROJECT_STATUS = {
  options: PROJECT_ROWS.map((r) => ({ label: r.label, value: r.value })),
  label: (v: number) => projectMap.get(v)?.label || String(v),
  tagType: (v: number): TagType => projectMap.get(v)?.tagType || 'default',
};
