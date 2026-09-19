<template>
  <div>
    <!-- 头部统计卡：用例规模 / 执行覆盖 / 近期通过率 / Flaky -->
    <n-card :bordered="false" class="proCard" title="测试总览">
      <n-spin :show="statsLoading">
        <div class="stat-cards" data-test-id="overview.stat-cards">
          <div class="stat-card">
            <div class="stat-num">{{ stats.caseTotal }}</div>
            <div class="stat-label">测试用例</div>
          </div>
          <div class="stat-card">
            <div class="stat-num">{{ stats.executedCases }}</div>
            <div class="stat-label">有执行的用例</div>
          </div>
          <div class="stat-card">
            <div class="stat-num">{{ stats.executions }}</div>
            <div class="stat-label">累计执行次数</div>
          </div>
          <div class="stat-card">
            <div class="stat-num" :class="passRateClass">{{ stats.passRateText }}</div>
            <div class="stat-label">近期通过率（近 {{ stats.runCount }} 次）</div>
          </div>
          <div class="stat-card">
            <div class="stat-num" :class="{ 'text-red-500': stats.flaky > 0 }">{{ stats.flaky }}</div>
            <div class="stat-label">Flaky 用例</div>
          </div>
        </div>
      </n-spin>
    </n-card>

    <!-- 趋势：通过率折线 + 失败数柱（双 y 轴），下挂失败 Top -->
    <n-card :bordered="false" title="测试趋势" class="proCard">
      <n-spin :show="trendLoading">
        <div
          ref="trendChartRef"
          style="width: 100%; height: 220px"
          v-if="trendRuns.length > 0"
          data-test-id="overview.trend-chart"
        ></div>
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

    <!-- Flaky：近期窗口内状态抖动的用例 -->
    <n-card :bordered="false" title="Flaky 用例" class="proCard" data-test-id="overview.flaky-panel">
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
                <n-tag :type="RUN_CASE_STATUS.tagType(f.lastStatus)" size="small">
                  {{ RUN_CASE_STATUS.label(f.lastStatus) }}
                </n-tag>
              </td>
              <td>run #{{ f.lastRunId }}</td>
              <td>
                <n-button text type="info" size="small" @click="goRun(f.lastRunId)">查看执行</n-button>
              </td>
            </tr>
          </tbody>
        </n-table>
      </n-spin>
    </n-card>
  </div>
</template>

<script lang="ts" setup>
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { RUN_CASE_STATUS, TEST_RUN_SOURCE } from '@/enums/test';
  import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import echarts from '@/utils/lib/echarts';
  import { getTestCases, getTestCaseStats, getTestFlaky, getTestTrends } from '@/api/test/index';
  import type { TestFlakyItem, TestTrendRun, TestFailTop } from '@/api/test/index';

  const route = useRoute();
  const router = useRouter();
  const projectId = computed(() => Number(route.params.projectId));

  function fmtDuration(ms: number): string {
    if (!ms || ms <= 0) return '-';
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
    const m = Math.floor(ms / 60000);
    const s = Math.round((ms % 60000) / 1000);
    return `${m}m${String(s).padStart(2, '0')}s`;
  }

  function goRun(runId: number) {
    router.push(`/project/${projectId.value}/test-runs/${runId}`);
  }

  // ---------- 头部统计 ----------
  const statsLoading = ref(false);
  const stats = ref({
    caseTotal: 0,
    executedCases: 0,
    executions: 0,
    passRateText: '-',
    runCount: 0,
    flaky: 0,
  });
  const passRateClass = computed(() => {
    const t = stats.value.passRateText;
    if (t === '-') return '';
    return Number(t.replace('%', '')) >= 80 ? 'text-green-500' : 'text-red-500';
  });

  async function loadStats() {
    statsLoading.value = true;
    try {
      const [casesRes, statsRes, trendRes, flakyRes] = await Promise.all([
        getTestCases(projectId.value, { pageNum: 1, pageSize: 1 }).catch(() => null),
        getTestCaseStats(projectId.value).catch(() => null),
        getTestTrends(projectId.value, 30).catch(() => null),
        getTestFlaky(projectId.value, 10).catch(() => null),
      ]);
      const s = { caseTotal: 0, executedCases: 0, executions: 0, passRateText: '-', runCount: 0, flaky: 0 };
      s.caseTotal = casesRes?.total || 0;
      const statList = statsRes?.list || [];
      s.executedCases = statList.length;
      s.executions = statList.reduce((acc, it) => acc + (it.total || 0), 0);
      const runs = trendRes?.runs || [];
      const passed = runs.reduce((acc, r) => acc + (r.passed || 0), 0);
      const total = runs.reduce((acc, r) => acc + (r.total || 0), 0);
      s.runCount = runs.length;
      if (total > 0) s.passRateText = `${Math.round((passed / total) * 100)}%`;
      s.flaky = (flakyRes?.list || []).length;
      stats.value = s;
    } finally {
      statsLoading.value = false;
    }
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
            `run #${r.id}（${TEST_RUN_SOURCE.label(r.source)}${r.branch ? ` · ${r.branch}` : ''}）`,
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

  onMounted(() => {
    loadStats();
    loadTrends();
    loadFlaky();
  });

  onUnmounted(disposeTrendChart);
</script>

<style scoped>
  .stat-cards {
    display: flex;
    flex-wrap: wrap;
    gap: 16px;
  }
  .stat-card {
    min-width: 140px;
    flex: 1;
    padding: 12px 16px;
    border-radius: 6px;
    background: rgba(128, 128, 128, 0.06);
  }
  .stat-num {
    font-size: 26px;
    font-weight: 600;
    line-height: 1.3;
  }
  .stat-label {
    margin-top: 4px;
    font-size: 12px;
    color: rgba(128, 128, 128, 0.9);
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
