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
  // test 域走后端公共 common.PageReq（pageNum/pageSize），
  // 与 project 等域的 page/size 是两套并存的后端约定，勿"顺手统一"
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

// ==================== 测试执行记录（Run，pytest 上报） ====================

/** 执行记录（一次批量上报：pytest session / CI job / 手工批次） */
export interface TestRunItem {
  id: number;
  projectId: number;
  source: string;
  branch: string;
  gitSha: string;
  env: string;
  triggeredBy: number;
  triggeredByName?: string;
  total: number;
  passed: number;
  failed: number;
  skipped: number;
  errors: number;
  durationMs: number;
  startedAt: string;
  finishedAt: string;
  createdAt: string;
}

export interface TestRunListResult {
  list: TestRunItem[];
  total: number;
  currentPage: number;
}

export interface TestRunListParams {
  source?: string;
  status?: string;
  branch?: string;
  pageNum?: number;
  pageSize?: number;
}

export interface TestRunCaseItem {
  id: number;
  testRunId: number;
  testCaseId: number;
  testCaseTitle?: string;
  externalKey: string;
  title: string;
  status: string;
  durationMs: number;
  message: string;
  bugTaskId: number;
  bugTaskTitle?: string;
  flaky: boolean;
}

export interface TestRunDetail extends TestRunItem {
  cases: TestRunCaseItem[];
}

/** 执行记录列表 */
export function getTestRuns(projectId: number, params?: TestRunListParams) {
  return Alova.Get<TestRunListResult>(`/v1/projects/${projectId}/test-runs`, { params });
}

/** 执行记录详情（逐用例状态/耗时/失败信息） */
export function getTestRunDetail(id: number) {
  return Alova.Get<TestRunDetail>(`/v1/test-runs/${id}`);
}

/** 删除执行记录（owner/maintainer） */
export function deleteTestRun(id: number) {
  return Alova.Delete(`/v1/test-runs/${id}`);
}

// ==================== 失败闭环 / Flaky / 趋势（#506） ====================

/** Flaky 用例（近期窗口内 pass/fail 混现） */
export interface TestFlakyItem {
  externalKey: string;
  title: string;
  passCount: number;
  failCount: number;
  lastStatus: string;
  lastRunId: number;
}

export interface TestFlakyResult {
  window: number;
  list: TestFlakyItem[];
}

/** 趋势：近期执行 + 失败 Top */
export interface TestTrendRun {
  id: number;
  source: string;
  branch: string;
  total: number;
  passed: number;
  failed: number;
  errors: number;
  skipped: number;
  durationMs: number;
  finishedAt: string;
}

export interface TestFailTop {
  externalKey: string;
  title: string;
  failCount: number;
  totalCount: number;
}

export interface TestTrendResult {
  runs: TestTrendRun[];
  topFailed: TestFailTop[];
}

/** 失败用例转缺陷（TestRunCaseItem 补字段见上） */
export interface TestRunCaseBugResult {
  taskId: number;
  created: boolean;
}

/** Flaky 用例列表 */
export function getTestFlaky(projectId: number, window = 10) {
  return Alova.Get<TestFlakyResult>(`/v1/projects/${projectId}/test-runs/flaky`, {
    params: { window },
  });
}

/** 测试趋势（近期执行 + 失败 Top） */
export function getTestTrends(projectId: number, limit = 30) {
  return Alova.Get<TestTrendResult>(`/v1/projects/${projectId}/test-runs/trends`, {
    params: { limit },
  });
}

/** 失败用例转缺陷任务（建任务并挂接；taskId 非 0 时挂接既有） */
export function caseToBug(
  runCaseId: number,
  data: { title?: string; priority?: number; taskId?: number }
) {
  return Alova.Post<TestRunCaseBugResult>(`/v1/test-run-cases/${runCaseId}/bug`, data);
}
