import { Alova } from '@/utils/http/alova/index';

// ==================== 类型定义 ====================

/** 测试用例 */
export interface TestCaseItem {
  id: number;
  projectId: number;
  requirementId: number;
  taskId: number;
  title: string;
  preconditions: string;
  steps: string;
  expectedResult: string;
  category: string;
  module: string;
  priority: string;
  source: string;
  creatorId: number;
  status: string;
  createdAt: string;
  updatedAt: string;
  creatorName?: string;
}

export interface TestCaseListResult {
  list: TestCaseItem[];
  total: number;
  currentPage: number;
}

export interface TestCaseListParams {
  category?: string;
  module?: string;
  status?: string;
  keyword?: string;
  pageNum?: number;
  pageSize?: number;
}

export interface TestCaseCreateData {
  requirementId?: number;
  taskId?: number;
  title: string;
  preconditions?: string;
  steps?: string;
  expectedResult?: string;
  category?: string;
  module?: string;
  priority: string;
  source?: string;
}

export interface TestCaseUpdateData {
  requirementId?: number;
  taskId?: number;
  title?: string;
  preconditions?: string;
  steps?: string;
  expectedResult?: string;
  category?: string;
  module?: string;
  priority?: string;
  status?: string;
}

/** 测试计划 */
export interface TestPlanItem {
  id: number;
  projectId: number;
  name: string;
  description: string;
  milestoneId: number;
  status: string;
  creatorId: number;
  createdAt: string;
  updatedAt: string;
  creatorName?: string;
}

export interface TestPlanListResult {
  list: TestPlanItem[];
  total: number;
  currentPage: number;
}

export interface TestPlanListParams {
  status?: string;
  pageNum?: number;
  pageSize?: number;
}

export interface TestPlanCreateData {
  name: string;
  description?: string;
  milestoneId?: number;
}

export interface TestPlanUpdateData {
  name?: string;
  description?: string;
  milestoneId?: number;
  status?: string;
}

/** 测试计划添加用例 */
export interface TestPlanAddCaseData {
  caseIds: number[];
  assigneeId?: number;
}

/** 测试执行 */
export interface TestCaseExecuteData {
  status: 'pass' | 'fail' | 'blocked' | 'skip';
  actualResult?: string;
  bugTaskId?: number;
}

/** 测试计划执行结果 */
export interface TestPlanCaseResult {
  id: number;
  testCaseId: number;
  testCaseTitle: string;
  assigneeId: number;
  assigneeName?: string;
  status: string;
  actualResult: string;
  bugTaskId: number;
  executedAt: string;
}

export interface TestPlanResultsResult {
  total: number;
  passed: number;
  failed: number;
  blocked: number;
  skipped: number;
  pending: number;
  results: TestPlanCaseResult[];
}

// ==================== 测试用例 API ====================

/** 用例列表 */
export function getTestCases(projectId: number, params?: TestCaseListParams) {
  return Alova.Get<TestCaseListResult>(`/v1/projects/${projectId}/test-cases`, { params });
}

/** 创建用例 */
export function createTestCase(projectId: number, data: TestCaseCreateData) {
  return Alova.Post<{ id: number }>(`/v1/projects/${projectId}/test-cases`, data);
}

/** 更新用例 */
export function updateTestCase(id: number, data: TestCaseUpdateData) {
  return Alova.Put(`/v1/test-cases/${id}`, data);
}

/** 删除用例 */
export function deleteTestCase(id: number) {
  return Alova.Delete(`/v1/test-cases/${id}`);
}

// ==================== 测试计划 API ====================

/** 计划列表 */
export function getTestPlans(projectId: number, params?: TestPlanListParams) {
  return Alova.Get<TestPlanListResult>(`/v1/projects/${projectId}/test-plans`, { params });
}

/** 创建计划 */
export function createTestPlan(projectId: number, data: TestPlanCreateData) {
  return Alova.Post<{ id: number }>(`/v1/projects/${projectId}/test-plans`, data);
}

/** 更新计划 */
export function updateTestPlan(id: number, data: TestPlanUpdateData) {
  return Alova.Put(`/v1/test-plans/${id}`, data);
}

/** 删除计划 */
export function deleteTestPlan(id: number) {
  return Alova.Delete(`/v1/test-plans/${id}`);
}

// ==================== 测试执行 API ====================

/** 添加用例到计划 */
export function addCasesToPlan(planId: number, data: TestPlanAddCaseData) {
  return Alova.Post(`/v1/test-plans/${planId}/cases`, data);
}

/** 执行用例 */
export function executeTestCase(planCaseId: number, data: TestCaseExecuteData) {
  return Alova.Put(`/v1/test-plan-cases/${planCaseId}/execute`, data);
}

/** 计划执行结果 */
export function getTestPlanResults(planId: number) {
  return Alova.Get<TestPlanResultsResult>(`/v1/test-plans/${planId}/results`);
}
