/**
 * 任务截止日期展示：逾期/临近着色，完成态不着色
 * 返回 naive-ui n-tag 的 type；无截止返回 null
 */
export function dueTagType(due: string | null | undefined, status?: string): 'error' | 'warning' | 'info' | null {
  if (!due) return null;
  // 已完成/已关闭的任务不再渲染紧迫感
  if (status === 'done' || status === 'closed') return 'info';
  const today = new Date();
  const todayStr = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`;
  const d = due.slice(0, 10);
  if (d < todayStr) return 'error';
  // 3 天内临近
  const soon = new Date(today.getTime() + 3 * 24 * 3600 * 1000);
  const soonStr = `${soon.getFullYear()}-${String(soon.getMonth() + 1).padStart(2, '0')}-${String(soon.getDate()).padStart(2, '0')}`;
  if (d <= soonStr) return 'warning';
  return 'info';
}

/** 截止日期短标签（列表列用）：逾期带天数 */
export function dueLabel(due: string | null | undefined): string {
  if (!due) return '-';
  const d = due.slice(0, 10);
  const today = new Date();
  const todayStr = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`;
  if (d < todayStr) {
    const diff = Math.floor((today.getTime() - new Date(d).getTime()) / (24 * 3600 * 1000));
    return `${d}（逾期${diff}天）`;
  }
  return d;
}
