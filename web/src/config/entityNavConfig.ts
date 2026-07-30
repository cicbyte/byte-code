export interface NavTab {
  key: string;
  label: string;
  path: string;
}

export interface EntityNavConfig {
  basePath: string;
  tabs: NavTab[];
}

const projectTabs: NavTab[] = [
  { key: 'overview', label: '概览', path: '/project/:projectId/overview' },
  { key: 'board', label: '任务看板', path: '/project/:projectId/board' },
  { key: 'tasks', label: '任务列表', path: '/project/:projectId/tasks' },
  { key: 'requirements', label: '需求池', path: '/project/:projectId/requirements' },
  { key: 'milestones', label: '里程碑', path: '/project/:projectId/milestones' },
  { key: 'sprints', label: 'Sprint', path: '/project/:projectId/sprints' },
  { key: 'members', label: '成员', path: '/project/:projectId/members' },
  { key: 'database', label: '数据库', path: '/project/:projectId/database' },
];

export const entityNavConfigs: Record<string, EntityNavConfig> = {
  project: { basePath: '/project', tabs: projectTabs },
};

export function getEntityTabs(entityType: string, entityId: number): NavTab[] {
  const config = entityNavConfigs[entityType];
  if (!config) return [];
  return config.tabs.map((tab) => ({
    ...tab,
    path: tab.path.replace(':projectId', String(entityId)),
  }));
}
