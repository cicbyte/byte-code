<template>
  <div>
    <n-card :bordered="false" :segmented="{ content: true }" class="mb-4" title="我的效率">
      <div class="stats-layout">
        <div class="stats-nums">
          <div class="stat-row">
            <n-statistic label="活跃任务" :value="stats?.activeTotal ?? 0" />
            <n-statistic label="已逾期" :value="stats?.overdue ?? 0" :value-style="stats?.overdue ? { color: '#d03050' } : undefined" />
            <n-statistic label="近30天完成" :value="stats?.completed30d ?? 0" />
            <n-statistic label="平均交付周期">
              <span class="stat-lead">{{ fmtLead(stats?.avgLeadHours) }}</span>
            </n-statistic>
          </div>
          <n-space size="small" align="center" class="stat-status">
            <n-tag
              v-for="sc in stats?.statusCounts || []"
              :key="sc.status"
              size="small"
              :bordered="false"
              :type="statusTagType(sc.status)"
            >
              {{ statusLabel(sc.status) }} {{ sc.count }}
            </n-tag>
          </n-space>
          <n-space v-if="stats?.byProject?.length" size="small" align="center" class="stat-status">
            <span class="stat-proj-label">活跃分布</span>
            <n-tag v-for="pr in stats.byProject" :key="pr.projectId" size="small" round>
              {{ pr.projectName || `#${pr.projectId}` }} · {{ pr.active }}
            </n-tag>
          </n-space>
        </div>
        <div v-show="stats" ref="trendRef" class="stats-chart" />
      </div>
    </n-card>

    <n-card :bordered="false" :segmented="{ content: true }">
      <template #header>
        我的任务
        <span class="header-hint">当前账号被指派的跨项目任务</span>
      </template>

      <!-- 过滤栏 -->
      <n-space class="mb-4" align="center">
        <n-radio-group v-model:value="statusFilter" size="small" @update:value="onFilterChange">
          <n-radio-button value="">未完成</n-radio-button>
          <n-radio-button value="all">全部</n-radio-button>
          <n-radio-button v-for="s in TASK_STATUS" :key="s.value" :value="s.value">{{ s.label }}</n-radio-button>
        </n-radio-group>
        <n-select
          v-model:value="projectId"
          :options="projectOptions"
          size="small"
          clearable
          placeholder="全部项目"
          style="width: 200px"
          @update:value="onFilterChange"
        />
        <n-input
          v-model:value="keyword"
          size="small"
          style="width: 220px"
          placeholder="标题 / 描述"
          clearable
          @keyup.enter="onFilterChange"
          @clear="onFilterChange"
        />
        <n-button size="small" @click="onFilterChange">查询</n-button>
      </n-space>

      <n-data-table
        :columns="columns"
        :data="list"
        :loading="loading"
        :row-key="(row: MyTaskItem) => row.id"
        size="small"
      />

      <div class="mt-4 flex justify-end" v-if="total > pagination.size">
        <n-pagination
          v-model:page="pagination.page"
          :page-size="pagination.size"
          :item-count="total"
          @update:page="load"
        />
      </div>
    </n-card>

    <TaskDetailModal ref="taskDetailRef" @updated="load" />
  </div>
</template>

<script lang="ts" setup>
  import { ref, reactive, computed, h, onMounted, onUnmounted, nextTick } from 'vue';
  import { useRouter } from 'vue-router';
  import { NButton, NTag, NSpace } from 'naive-ui';
  import type { DataTableColumns } from 'naive-ui';
  import { getMyTasks, getMyTaskStats, getProjects } from '@/api/project/index';
  import type { MyTaskItem, MyTaskStats } from '@/api/project/index';
  import echarts from '@/utils/lib/echarts';
  import TaskDetailModal from '@/views/project/components/TaskDetailModal.vue';
  import { dueTagType, dueLabel } from '@/utils/taskDue';
  import { TASK_STATUS, statusLabel, statusTagType, priorityTagType, typeTagType, typeLabel } from '@/enums/task';

  const router = useRouter();
  const loading = ref(false);
  const list = ref<MyTaskItem[]>([]);
  const total = ref(0);
  const pagination = reactive({ page: 1, size: 20 });
  const statusFilter = ref('');
  const projectId = ref<number | null>(null);
  const keyword = ref('');
  const projects = ref<Array<{ id: number; name: string }>>([]);
  const taskDetailRef = ref();

  const projectOptions = computed(() =>
    projects.value.map((p) => ({ label: p.name, value: p.id }))
  );

  // ---- 我的效率统计卡 ----
  const stats = ref<MyTaskStats | null>(null);
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

  function fmtLead(h?: number): string {
    if (!h) return '-';
    return h >= 48 ? `${(h / 24).toFixed(1)} 天` : `${Math.round(h)} h`;
  }

  function initTrend() {
    if (!trendRef.value || !stats.value) return;
    disposeChart();
    chart = echarts.init(trendRef.value);
    chart.setOption({
      tooltip: { trigger: 'axis' },
      grid: { left: '2%', right: '3%', top: 24, bottom: '3%', containLabel: true },
      xAxis: { type: 'category', data: stats.value.trend.map((p) => p.day.slice(5)) },
      yAxis: { type: 'value', minInterval: 1 },
      series: [{ name: '完成', type: 'line', smooth: true, data: stats.value.trend.map((p) => p.done) }],
    });
    window.addEventListener('resize', handleChartResize);
  }

  async function loadStats() {
    try {
      const res = await getMyTaskStats();
      stats.value = res || null;
      await nextTick();
      initTrend();
    } catch {
      // http 层统一提示；统计失败不阻塞任务列表
    }
  }

  const columns: DataTableColumns<MyTaskItem> = [
    {
      title: '项目',
      key: 'projectName',
      width: 160,
      ellipsis: { tooltip: true },
      render: (row) =>
        h(
          NButton,
          { text: true, type: 'primary', size: 'small', onClick: () => gotoProject(row) },
          () => row.projectName || `#${row.projectId}`
        ),
    },
    {
      title: '标题',
      key: 'title',
      ellipsis: { tooltip: true },
      render: (row) =>
        h(
          NButton,
          { text: true, size: 'small', onClick: () => taskDetailRef.value?.openModal(row.id) },
          () => row.title
        ),
    },
    {
      title: '类型',
      key: 'type',
      width: 90,
      render: (row) =>
        h(NTag, { size: 'small', bordered: false, type: typeTagType(row.type) }, () => typeLabel(row.type)),
    },
    {
      title: '优先级',
      key: 'priority',
      width: 80,
      render: (row) => h(NTag, { size: 'small', type: priorityTagType(row.priority) }, () => `P${row.priority}`),
    },
    {
      title: '标签',
      key: 'tags',
      width: 160,
      render: (row) =>
        h(
          NSpace,
          { size: 4 },
          () => (row.tags || []).map((t) => h(NTag, { size: 'small', round: true, bordered: false }, () => t))
        ),
    },
    {
      title: '状态',
      key: 'status',
      width: 110,
      render: (row) =>
        h(NTag, { size: 'small', type: statusTagType(row.status) }, () => statusLabel(row.status)),
    },
    {
      title: '截止',
      key: 'dueDate',
      width: 150,
      render: (row) =>
        row.dueDate
          ? h(NTag, { size: 'small', type: dueTagType(row.dueDate, row.status) || 'default' }, () => dueLabel(row.dueDate))
          : h('span', {}, () => '-'),
    },
    { title: '更新时间', key: 'updatedAt', width: 170 },
    {
      title: '操作',
      key: 'actions',
      width: 90,
      render: (row) =>
        h(
          NButton,
          { size: 'tiny', onClick: () => taskDetailRef.value?.openModal(row.id) },
          () => '详情'
        ),
    },
  ];

  // 过滤条件变化回到第一页再查
  function onFilterChange() {
    pagination.page = 1;
    load();
  }

  async function load() {
    loading.value = true;
    try {
      const res = await getMyTasks({
        ...(statusFilter.value ? { status: statusFilter.value } : {}),
        ...(projectId.value ? { projectId: projectId.value } : {}),
        ...(keyword.value ? { keyword: keyword.value } : {}),
        page: pagination.page,
        size: pagination.size,
      });
      list.value = res?.list || [];
      total.value = res?.total || 0;
    } catch {
      // http 层统一提示
    } finally {
      loading.value = false;
    }
  }

  async function loadProjects() {
    try {
      const res = await getProjects({ size: 100 });
      projects.value = (res?.list || []).map((p: any) => ({ id: p.id, name: p.name }));
    } catch {
      // ignore
    }
  }

  function gotoProject(row: MyTaskItem) {
    router.push(`/project/${row.projectId}/tasks`);
  }

  onMounted(() => {
    load();
    loadProjects();
    loadStats();
  });

  onUnmounted(disposeChart);
</script>

<style lang="less" scoped>
  .header-hint {
    margin-left: 8px;
    font-size: 12px;
    color: #999;
    font-weight: normal;
  }

  .stats-layout {
    display: flex;
    gap: 24px;
    align-items: stretch;
  }

  .stats-nums {
    flex: 1;
    min-width: 280px;
    display: flex;
    flex-direction: column;
    gap: 12px;
    justify-content: center;
  }

  .stat-row {
    display: flex;
    gap: 40px;
    flex-wrap: wrap;
  }

  .stat-lead {
    font-size: 24px;
  }

  .stat-status {
    flex-wrap: wrap;
  }

  .stat-proj-label {
    font-size: 12px;
    color: #999;
  }

  .stats-chart {
    width: 46%;
    min-width: 320px;
    height: 180px;
  }
</style>
