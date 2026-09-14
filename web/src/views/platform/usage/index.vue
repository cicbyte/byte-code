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
  </div>
</template>

<script lang="ts" setup>
  import { ref, reactive, h, onMounted } from 'vue';
  import { NTag, useMessage } from 'naive-ui';
  import type { DataTableColumns } from 'naive-ui';
  import { getUsageOverview } from '@/api/platform/index';
  import type { UsageEndpointStat, UsageErrorItem, UsageOverviewResult } from '@/api/platform/index';

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

  function pct(n: number): string {
    if (!overview.totalCalls) return '0%';
    return Math.round((n / overview.totalCalls) * 100) + '%';
  }

  async function load() {
    loading.value = true;
    try {
      const res = await getUsageOverview(days.value);
      if (res) {
        Object.assign(overview, res);
        overview.endpoints = res.endpoints || [];
        overview.errors = res.errors || [];
      }
    } catch {
      message.error('加载使用统计失败');
    } finally {
      loading.value = false;
    }
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

  onMounted(load);
</script>
