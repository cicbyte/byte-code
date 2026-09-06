/**
 * 任务字典：状态/类型/优先级 的 label 与 tag 配色统一出口。
 * 状态与后端 internal/consts.TaskStatus* 一致；类型与 tasks 表
 * CHECK(type IN feature/bug/chore/test) 一致——改这里须同步后端。
 */

/** naive-ui n-tag 的 type 取值（根包未导出该联合类型，本地声明） */
export type TagType = 'default' | 'primary' | 'success' | 'info' | 'warning' | 'error';

export interface TaskStatusDict {
  value: string;
  label: string;
  tagType: TagType;
}

export const TASK_STATUS: TaskStatusDict[] = [
  { value: 'open', label: '待处理', tagType: 'default' },
  { value: 'in_progress', label: '进行中', tagType: 'info' },
  { value: 'blocked', label: '已阻塞', tagType: 'error' },
  { value: 'review', label: '审核中', tagType: 'warning' },
  { value: 'done', label: '已完成', tagType: 'success' },
  { value: 'closed', label: '已关闭', tagType: 'success' },
];

/** 任务类型（受表 CHECK 约束，勿加字典外的值） */
export const TASK_TYPES = [
  { value: 'bug', label: 'Bug', tagType: 'error' },
  { value: 'feature', label: 'Feature', tagType: 'success' },
  { value: 'chore', label: 'Chore', tagType: 'default' },
  { value: 'test', label: 'Test', tagType: 'info' },
] as const;

/** 看板列（closed 不入板，与后端 TaskActiveStatuses + done 口径一致；blocked 为"等人"信号列） */
export const BOARD_COLUMNS = TASK_STATUS.filter((s) => s.value !== 'closed');

const statusMap = new Map(TASK_STATUS.map((s) => [s.value, s]));
const typeMap = new Map<string, { label: string; tagType: TagType }>(
  TASK_TYPES.map((t) => [t.value, { label: t.label, tagType: t.tagType }]),
);

export function statusLabel(s: string): string {
  return statusMap.get(s)?.label || s;
}

export function statusTagType(s: string): TagType {
  return statusMap.get(s)?.tagType || 'default';
}

export function typeLabel(t: string): string {
  return typeMap.get(t)?.label || t;
}

export function typeTagType(t: string): TagType {
  return typeMap.get(t)?.tagType || 'default';
}

/** 优先级配色：4=紧急红 / 3=高橙 / 2=中蓝 / 1=低灰 */
export function priorityTagType(p: number | undefined | null): TagType {
  if (p === undefined || p === null) return 'default';
  if (p >= 4) return 'error';
  if (p >= 3) return 'warning';
  if (p >= 2) return 'info';
  return 'default';
}
