<template>
  <div>
    <!-- 趋势总览：通过率折线 + 失败数柱（双 y 轴），下挂失败 Top -->
    <n-card :bordered="false" title="测试趋势" class="proCard" v-if="trendRuns.length > 0 || trendLoading">
      <n-spin :show="trendLoading">
        <div ref="trendChartRef" style="width: 100%; height: 220px" v-if="trendRuns.length > 0"></div>
        <n-empty description="暂无执行数据" v-else-if="!trendLoading" style="padding: 24px 0" />
        <div class="mt-2 flex flex-wrap gap-4" v-if="topFailed.length > 0">
          <div style="min-width: 320px; flex: 1">
            <div class="text-xs text-gray-400 mb-1">失败 Top（近期窗口）</div>
            <div v-for="tf in topFailed" :key="tf.externalKey" class="fail-top-row">
              <span class="fail-top-name" :title="tf.externalKey">{{ tf.title || tf.externalKey }}</span>
              <div class="fail-top-bar-wrap">
                <div class="fail-top-bar" :style="{ width: failTopWidth(tf) }"></div>
              </div>
              <span class="text-xs text-gray-500" style="flex-shrink: 0">{{ tf.failCount }}/{{ tf.totalCount }}</span>
            </div>
          </div>
        </div>
      </n-spin>
    </n-card>

    <n-card :bordered="false" class="proCard">
      <n-tabs type="line" v-model:value="activeTab">
        <n-tab-pane name="runs" tab="执行记录">
          <template #tab>
            执行记录
          </template>
          <n-space size="small" align="center" justify="space-between" class="mb-2">
            <n-space size="small" align="center">
              <n-select
                v-model:value="filters.source"
                :options="sourceOptions"
                size="small"
                style="width: 110px"
                placeholder="来源"
                @update:value="reload"
              />
              <n-select
                v-model:value="filters.status"
                :options="statusOptions"
                size="small"
                style="width: 100px"
                placeholder="结果"
                @update:value="reload"
              />
              <n-input
                v-model:value="filters.branch"
                size="small"
                style="width: 140px"
                placeholder="分支"
                clearable
                @keyup.enter="reload"
                @clear="reload"
              />
            </n-space>
          </n-space>

          <n-spin :show="loading">
            <EmptyState
              type="doc"
              title="暂无执行记录"
              description="通过 byte-code-pytest 插件（pytest --bcode）、bcode test --upload 或 CI 上报到此处"
              v-if="!loading && runList.length === 0"
            />
            <n-table v-else :bordered="false" :single-line="false" size="small">
              <thead>
                <tr>
                  <th class="col-idx">#</th>
                  <th>来源</th>
                  <th>分支 / Commit</th>
                  <th>环境</th>
                  <th>触发者</th>
                  <th>结果</th>
                  <th>耗时</th>
                  <th>结束时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(item, __ix) in runList" :key="item.id">
                  <td class="col-idx">{{ __ix + 1 }}</td>
                  <td>
                    <n-tag :type="sourceTagType(item.source)" size="small">{{ sourceLabel(item.source) }}</n-tag>
                  </td>
                  <td>
                    <span v-if="item.branch || item.gitSha">
                      {{ item.branch || '-' }}<span v-if="item.gitSha" class="text-gray-400"> @{{ item.gitSha.slice(0, 8) }}</span>
                    </span>
                    <span v-else>-</span>
                  </td>
                  <td>{{ item.env || '-' }}</td>
                  <td>{{ item.triggeredByName || '-' }}</td>
                  <td>
                    <n-space size="small" :wrap="false">
                      <n-tag type="success" size="small">{{ item.passed }} 通过</n-tag>
                      <n-tag v-if="item.failed" type="error" size="small">{{ item.failed }} 失败</n-tag>
                      <n-tag v-if="item.errors" type="warning" size="small">{{ item.errors }} 错误</n-tag>
                      <n-tag v-if="item.skipped" size="small">{{ item.skipped }} 跳过</n-tag>
                    </n-space>
                  </td>
                  <td>{{ fmtDuration(item.durationMs) }}</td>
                  <td>{{ item.finishedAt || item.createdAt }}</td>
                  <td>
                    <n-space size="small">
                      <n-button text type="info" @click="openDetail(item)">详情</n-button>
                      <n-button text type="error" @click="handleDelete(item)">删除</n-button>
                    </n-space>
                  </td>
                </tr>
              </tbody>
            </n-table>
          </n-spin>

          <div class="mt-4 flex justify-end" v-if="total > pagination.pageSize">
            <n-pagination
              v-model:page="pagination.page"
              :page-size="pagination.pageSize"
              :item-count="total"
              @update:page="loadData"
            />
          </div>
        </n-tab-pane>

        <!-- Flaky：近期窗口内状态抖动的用例 -->
        <n-tab-pane name="flaky" :tab="`Flaky 用例${flakyList.length ? `（${flakyList.length}）` : ''}`">
          <n-spin :show="flakyLoading">
            <EmptyState
              type="doc"
              title="近期无 Flaky 用例"
              description="同一测试在最近多次执行中既通过又失败时会出现在这里"
              v-if="!flakyLoading && flakyList.length === 0"
            />
            <n-table v-else :bordered="false" :single-line="false" size="small">
              <thead>
                <tr>
                  <th>用例</th>
                  <th>通过次数</th>
                  <th>失败/错误</th>
                  <th>最近状态</th>
                  <th>最近执行</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="f in flakyList" :key="f.externalKey">
                  <td>
                    <div class="text-xs" style="word-break: break-all">{{ f.title || f.externalKey }}</div>
                    <div class="text-gray-400 text-xs" style="word-break: break-all">{{ f.externalKey }}</div>
                  </td>
                  <td><n-tag type="success" size="small">{{ f.passCount }}</n-tag></td>
                  <td><n-tag type="error" size="small">{{ f.failCount }}</n-tag></td>
                  <td>
                    <n-tag :type="caseTagType(f.lastStatus)" size="small">{{ caseLabel(f.lastStatus) }}</n-tag>
                  </td>
                  <td>run #{{ f.lastRunId }}</td>
                  <td>
                    <n-button text type="info" size="small" @click="openRunById(f.lastRunId)">查看执行</n-button>
                  </td>
                </tr>
              </tbody>
            </n-table>
          </n-spin>
        </n-tab-pane>
      </n-tabs>
    </n-card>

    <!-- 执行详情：汇总 + 逐用例结果（失败可转缺陷） -->
    <n-modal
      v-model:show="showDetailModal"
      preset="card"
      :title="`执行记录 #${currentRun?.id || ''}`"
      style="width: 980px; max-height: 85vh"
    >
      <n-spin :show="detailLoading">
        <template v-if="runDetail">
          <n-space class="mb-4" :size="24" align="center">
            <n-statistic label="总计" :value="runDetail.total" />
            <n-statistic label="通过" :value="runDetail.passed">
              <template #suffix><span class="text-green-500 text-xs">passed</span></template>
            </n-statistic>
            <n-statistic label="失败" :value="runDetail.failed">
              <template #suffix><span class="text-red-500 text-xs">failed</span></template>
            </n-statistic>
            <n-statistic label="错误" :value="runDetail.errors">
              <template #suffix><span class="text-yellow-500 text-xs">error</span></template>
            </n-statistic>
            <n-statistic label="跳过" :value="runDetail.skipped" />
            <n-statistic label="耗时" :value="fmtDuration(runDetail.durationMs)" />
          </n-space>
          <n-space class="mb-3" size="small" align="center">
            <n-tag size="small">{{ sourceLabel(runDetail.source) }}</n-tag>
            <span v-if="runDetail.branch" class="text-gray-500 text-xs">{{ runDetail.branch }}</span>
            <span v-if="runDetail.gitSha" class="text-gray-400 text-xs">{{ runDetail.gitSha.slice(0, 8) }}</span>
            <span v-if="runDetail.startedAt" class="text-gray-400 text-xs">
              {{ runDetail.startedAt }} ~ {{ runDetail.finishedAt }}
            </span>
            <n-select
              v-model:value="caseFilter"
              :options="caseFilterOptions"
              size="small"
              style="width: 110px"
            />
          </n-space>
          <n-table :bordered="false" :single-line="false" size="small">
            <thead>
              <tr>
                <th style="width: 38%">用例</th>
                <th>映射平台用例</th>
                <th>状态</th>
                <th>耗时</th>
                <th>失败信息 / 缺陷</th>
              </tr>
            </thead>
            <tbody>
              <template v-for="c in filteredCases" :key="c.id">
                <tr>
                  <td>
                    <div class="text-xs" style="word-break: break-all">{{ c.title || c.externalKey }}</div>
                    <div v-if="c.externalKey && c.title" class="text-gray-400 text-xs" style="word-break: break-all">
                      {{ c.externalKey }}
                    </div>
                  </td>
                  <td>
                    <span v-if="c.testCaseId">{{ c.testCaseTitle || `#${c.testCaseId}` }}</span>
                    <span v-else class="text-gray-400">未映射</span>
                  </td>
                  <td>
                    <n-space size="small" :wrap="false">
                      <n-tag :type="caseTagType(c.status)" size="small">{{ caseLabel(c.status) }}</n-tag>
                      <n-tag v-if="c.flaky" type="warning" size="small" :bordered="false">flaky</n-tag>
                    </n-space>
                  </td>
                  <td>{{ fmtDuration(c.durationMs) }}</td>
                  <td>
                    <n-space size="small" align="center" :wrap="false">
                      <n-button v-if="c.message" text type="info" size="small" @click="toggleExpand(c.id)">
                        {{ expanded.has(c.id) ? '收起' : '查看' }}
                      </n-button>
                      <router-link
                        v-if="c.bugTaskId"
                        :to="{ name: 'project_tasks' }"
                        class="text-xs"
                        style="text-decoration: none"
                      >
                        <n-tag type="error" size="small" :bordered="false">缺陷 #{{ c.bugTaskId }}</n-tag>
                      </router-link>
                      <n-button
                        v-else-if="c.status === 'fail' || c.status === 'error'"
                        text
                        type="error"
                        size="small"
                        @click="openBugDialog(c)"
                      >
                        转缺陷
                      </n-button>
                      <span v-if="!c.message && !c.bugTaskId && c.status !== 'fail' && c.status !== 'error'" class="text-gray-400">-</span>
                    </n-space>
                  </td>
                </tr>
                <tr v-if="expanded.has(c.id)">
                  <td colspan="5">
                    <pre class="fail-msg">{{ c.message }}</pre>
                  </td>
                </tr>
              </template>
            </tbody>
          </n-table>
        </template>
      </n-spin>
    </n-modal>

    <!-- 转缺陷弹窗：建 bug 任务并挂接（未指派任务会广播给绑定 agent 认领） -->
    <n-modal
      v-model:show="showBugModal"
      preset="dialog"
      title="失败用例转缺陷任务"
      positive-text="创建并挂接"
      negative-text="取消"
      @positive-click="submitBug"
      style="width: 560px"
    >
      <n-form :model="bugForm" label-placement="left" :label-width="72" class="py-4">
        <n-form-item label="任务标题">
          <n-input v-model:value="bugForm.title" placeholder="缺省按用例标题生成" />
        </n-form-item>
        <n-form-item label="优先级">
          <n-select v-model:value="bugForm.priority" :options="prioOptions" size="small" />
        </n-form-item>
        <n-alert type="info" :bordered="false">
          任务描述将自动附带失败信息与执行上下文（分支/commit/run 入口）；未指派任务会广播给项目内已接入的 agent 认领，修复后
          bcode test --run 重跑即形成验证闭环。
        </n-alert>
      </n-form>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { ref, reactive, computed, onMounted, onUnmounted, nextTick } from 'vue';
  import { useRoute } from 'vue-router';
  import { useMessage, useDialog } from 'naive-ui';
  import echarts from '@/utils/lib/echarts';
  import {
    getTestRuns,
    getTestRunDetail,
    deleteTestRun,
    getTestFlaky,
    getTestTrends,
    caseToBug,
  } from '@/api/test/index';
  import type { TestRunItem, TestRunDetail, TestFlakyItem, TestTrendRun, TestFailTop } from '@/api/test/index';

  const message = useMessage();
  const dialog = useDialog();
  const route = useRoute();
  const projectId = computed(() => Number(route.params.projectId));

  // ---------- 字典 ----------
  const sourceOptions = [
    { label: '全部来源', value: '' },
    { label: 'pytest', value: 'pytest' },
    { label: 'CI', value: 'ci' },
    { label: 'JUnit', value: 'junit' },
    { label: '手工', value: 'manual' },
  ];
  const statusOptions = [
    { label: '全部结果', value: '' },
    { label: '通过', value: 'pass' },
    { label: '失败', value: 'fail' },
  ];
  const prioOptions = [
    { label: 'P1 紧急', value: 1 },
    { label: 'P2 高', value: 2 },
    { label: 'P3 中', value: 3 },
    { label: 'P4 低', value: 4 },
  ];

  const sourceLabels: Record<string, string> = { manual: '手工', pytest: 'pytest', ci: 'CI', junit: 'JUnit' };
  function sourceLabel(s: string) {
    return sourceLabels[s] || s;
  }
  function sourceTagType(s: string): 'success' | 'info' | 'warning' | 'default' {
    if (s === 'pytest') return 'success';
    if (s === 'ci') return 'info';
    if (s === 'junit') return 'warning';
    return 'default';
  }

  const caseLabels: Record<string, string> = { pass: '通过', fail: '失败', error: '错误', skip: '跳过' };
  function caseLabel(s: string) {
    return caseLabels[s] || s;
  }
  function caseTagType(s: string): 'success' | 'error' | 'warning' | 'default' {
    if (s === 'pass') return 'success';
    if (s === 'fail') return 'error';
    if (s === 'error') return 'warning';
    return 'default';
  }

  function fmtDuration(ms: number): string {
    if (!ms || ms <= 0) return '-';
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
    const m = Math.floor(ms / 60000);
    const s = Math.round((ms % 60000) / 1000);
    return `${m}m${String(s).padStart(2, '0')}s`;
  }

  // ---------- 执行记录列表 ----------
  const activeTab = ref('runs');
  const loading = ref(false);
  const runList = ref<TestRunItem[]>([]);
  const total = ref(0);
  const pagination = reactive({ page: 1, pageSize: 10 });
  const filters = reactive({ source: null as string | null, status: null as string | null, branch: '' });

  async function loadData() {
    loading.value = true;
    try {
      const res = await getTestRuns(projectId.value, {
        pageNum: pagination.page,
        pageSize: pagination.pageSize,
        source: filters.source || undefined,
        status: filters.status || undefined,
        branch: filters.branch || undefined,
      });
      if (res) {
        runList.value = res.list || [];
        total.value = res.total || 0;
      }
    } catch (e) {
      // ignore
    } finally {
      loading.value = false;
    }
  }

  function reload() {
    pagination.page = 1;
    loadData();
  }

  // ---------- 趋势 ----------
  const trendLoading = ref(false);
  const trendRuns = ref<TestTrendRun[]>([]);
  const topFailed = ref<TestFailTop[]>([]);
  const trendChartRef = ref<HTMLElement | null>(null);
  let trendChart: echarts.ECharts | null = null;

  function disposeTrendChart() {
    window.removeEventListener('resize', resizeTrendChart);
    trendChart?.dispose();
    trendChart = null;
  }

  function resizeTrendChart() {
    trendChart?.resize();
  }

  function renderTrend() {
    if (!trendChartRef.value || trendRuns.value.length === 0) return;
    disposeTrendChart();
    trendChart = echarts.init(trendChartRef.value);
    const runs = trendRuns.value;
    trendChart.setOption({
      tooltip: {
        trigger: 'axis',
        formatter(ps: any) {
          const p = Array.isArray(ps) ? ps[0] : ps;
          const r = runs[p.dataIndex];
          if (!r) return '';
          const rate = r.total > 0 ? Math.round((r.passed / r.total) * 100) : 0;
          return [
            `run #${r.id}（${sourceLabel(r.source)}${r.branch ? ` · ${r.branch}` : ''}）`,
            `通过率 ${rate}%（${r.passed}/${r.total}）`,
            `失败 ${r.failed} · 错误 ${r.errors} · 跳过 ${r.skipped}`,
            `耗时 ${fmtDuration(r.durationMs)}`,
            r.finishedAt,
          ].join('<br/>');
        },
      },
      legend: { data: ['通过率', '失败+错误'], top: 0 },
      grid: { left: 44, right: 44, top: 56, bottom: 28 },
      xAxis: {
        type: 'category',
        data: runs.map((r) => `#${r.id}`),
        axisLabel: { fontSize: 10 },
      },
      yAxis: [
        { type: 'value', name: '通过率', min: 0, max: 100, axisLabel: { formatter: '{value}%' } },
        { type: 'value', name: '失败', minInterval: 1 },
      ],
      series: [
        {
          name: '通过率',
          type: 'line',
          smooth: true,
          symbolSize: 5,
          yAxisIndex: 0,
          data: runs.map((r) => (r.total > 0 ? Math.round((r.passed / r.total) * 100) : 0)),
          itemStyle: { color: '#18a058' },
          areaStyle: { color: 'rgba(24,160,88,0.10)' },
        },
        {
          name: '失败+错误',
          type: 'bar',
          yAxisIndex: 1,
          barMaxWidth: 14,
          data: runs.map((r) => r.failed + r.errors),
          itemStyle: { color: 'rgba(208,48,80,0.55)' },
        },
      ],
    });
    window.addEventListener('resize', resizeTrendChart);
  }

  function failTopWidth(tf: TestFailTop): string {
    const max = topFailed.value[0]?.failCount || 1;
    return `${Math.max(6, Math.round((tf.failCount / max) * 100))}%`;
  }

  async function loadTrends() {
    trendLoading.value = true;
    try {
      const res = await getTestTrends(projectId.value, 30);
      if (res) {
        trendRuns.value = res.runs || [];
        topFailed.value = res.topFailed || [];
      }
      await nextTick();
      renderTrend();
    } catch (e) {
      // ignore
    } finally {
      trendLoading.value = false;
    }
  }

  // ---------- Flaky 面板 ----------
  const flakyLoading = ref(false);
  const flakyList = ref<TestFlakyItem[]>([]);

  async function loadFlaky() {
    flakyLoading.value = true;
    try {
      const res = await getTestFlaky(projectId.value, 10);
      if (res) {
        flakyList.value = res.list || [];
      }
    } catch (e) {
      // ignore
    } finally {
      flakyLoading.value = false;
    }
  }

  // ---------- 详情 ----------
  const showDetailModal = ref(false);
  const detailLoading = ref(false);
  const currentRun = ref<TestRunItem | null>(null);
  const runDetail = ref<TestRunDetail | null>(null);
  const expanded = ref(new Set<number>());
  const caseFilter = ref('all');

  const caseFilterOptions = [
    { label: '全部用例', value: 'all' },
    { label: '仅失败/错误', value: 'bad' },
    { label: '仅通过', value: 'pass' },
  ];

  const filteredCases = computed(() => {
    if (!runDetail.value) return [];
    if (caseFilter.value === 'bad') return runDetail.value.cases.filter((c) => c.status === 'fail' || c.status === 'error');
    if (caseFilter.value === 'pass') return runDetail.value.cases.filter((c) => c.status === 'pass');
    return runDetail.value.cases;
  });

  function toggleExpand(id: number) {
    if (expanded.value.has(id)) {
      expanded.value.delete(id);
    } else {
      expanded.value.add(id);
    }
    // 触发响应式更新（Set 原地变更）
    expanded.value = new Set(expanded.value);
  }

  async function openDetail(item: TestRunItem) {
    currentRun.value = item;
    showDetailModal.value = true;
    await loadDetail(item.id);
  }

  async function openRunById(runId: number) {
    showDetailModal.value = true;
    currentRun.value = null;
    await loadDetail(runId);
  }

  async function loadDetail(runId: number) {
    detailLoading.value = true;
    runDetail.value = null;
    expanded.value = new Set();
    try {
      const det = await getTestRunDetail(runId);
      runDetail.value = det;
      // Flaky 面板跳转路径没有列表行对象，标题的 run 号从详情回填
      if (!currentRun.value && det) {
        currentRun.value = det;
      }
      caseFilter.value = det && det.failed + det.errors > 0 ? 'bad' : 'all';
    } catch (e) {
      message.error('加载执行详情失败');
    } finally {
      detailLoading.value = false;
    }
  }

  // ---------- 转缺陷 ----------
  const showBugModal = ref(false);
  const bugSubmitting = ref(false);
  const bugForm = reactive({ runCaseId: 0, title: '', priority: 2 });

  function openBugDialog(c: { id: number; title: string }) {
    bugForm.runCaseId = c.id;
    bugForm.title = '';
    bugForm.priority = 2;
    showBugModal.value = true;
  }

  async function submitBug() {
    if (bugSubmitting.value) return false;
    bugSubmitting.value = true;
    try {
      const res = await caseToBug(bugForm.runCaseId, {
        title: bugForm.title || undefined,
        priority: bugForm.priority,
      });
      message.success(`已创建缺陷任务 #${res.taskId}（已广播绑定 agent）`);
      // 详情刷新拿挂钩回显
      if (runDetail.value) {
        await loadDetail(runDetail.value.id);
      }
      loadTrends();
      return true;
    } catch (e: any) {
      message.error(e?.message || '转缺陷失败');
      return false;
    } finally {
      bugSubmitting.value = false;
    }
  }

  function handleDelete(item: TestRunItem) {
    dialog.warning({
      title: '删除执行记录',
      content: `确定删除执行记录 #${item.id}（${item.total} 用例）？该操作不可恢复。`,
      positiveText: '删除',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await deleteTestRun(item.id);
          message.success('已删除');
          loadData();
          loadTrends();
        } catch (e) {
          message.error('删除失败');
        }
      },
    });
  }

  onMounted(() => {
    loadData();
    loadTrends();
    loadFlaky();
  });

  onUnmounted(disposeTrendChart);
</script>

<style scoped>
  .fail-msg {
    margin: 0;
    padding: 8px 12px;
    max-height: 260px;
    overflow: auto;
    background: rgba(128, 128, 128, 0.08);
    border-radius: 4px;
    font-size: 12px;
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-all;
  }
  .fail-top-row {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 2px 0;
  }
  .fail-top-name {
    width: 40%;
    flex-shrink: 0;
    font-size: 12px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .fail-top-bar-wrap {
    flex: 1;
    height: 8px;
    border-radius: 4px;
    background: rgba(128, 128, 128, 0.12);
    overflow: hidden;
  }
  .fail-top-bar {
    height: 100%;
    border-radius: 4px;
    background: rgba(208, 48, 80, 0.75);
  }
</style>
