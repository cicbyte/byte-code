// 活动流/审计动作与目标类型字典：platform/activities 与项目 overview 共用
// （此前 overview 本地一份映射、activities 页英文直出）
export const ACTION_LABELS: Record<string, string> = {
  create: '创建了', update: '更新了', delete: '删除了',
  'project.created': '创建了项目',
  'task.created': '创建了任务', 'task.claimed': '认领了任务', 'task.completed': '完成了任务',
  'sprint.created': '创建了迭代', 'sprint.deleted': '删除了迭代',
  'test_plan.created': '创建了测试计划', 'test_plan.updated': '更新了测试计划',
  'test_plan.deleted': '删除了测试计划', 'test_plan.added_cases': '向计划添加了用例',
  'test_case.created': '创建了测试用例', 'test_case.deleted': '删除了测试用例',
  'test_case.executed': '执行了测试用例',
  'attachment.uploaded': '上传了附件', 'attachment.deleted': '删除了附件',
};

export function actionText(a: string): string {
  return ACTION_LABELS[a] || `执行了 ${a}`;
}

export const TARGET_TYPE_LABELS: Record<string, string> = {
  project: '项目', requirement: '需求', task: '任务', sprint: '迭代',
  test_plan: '测试计划', test_case: '测试用例', attachment: '附件',
  doc: '文档', comment: '评论', ai_user: 'Agent', user: '用户',
};
