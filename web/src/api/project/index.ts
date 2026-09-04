import { Alova } from '@/utils/http/alova/index';

// ==================== 类型定义 ====================

/** 项目 */
export interface ProjectItem {
  id: number;
  name: string;
  description: string;
  createdBy: number;
  creatorName: string;
  status: number;
  createdAt: string;
  updatedAt: string;
}

export interface ProjectListResult {
  list: ProjectItem[];
  total: number;
  page: number;
  size: number;
}

export interface ProjectListParams {
  status?: number;
  keyword?: string;
  page?: number;
  size?: number;
}

export interface ProjectCreateData {
  name: string;
  description?: string;
}

export interface ProjectUpdateData {
  name?: string;
  description?: string;
  status?: number;
}

/** 项目成员 */
export interface MemberItem {
  id: number;
  userId: number;
  username: string;
  realName: string;
  role: string;
  joinedAt: string;
}

export interface MemberListResult {
  list: MemberItem[];
}

export interface MemberAddData {
  userId: number;
  role?: string;
}

/** 任务 */
export interface TaskItem {
  id: number;
  projectId: number;
  requirementId: number;
  sprintId: number;
  title: string;
  description: string;
  tags?: string[] | null;
  type: string;
  status: string;
  priority: number;
  assigneeId: number;
  assigneeName: string;
  creatorId: number;
  creatorName: string;
  parentTaskId: number;
  artifacts: string;
  requiresHumanReview: number;
  humanReviewStatus: string;
  sortOrder: number;
  source: string;
  createdAt: string;
  updatedAt: string;
}

export interface TaskListResult {
  list: TaskItem[];
  total: number;
}

export interface TaskListParams {
  status?: string;
  type?: string;
  sprintId?: number;
  assigneeId?: number;
  tagId?: number;
  keyword?: string;
  page?: number;
  size?: number;
}

export interface TaskCreateData {
  requirementId?: number;
  sprintId?: number;
  title: string;
  description?: string;
  type?: string;
  priority?: number;
  assigneeId?: number;
  parentTaskId?: number;
}

export interface TaskUpdateData {
  title?: string;
  description?: string;
  type?: string;
  status?: string;
  priority?: number;
  assigneeId?: number;
  sprintId?: number;
  parentTaskId?: number;
  sortOrder?: number;
}

export interface TaskImportData {
  requirementId: number;
  sprintId?: number;
}

export interface TaskImportResult {
  taskIds: number[];
}

/** 评论 */
export interface CommentItem {
  id: number;
  taskId: number;
  userId: number;
  username: string;
  realName: string;
  content: string;
  userType: string;
  createdAt: string;
}

export interface CommentListResult {
  list: CommentItem[];
}

export interface CommentCreateData {
  content: string;
  userType?: string;
}

/** AI 执行日志 */
export interface AiLogItem {
  id: number;
  taskId: number;
  aiUserId: number;
  aiUsername?: string;
  action: string;
  detail: string;
  status: string;
  createdAt: string;
}

export interface AiLogListResult {
  list: AiLogItem[];
}

export interface AiLogCreateData {
  aiUserId: number;
  action: string;
  detail?: string;
  status?: string;
}

/** Sprint */
export interface SprintItem {
  id: number;
  projectId: number;
  name: string;
  goal: string;
  startDate: string;
  endDate: string;
  status: string;
  createdAt: string;
}

export interface SprintListResult {
  list: SprintItem[];
}

export interface SprintListParams {
  status?: string;
}

export interface SprintCreateData {
  name: string;
  goal?: string;
  startDate: string;
  endDate: string;
}

export interface SprintUpdateData {
  name?: string;
  goal?: string;
  startDate?: string;
  endDate?: string;
  status?: string;
}

/** 燃尽图 */
export interface BurndownItem {
  date: string;
  remaining: number;
  completed: number;
}

export interface BurndownResult {
  items: BurndownItem[];
}

/** 需求 */
export interface RequirementItem {
  id: number;
  projectId: number;
  parentId: number;
  type: string;
  title: string;
  description: string;
  status: string;
  priority: number;
  assigneeId: number;
  assigneeName: string;
  creatorId: number;
  creatorName: string;
  milestoneId: number;
  sortOrder: number;
  acceptanceCriteria: string;
  source: string;
  children?: RequirementItem[];
  createdAt: string;
  updatedAt: string;
}

export interface RequirementListResult {
  list: RequirementItem[];
  total: number;
}

export interface RequirementListParams {
  type?: string;
  status?: string;
  parentId?: number;
  page?: number;
  size?: number;
}

export interface RequirementCreateData {
  parentId?: number;
  type?: string;
  title: string;
  description?: string;
  priority?: number;
  assigneeId?: number;
  milestoneId?: number;
  acceptanceCriteria?: string;
}

export interface RequirementUpdateData {
  title?: string;
  description?: string;
  type?: string;
  status?: string;
  priority?: number;
  assigneeId?: number;
  milestoneId?: number;
  acceptanceCriteria?: string;
  sortOrder?: number;
}

/** 里程碑 */
export interface MilestoneItem {
  id: number;
  projectId: number;
  name: string;
  description: string;
  targetDate: string;
  status: string;
  createdAt: string;
}

export interface MilestoneListResult {
  list: MilestoneItem[];
}

export interface MilestoneCreateData {
  name: string;
  description?: string;
  targetDate?: string;
}

// ==================== 项目 API ====================

/** 项目列表 */
export function getProjects(params?: ProjectListParams) {
  return Alova.Get<ProjectListResult>('/v1/projects', { params });
}

/** 创建项目 */
export function createProject(data: ProjectCreateData) {
  return Alova.Post<{ id: number }>('/v1/projects', data);
}

/** 项目详情 */
export function getProject(id: number) {
  return Alova.Get<ProjectItem>(`/v1/projects/${id}`);
}

/** 更新项目 */
export function updateProject(id: number, data: ProjectUpdateData) {
  return Alova.Put(`/v1/projects/${id}`, data);
}

/** 删除项目 */
export function deleteProject(id: number) {
  return Alova.Delete(`/v1/projects/${id}`);
}

// ==================== 项目成员 API ====================

/** 成员列表 */
export function getMembers(projectId: number) {
  return Alova.Get<MemberListResult>(`/v1/projects/${projectId}/members`);
}

/** 添加成员 */
export function addMember(projectId: number, data: MemberAddData) {
  return Alova.Post(`/v1/projects/${projectId}/members`, data);
}

/** 移除成员 */
export function removeMember(projectId: number, userId: number) {
  return Alova.Delete(`/v1/projects/${projectId}/members/${userId}`);
}

// ==================== 任务 API ====================

/** 任务列表 */
export function getTasks(projectId: number, params?: TaskListParams) {
  return Alova.Get<TaskListResult>(`/v1/projects/${projectId}/tasks`, { params });
}

/** 创建任务 */
export function createTask(projectId: number, data: TaskCreateData) {
  return Alova.Post<{ id: number }>(`/v1/projects/${projectId}/tasks`, data);
}

/** 任务详情 */
export function getTask(id: number) {
  return Alova.Get<TaskItem>(`/v1/tasks/${id}`);
}

/** 更新任务 */
export function updateTask(id: number, data: TaskUpdateData) {
  return Alova.Put(`/v1/tasks/${id}`, data);
}

/** 删除任务 */
export function deleteTask(id: number) {
  return Alova.Delete(`/v1/tasks/${id}`);
}

/** AI 认领任务 */
export function claimTask(id: number) {
  return Alova.Post(`/v1/tasks/${id}/claim`);
}

/** AI 完成任务 */
export function completeTask(id: number, data?: { artifacts?: string }) {
  return Alova.Post(`/v1/tasks/${id}/complete`, data);
}

/** 审核任务 */
export function reviewTask(id: number, data: { status: 'approved' | 'rejected'; comment?: string }) {
  return Alova.Post(`/v1/tasks/${id}/review`, data);
}

/** 从需求导入任务 */
export function importTasks(projectId: number, data: TaskImportData) {
  return Alova.Post<TaskImportResult>(`/v1/projects/${projectId}/tasks/import`, data);
}

// ==================== 评论 API ====================

/** 评论列表 */
export function getComments(taskId: number) {
  return Alova.Get<CommentListResult>(`/v1/tasks/${taskId}/comments`);
}

/** 创建评论 */
export function createComment(taskId: number, data: CommentCreateData) {
  return Alova.Post<{ id: number }>(`/v1/tasks/${taskId}/comments`, data);
}

// ==================== AI 执行日志 API ====================

/** AI 执行日志列表 */
export function getAiLogs(taskId: number) {
  return Alova.Get<AiLogListResult>(`/v1/tasks/${taskId}/ai-logs`);
}

/** 创建 AI 执行日志 */
export function createAiLog(taskId: number, data: AiLogCreateData) {
  return Alova.Post<{ id: number }>(`/v1/tasks/${taskId}/ai-logs`, data);
}

// ==================== Sprint API ====================

/** Sprint 列表 */
export function getSprints(projectId: number, params?: SprintListParams) {
  return Alova.Get<SprintListResult>(`/v1/projects/${projectId}/sprints`, { params });
}

/** 创建 Sprint */
export function createSprint(projectId: number, data: SprintCreateData) {
  return Alova.Post<{ id: number }>(`/v1/projects/${projectId}/sprints`, data);
}

/** Sprint 详情 */
export function getSprint(id: number) {
  return Alova.Get<SprintItem>(`/v1/sprints/${id}`);
}

/** 更新 Sprint */
export function updateSprint(id: number, data: SprintUpdateData) {
  return Alova.Put(`/v1/sprints/${id}`, data);
}

/** 删除 Sprint */
export function deleteSprint(id: number) {
  return Alova.Delete(`/v1/sprints/${id}`);
}

/** 添加任务到 Sprint */
export function addTaskToSprint(sprintId: number, taskId: number) {
  return Alova.Post(`/v1/sprints/${sprintId}/tasks`, { taskId });
}

/** 从 Sprint 移除任务 */
export function removeTaskFromSprint(sprintId: number, taskId: number) {
  return Alova.Delete(`/v1/sprints/${sprintId}/tasks/${taskId}`);
}

/** 燃尽图数据 */
export function getSprintBurndown(id: number) {
  return Alova.Get<BurndownResult>(`/v1/sprints/${id}/burndown`);
}

// ==================== 需求 API ====================

/** 需求列表 */
export function getRequirements(projectId: number, params?: RequirementListParams) {
  return Alova.Get<RequirementListResult>(`/v1/projects/${projectId}/requirements`, { params });
}

/** 创建需求 */
export function createRequirement(projectId: number, data: RequirementCreateData) {
  return Alova.Post<{ id: number }>(`/v1/projects/${projectId}/requirements`, data);
}

/** 更新需求 */
export function updateRequirement(id: number, data: RequirementUpdateData) {
  return Alova.Put(`/v1/requirements/${id}`, data);
}

/** 删除需求 */
export function deleteRequirement(id: number) {
  return Alova.Delete(`/v1/requirements/${id}`);
}

// ==================== 里程碑 API ====================

/** 里程碑列表 */
export function getMilestones(projectId: number) {
  return Alova.Get<MilestoneListResult>(`/v1/projects/${projectId}/milestones`);
}

export function updateMilestone(id: number, data: Partial<MilestoneItem>) {
  return Alova.Put(`/v1/milestones/${id}`, data);
}

export function deleteMilestone(id: number) {
  return Alova.Delete(`/v1/milestones/${id}`);
}

/** 创建里程碑 */
export function createMilestone(projectId: number, data: MilestoneCreateData) {
  return Alova.Post<{ id: number }>(`/v1/projects/${projectId}/milestones`, data);
}


// ==================== 评论扩展 ====================

export function updateComment(id: number, content: string) {
  return Alova.Put(`/v1/comments/${id}`, { content });
}

export function deleteComment(id: number) {
  return Alova.Delete(`/v1/comments/${id}`);
}
