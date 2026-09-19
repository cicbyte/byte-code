// 测试域字典：用例状态/分类/优先级（cases.vue 此前全部组件内硬编码）
export const CASE_STATUS = {
  options: [
    { label: '草稿', value: 'draft' },
    { label: '启用', value: 'active' },
    { label: '废弃', value: 'deprecated' },
  ],
  label(s: string): string {
    return this.options.find((o) => o.value === s)?.label || s;
  },
  tagType(s: string): 'default' | 'success' | 'error' {
    if (s === 'active') return 'success';
    if (s === 'deprecated') return 'error';
    return 'default';
  },
};

export const CASE_CATEGORY_OPTIONS = [
  { label: '功能测试', value: 'functional' },
  { label: '性能测试', value: 'performance' },
  { label: '安全测试', value: 'security' },
  { label: '兼容性测试', value: 'compatibility' },
];

/** 用例优先级配色（P0-P3 字符串域，区别于任务域的数值 priorityTagType） */
export function casePriorityTagType(p: string): 'error' | 'warning' | 'info' | 'default' {
  if (p === 'P0') return 'error';
  if (p === 'P1') return 'warning';
  if (p === 'P2') return 'info';
  return 'default';
}

/** 执行记录用例状态（test_run_cases.status，pytest 四态） */
export const RUN_CASE_STATUS = {
  options: [
    { label: '通过', value: 'pass' },
    { label: '失败', value: 'fail' },
    { label: '错误', value: 'error' },
    { label: '跳过', value: 'skip' },
  ],
  label(s: string): string {
    return this.options.find((o) => o.value === s)?.label || s;
  },
  tagType(s: string): 'success' | 'error' | 'warning' | 'default' {
    if (s === 'pass') return 'success';
    if (s === 'fail') return 'error';
    if (s === 'error') return 'warning';
    return 'default';
  },
};

/** 执行记录来源（test_runs.source） */
export const TEST_RUN_SOURCE = {
  options: [
    { label: 'pytest', value: 'pytest' },
    { label: 'CI', value: 'ci' },
    { label: 'JUnit', value: 'junit' },
    { label: '手工', value: 'manual' },
  ],
  label(s: string): string {
    return this.options.find((o) => o.value === s)?.label || s;
  },
  tagType(s: string): 'success' | 'info' | 'warning' | 'default' {
    if (s === 'pytest') return 'success';
    if (s === 'ci') return 'info';
    if (s === 'junit') return 'warning';
    return 'default';
  },
};
