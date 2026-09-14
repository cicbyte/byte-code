<template>
  <div>
    <n-card :bordered="false" class="proCard">
      <template #header>
        使用分析
        <n-tag size="small" :bordered="false" class="ml-2">近 {{ days }} 天</n-tag>
      </template>
      <template #header-extra>
        <n-space :size="8" align="center">
          <n-radio-group v-model:value="days" size="small" @update:value="load">
            <n-radio-button :value="1">今天</n-radio-button>
            <n-radio-button :value="7">7 天</n-radio-button>
            <n-radio-button :value="30">30 天</n-radio-button>
          </n-radio-group>
          <n-button size="small" @click="load">刷新</n-button>
        </n-space>
      </template>

      <!-- 总量概览：cli/web 流量构成（纯文本插值——数据页不依赖 rAF 动画，
           后台标签页/rAF 挂起环境也能精确显示） -->
      <n-grid cols="1 s:3" responsive="screen" :x-gap="12" class="mb-4">
        <n-grid-item>
          <n-statistic label="总调用" :value="overview.totalCalls" />
        </n-grid-item>
        <n-grid-item>
          <n-statistic label="CLI（agent）" :value="overview.cliCalls">
            <template #suffix>
              <span class="text-xs text-gray-400">{{ pct(overview.cliCalls) }}</span>
            </template>
          </n-statistic>
        </n-grid-item>
        <n-grid-item>
          <n-statistic label="Web（人类）" :value="overview.webCalls">
            <template #suffix>
              <span class="text-xs text-gray-400">{{ pct(overview.webCalls) }}</span>
            </template>
          </n-statistic>
        </n-grid-item>
      </n-grid>

      <!-- 命令热度：调试视角（哪个端点慢/常错/谁在用） -->
      <n-h6>命令热度</n-h6>
      <n-data-table
        :columns="heatColumns"
        :data="overview.endpoints"
        :loading="loading"
        :row-key="(row: UsageEndpointStat) => row.method + row.endpoint"
        size="small"
        :max-height="420"
      />

      <!-- 错误 TopN：排障入口 -->
      <n-h6 class="mt-4">错误 Top 20</n-h6>
      <n-data-table
        :columns="errColumns"
        :data="overview.errors"
        :loading="loading"
        :row-key="(row: UsageErrorItem) => row.method + row.endpoint + row.errorCode + row.statusCode"
        size="small"
        :max-height="320"
      />
    </n-card>

    <!-- 分析报告（#426）：趋势 / CLI 漏斗 / 版本分布 / 任务效率 -->
    <n-card :bordered="false" class="proCard mt-4" :loading="reportLoading">
      <template #header>
        分析报告
        <n-tag size="small" :bordered="false" class="ml-2">近 {{ days }} 天</n-tag>
      </template>
      <n-spin :show="reportLoading">
        <!-- 每日趋势：usage_daily 物化（明细 30 天清理，趋势长期保留） -->
        <n-h6>每日调用趋势</n-h6>
        <div ref="trendRef" class="trend-chart"></div>
        <div v-if="!reportLoading && report.daily.length === 0" class="text-xs text-gray-400">
          窗口内暂无调用记录
        </div>

        <n-grid cols="1 s:2" responsive="screen" :x-gap="24" class="mt-4">
          <!-- CLI 工作流漏斗：去重账号数 -->
          <n-grid-item>
            <n-h6>CLI 工作流漏斗（去重账号）</n-h6>
            <div v-if="report.funnel.length === 0" class="text-xs text-gray-400">窗口内无 CLI 流量</div>
            <div v-for="s in report.funnel" :key="s.step" class="funnel-row">
              <span class="funnel-label">{{ s.step }}. {{ s.label }}</span>
              <div class="funnel-track">
                <div class="funnel-bar" :style="{ width: funnelPct(s.actors) }"></div>
              </div>
              <span class="funnel-num">{{ s.actors }}</span>
            </div>
            <div class="text-xs text-gray-400 mt-1">
              会话 → 看任务 → 认领 → 完成的账号到达数；未识别版本不影响漏斗统计
            </div>
          </n-grid-item>

          <!-- CLI 版本分布：UA 解析 bcode/x.y.z，空=未识别 -->
          <n-grid-item>
            <n-h6>CLI 版本分布</n-h6>
            <n-table v-if="report.versions.length" :bordered="false" :single-line="false" size="small">
              <thead>
                <tr><th>版本</th><th>调用量</th><th>账号数</th><th>最近使用</th></tr>
              </thead>
              <tbody>
                <tr v-for="v in report.versions" :key="v.version">
                  <td>
                    <n-tag size="small" :bordered="false" :type="v.version === '未识别' ? 'warning' : 'success'">
                      {{ v.version }}
                    </n-tag>
                  </td>
                  <td>{{ v.calls }}</td>
                  <td>{{ v.actors }}</td>
                  <td>{{ (v.lastSeen || '').slice(0, 16) }}</td>
                </tr>
              </tbody>
            </n-table>
            <div v-else class="text-xs text-gray-400">窗口内无 CLI 流量</div>
            <div class="text-xs text-gray-400 mt-1">「未识别」= CLI UA 未带版本号（旧版客户端）</div>
          </n-grid-item>
        </n-grid>

        <!-- 任务效率：交付周期 + 人均表 -->
        <n-h6 class="mt-4">任务效率</n-h6>
        <n-grid cols="1 s:3" responsive="screen" :x-gap="12" class="mb-3">
          <n-grid-item>
            <n-statistic label="完成任务" :value="report.efficiency.completedTotal" />
          </n-grid-item>
          <n-grid-item>
            <n-statistic label="平均交付周期" :value="report.efficiency.avgLeadHours">
              <template #suffix><span class="text-xs text-gray-400">小时</span></template>
            </n-statistic>
          </n-grid-item>
          <n-grid-item>
            <n-statistic label="阻塞上报" :value="report.efficiency.blockedTotal" />
          </n-grid-item>
        </n-grid>
        <n-data-table
          :columns="effColumns"
          :data="report.efficiency.byActor"
          :loading="reportLoading"
          :row-key="(row: UsageActorEfficiency) => row.assigneeId"
          size="small"
          :max-height="320"
        />
      </n-spin>
    </n-card>
  </div>
</template>

<script lang="ts" setup>
  import { ref, reactive, h, onMounted, onUnmounted, nextTick } from 'vue';
  import { NTag, useMessage } from 'naive-ui';
  import type { DataTableColumns } from 'naive-ui';
  import echarts from '@/utils/lib/echarts';
  import { getUsageOverview, getUsageReport } from '@/api/platform/index';
  import type {
    UsageEndpointStat,
    UsageErrorItem,
    UsageOverviewResult,
    UsageReportResult,
    UsageActorEfficiency,
  } from '@/api/platform/index';

  const message = useMessage();
  const loading = ref(false);
  const days = ref(7);
  const overview = reactive<UsageOverviewResult>({
    windowDays: 7,
    totalCalls: 0,
    cliCalls: 0,
    webCalls: 0,
    endpoints: [],
    errors: [],
  });

  const reportLoading = ref(false);
  const report = reactive<UsageReportResult>({
    windowDays: 7,
    daily: [],
    funnel: [],
    versions: [],
    efficiency: { completedTotal: 0, avgLeadHours: 0, blockedTotal: 0, byActor: [] },
  });

  function pct(n: number): string {
    if (!overview.totalCalls) return '0%';
    return Math.round((n / overview.totalCalls) * 100) + '%';
  }

  // 漏斗条宽：相对第一步（第一步 100%）
  function funnelPct(actors: number): string {
    const first = report.funnel[0]?.actors || 0;
    if (!first || !actors) return '0%';
    return Math.max(2, Math.round((actors / first) * 100)) + '%';
  }

  async function load() {
    loading.value = true;
    reportLoading.value = true;
    try {
      const [ov, rp] = await Promise.all([getUsageOverview(days.value), getUsageReport(days.value)]);
      if (ov) {
        Object.assign(overview, ov);
        overview.endpoints = ov.endpoints || [];
        overview.errors = ov.errors || [];
      }
      if (rp) {
        Object.assign(report, rp);
        report.daily = rp.daily || [];
        report.funnel = rp.funnel || [];
        report.versions = rp.versions || [];
        report.efficiency = rp.efficiency || { completedTotal: 0, avgLeadHours: 0, blockedTotal: 0, byActor: [] };
        await nextTick();
        initTrend();
      }
    } catch {
      message.error('加载使用统计失败');
    } finally {
      loading.value = false;
      reportLoading.value = false;
    }
  }

  // ---- 趋势图（echarts 复用：折线两系列 cli/web；清理模式对齐 console） ----
  const trendRef = ref<HTMLDivElement | null>(null);
  let chart: echarts.ECharts | null = null;

  function handleChartResize() {
    chart?.resize();
  }

  function disposeChart() {
    window.removeEventListener('resize', handleChartResize);
    chart?.dispose();
    chart = null;
  }

  function initTrend() {
    if (!trendRef.value) return;
    disposeChart();
    chart = echarts.init(trendRef.value);
    chart.setOption({
      tooltip: { trigger: 'axis' },
      legend: { data: ['CLI', 'Web'] },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: { type: 'category', data: report.daily.map((d) => d.day.slice(5)) },
      yAxis: { type: 'value' },
      series: [
        { name: 'CLI', type: 'line', smooth: true, data: report.daily.map((d) => d.cliCalls) },
        { name: 'Web', type: 'line', smooth: true, data: report.daily.map((d) => d.webCalls) },
      ],
    });
    window.addEventListener('resize', handleChartResize);
  }

  function methodTagType(m: string): 'success' | 'info' | 'warning' | 'error' | 'default' {
    return ({ GET: 'success', POST: 'info', PUT: 'warning', DELETE: 'error' } as const)[m] || 'default';
  }

  const heatColumns: DataTableColumns<UsageEndpointStat> = [
    {
      title: '方法',
      key: 'method',
      width: 70,
      render: (row) => h(NTag, { size: 'tiny', type: methodTagType(row.method), bordered: false }, () => row.method),
    },
    { title: '端点', key: 'endpoint', ellipsis: { tooltip: true } },
    {
      title: '调用量',
      key: 'total',
      width: 90,
      sorter: (a, b) => a.total - b.total,
      render: (row) =>
        h('span', {}, [
          h('span', {}, String(row.total)),
          h('span', { class: 'text-xs ml-1' }, `(CLI ${row.cliCount} / Web ${row.webCount})`),
        ]),
    },
    {
      title: '失败率',
      key: 'errorRate',
      width: 90,
      sorter: (a, b) => a.errorRate - b.errorRate,
      render: (row) =>
        row.errCount === 0
          ? h('span', { class: 'text-gray-400' }, '-')
          : h(
              NTag,
              { size: 'small', type: row.errorRate > 0.2 ? 'error' : 'warning' },
              () => (row.errorRate * 100).toFixed(1) + '%',
            ),
    },
    { title: 'P50', key: 'p50ms', width: 80, sorter: (a, b) => a.p50ms - b.p50ms, render: (r) => `${r.p50ms}ms` },
    {
      title: 'P95',
      key: 'p95ms',
      width: 80,
      sorter: (a, b) => a.p95ms - b.p95ms,
      render: (row) =>
        h('span', { class: row.p95ms > 1000 ? 'text-red-500' : '' }, `${row.p95ms}ms`),
    },
    { title: '均值', key: 'avgMs', width: 80, render: (r) => `${r.avgMs}ms` },
    { title: '最近使用', key: 'lastUsed', width: 150, render: (r) => (r.lastUsed || '').slice(0, 16) },
  ];

  const errColumns: DataTableColumns<UsageErrorItem> = [
    { title: '次数', key: 'count', width: 70, sorter: (a, b) => a.count - b.count },
    {
      title: '业务码',
      key: 'errorCode',
      width: 80,
      render: (row) => h(NTag, { size: 'small', type: 'error', bordered: false }, () => String(row.errorCode || row.statusCode)),
    },
    { title: '方法', key: 'method', width: 70 },
    { title: '端点', key: 'endpoint', ellipsis: { tooltip: true } },
    { title: '最近发生', key: 'lastSeen', width: 150, render: (r) => (r.lastSeen || '').slice(0, 16) },
  ];

  const effColumns: DataTableColumns<UsageActorEfficiency> = [
    { title: '账号', key: 'assigneeName', width: 160 },
    {
      title: '类型',
      key: 'actorType',
      width: 80,
      render: (row) =>
        h(
          NTag,
          { size: 'small', bordered: false, type: row.actorType === 'ai' ? 'info' : 'default' },
          () => (row.actorType === 'ai' ? 'Agent' : '人类'),
        ),
    },
    { title: '完成任务', key: 'completed', width: 100, sorter: (a, b) => a.completed - b.completed },
    {
      title: '平均交付周期',
      key: 'avgLeadHours',
      width: 130,
      sorter: (a, b) => a.avgLeadHours - b.avgLeadHours,
      render: (r) => (r.avgLeadHours ? `${r.avgLeadHours} 小时` : '-'),
    },
    {
      title: '阻塞上报',
      key: 'blocked',
      width: 100,
      sorter: (a, b) => a.blocked - b.blocked,
      render: (r) => (r.blocked ? String(r.blocked) : '-'),
    },
  ];

  onMounted(load);
  onUnmounted(disposeChart);
</script>

<style lang="less" scoped>
  .trend-chart {
    width: 100%;
    height: 260px;
  }

  .funnel-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;

    .funnel-label {
      width: 110px;
      font-size: 12px;
      color: var(--n-text-color-2, #666);
      flex-shrink: 0;
    }

    .funnel-track {
      flex: 1;
      height: 18px;
      background: var(--n-color-disabled, #f3f3f6);
      border-radius: 4px;
      overflow: hidden;
    }

    .funnel-bar {
      height: 100%;
      background: linear-gradient(90deg, #1890ff, #69c0ff);
      border-radius: 4px;
      transition: width 0.3s;
    }

    .funnel-num {
      width: 40px;
      text-align: right;
      font-size: 12px;
      flex-shrink: 0;
    }
  }
</style>
