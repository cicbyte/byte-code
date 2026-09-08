import { Alova } from '@/utils/http/alova/index';

// ==================== 类型定义 ====================

/** 项目 */
export interface ProjectItem {
  id: number;
  code: string;
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
  /** human/ai（@提及候选只列 human） */
  userType?: string;
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
  checklist?: string;
  relatedRefs?: string;
  type: string;
  status: string;
  priority: number;
  assigneeId: number;
  assigneeName: string;
  creatorId: number;
  creatorName: string;
  parentTaskId: number;
  artifacts: string;
  /** 截止日期（Y-m-d），空串无截止 */
  dueDate: string;
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

// ==================== 我的任务（跨项目聚合） ====================

export interface MyTaskItem extends TaskItem {
  projectName: string;
}

export interface MyTaskListResult {
  list: MyTaskItem[];
  total: number;
}

export interface MyTaskListParams {
  status?: string;
  projectId?: number;
  keyword?: string;
  page?: number;
  size?: number;
}

/** 我的任务（当前用户被指派的跨项目任务；status 缺省=进行中三态，all=全部） */
export function getMyTasks(params?: MyTaskListParams) {
  return Alova.Get<MyTaskListResult>('/v1/my-tasks', { params });
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
  /** 截止日期（Y-m-d），缺省无截止 */
  dueDate?: string;
  /** 步骤清单 JSON [{text,done}] */
  checklist?: string;
  /** 跨项目引用 JSON [{type,projectId,id,title}] */
  relatedRefs?: string;
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
  /** 步骤清单 JSON [{text,done}]；打勾即进展（顺带续租约） */
  checklist?: string;
  /** 截止日期（Y-m-d）；传空串清除 */
  dueDate?: string;
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

// ==================== QA 库 ====================

export interface QaItem {
  id: number;
  question: string;
  answer: string;
  tags: string;
  hits: number;
  status: string;
  updater: string;
  updatedAt: string;
}

export function getQas(projectId: number, params?: { keyword?: string; tag?: string }) {
  return Alova.Get<{ list: QaItem[] }>(`/v1/projects/${projectId}/qas`, { params });
}

export function upsertQa(projectId: number, data: { question: string; answer: string; tags?: string }) {
  return Alova.Post<{ id: number; updated: boolean }>(`/v1/projects/${projectId}/qas`, data);
}

export function hitQa(projectId: number, id: number) {
  return Alova.Post(`/v1/projects/${projectId}/qas/${id}/hit`, {});
}

export function archiveQa(projectId: number, id: number) {
  return Alova.Post(`/v1/projects/${projectId}/qas/${id}/archive`, {});
}

// ==================== 专题（long-task） ====================

export interface TopicPhase {
  id: number;
  title: string;
  detail: string;
  status: string;
  assigneeId: number;
  assigneeName: string;
  artifacts: string;
  taskId: number;
  taskTitle?: string;
  taskStatus?: string;
  taskAssignee?: string;
  sortOrder: number;
}

export interface TopicItem {
  id: number;
  title: string;
  goal: string;
  acceptance: string;
  docPath: string;
  assigneeId: number;
  assigneeName: string;
  status: string;
  createdAt: string;
  phases: TopicPhase[];
  phaseTotal: number;
  phaseDone: number;
  lastHandoff: string;
}

export function getTopics(projectId: number, params?: { status?: string }) {
  return Alova.Get<{ list: TopicItem[] }>(`/v1/projects/${projectId}/topics`, { params });
}

export function createTopic(projectId: number, data: {
  title: string; goal?: string; acceptance?: string; docPath?: string; assigneeId?: number;
}) {
  return Alova.Post<{ id: number }>(`/v1/projects/${projectId}/topics`, data);
}

export function upsertTopicPhases(projectId: number, topicId: number, phases: Array<{ title: string; detail?: string }>) {
  return Alova.Post(`/v1/projects/${projectId}/topics/${topicId}/phases`, { phases });
}

export function updateTopicPhase(
  projectId: number, topicId: number, phaseId: number,
  data: { title?: string; detail?: string; assigneeId?: number; artifacts?: string },
) {
  return Alova.Put(`/v1/projects/${projectId}/topics/${topicId}/phases/${phaseId}`, data);
}

export function deleteTopicPhase(projectId: number, topicId: number, phaseId: number) {
  return Alova.Delete(`/v1/projects/${projectId}/topics/${topicId}/phases/${phaseId}`);
}

export function appendTopicPhase(projectId: number, topicId: number, title: string, detail?: string) {
  return Alova.Post<{ id: number }>(`/v1/projects/${projectId}/topics/${topicId}/phases/add`, { title, detail });
}

export function toggleTopicPhase(projectId: number, topicId: number, phaseId: number, status: string) {
  return Alova.Post(`/v1/projects/${projectId}/topics/${topicId}/phases/${phaseId}/toggle`, { status });
}

export function convertTopicPhase(projectId: number, topicId: number, phaseId: number) {
  return Alova.Post<{ taskId: number }>(`/v1/projects/${projectId}/topics/${topicId}/phases/${phaseId}/convert`, {});
}

export function logTopic(projectId: number, topicId: number, action: 'progress' | 'handoff', detail: string) {
  return Alova.Post(`/v1/projects/${projectId}/topics/${topicId}/log`, { action, detail });
}

export function finishTopic(projectId: number, topicId: number, result: 'completed' | 'abandoned') {
  return Alova.Post(`/v1/projects/${projectId}/topics/${topicId}/finish`, { result });
}

// ==================== 跨项目反馈 ====================

export interface FeedbackItem {
  id: number;
  title: string;
  content: string;
  status: string;
  sourceProjectId: number;
  sourceProjectName: string;
  sourceTaskId: number;
  convertedTaskId: number;
  dismissReason: string;
  createdAt: string;
}

export interface FeedbackListResult {
  list: FeedbackItem[];
  total: number;
}

export function getFeedbacks(projectId: number, params?: { status?: string; page?: number; size?: number }) {
  return Alova.Get<FeedbackListResult>(`/v1/projects/${projectId}/feedbacks`, { params });
}

export function createFeedback(projectId: number, data: { title: string; content?: string; sourceTaskId?: number }) {
  return Alova.Post<{ id: number }>(`/v1/projects/${projectId}/feedbacks`, data);
}

export function convertFeedback(projectId: number, id: number, data?: { title?: string; sprintId?: number }) {
  return Alova.Post<{ taskId: number }>(`/v1/projects/${projectId}/feedbacks/${id}/convert`, data || {});
}

export function dismissFeedback(projectId: number, id: number, reason: string) {
  return Alova.Post(`/v1/projects/${projectId}/feedbacks/${id}/dismiss`, { reason });
}

// ==================== 项目关联 ====================

export interface RelationItem {
  id: number;
  projectId: number;
  name: string;
  createdAt: string;
}

export function getRelations(projectId: number) {
  return Alova.Get<{ list: RelationItem[] }>(`/v1/projects/${projectId}/relations`);
}

export function addRelation(projectId: number, relatedProjectId: number) {
  return Alova.Post(`/v1/projects/${projectId}/relations`, { relatedProjectId });
}

export function removeRelation(projectId: number, relationId: number) {
  return Alova.Delete(`/v1/projects/${projectId}/relations/${relationId}`);
}

/** 移除 Agent 项目准入（协议接入的 agent 行用；会话一并失效） */
export function removeAgentProject(projectId: number, agentId: number) {
  return Alova.Delete(`/v1/projects/${projectId}/agents/${agentId}`);
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

/** 审核任务 */
export function reviewTask(id: number, data: { status: 'approved' | 'rejected'; comment?: string }) {
  return Alova.Post(`/v1/tasks/${id}/review`, data);
}

/** 上报阻塞（assignee/owner；进行中→已阻塞，原因必填） */
export function blockTask(id: number, reason: string) {
  return Alova.Post(`/v1/tasks/${id}/block`, { reason });
}

/** 解除阻塞（已阻塞→进行中） */
export function unblockTask(id: number) {
  return Alova.Post(`/v1/tasks/${id}/unblock`, {});
}

/** 重开任务（done/closed→open 重新入池；原因必填，人类专属） */
export function reopenTask(id: number, reason: string) {
  return Alova.Post(`/v1/tasks/${id}/reopen`, { reason });
}

/** 创建评论 */
export function createComment(taskId: number, data: CommentCreateData) {
  return Alova.Post<{ id: number }>(`/v1/tasks/${taskId}/comments`, data);
}

/** 评论列表 */
export function getComments(taskId: number) {
  return Alova.Get<CommentListResult>(`/v1/tasks/${taskId}/comments`);
}

// ==================== AI 执行日志 API ====================

/** AI 执行日志列表 */
export function getAiLogs(taskId: number) {
  return Alova.Get<AiLogListResult>(`/v1/tasks/${taskId}/ai-logs`);
}

/** Sprint 列表 */
export function getSprints(projectId: number, params?: SprintListParams) {
  const query = params?.status ? `?status=${params.status}` : '';
  return Alova.Get<SprintListResult>(`/v1/projects/${projectId}/sprints${query}`);
}

/** 创建 Sprint */
export function createSprint(projectId: number, data: SprintCreateData) {
  return Alova.Post<{ id: number }>(`/v1/projects/${projectId}/sprints`, data);
}

/** 从需求导入生成任务（含子需求，事务） */
export function importTasks(projectId: number, data: { requirementId: number; sprintId?: number }) {
  return Alova.Post<{ taskIds: number[] }>(`/v1/projects/${projectId}/tasks/import`, data);
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
