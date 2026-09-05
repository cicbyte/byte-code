// 通知类型与来源字典：通知中心（platform/notifications）与 Header 铃铛
// （NotificationIcon）共用。单一出口防两处口径漂移——此前 info 一处
// 显示"通知"、另一处显示"信息"。未知类型显示原值（诚实回落）
export const NOTICE_TYPE_LABELS: Record<string, string> = {
  info: '信息',
  warning: '警告',
  success: '成功',
  error: '错误',
};

export const NOTICE_TYPE_TAG: Record<string, 'info' | 'success' | 'warning' | 'error'> = {
  info: 'info',
  warning: 'warning',
  success: 'success',
  error: 'error',
};

export const NOTICE_SOURCE_LABELS: Record<string, string> = {
  task: '任务',
  project: '项目',
};

export function noticeTypeLabel(t: string): string {
  return NOTICE_TYPE_LABELS[t] || t;
}

export function noticeTypeTagType(t: string): 'info' | 'success' | 'warning' | 'error' {
  return NOTICE_TYPE_TAG[t] || 'info';
}

/** 筛选下拉选项（label/value 与字典同源） */
export const noticeTypeOptions = Object.entries(NOTICE_TYPE_LABELS).map(([value, label]) => ({
  label,
  value,
}));
