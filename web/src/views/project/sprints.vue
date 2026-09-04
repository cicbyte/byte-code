<template>
  <div>
    <n-card :bordered="false" title="Sprint 管理" class="proCard">
      <template #header-extra>
        <n-button type="primary" @click="showCreate = true">新建 Sprint</n-button>
      </template>

      <n-spin :show="loading">
        <n-empty v-if="!loading && sprints.length === 0" description="暂无 Sprint" />
        <n-table v-else :bordered="false" :single-line="false" size="small">
            <thead>
            <tr>
              <th>名称</th>
              <th>目标</th>
              <th>开始日期</th>
              <th>结束日期</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="sprint in sprints" :key="sprint.id">
              <td>{{ sprint.name }}</td>
              <td>{{ sprint.goal }}</td>
              <td>{{ sprint.startDate }}</td>
              <td>{{ sprint.endDate }}</td>
              <td>
                <n-tag :type="statusMap[sprint.status]?.type || 'default'" size="small">
                  {{ statusMap[sprint.status]?.label || sprint.status }}
                </n-tag>
              </td>
              <td>
                <n-button text type="primary" size="small" @click="openBurndown(sprint)">燃尽图</n-button>
              </td>
            </tr>
          </tbody>
        </n-table>
      </n-spin>
    </n-card>

    <n-modal v-model:show="showCreate" title="新建 Sprint" preset="card" style="width: 500px">
      <n-form ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="80">
        <n-form-item label="名称" path="name">
          <n-input v-model:value="form.name" placeholder="Sprint 1" />
        </n-form-item>
        <n-form-item label="目标" path="goal">
          <n-input v-model:value="form.goal" type="textarea" placeholder="Sprint 目标" />
        </n-form-item>
        <n-form-item label="开始日期" path="startDate">
          <n-date-picker v-model:formatted-value="form.startDate" type="date" placeholder="选择开始日期" style="width: 100%" />
        </n-form-item>
        <n-form-item label="结束日期" path="endDate">
          <n-date-picker v-model:formatted-value="form.endDate" type="date" placeholder="选择结束日期" style="width: 100%" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showCreate = false">取消</n-button>
          <n-button type="primary" @click="handleCreate" :loading="submitting">创建</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 燃尽图抽屉 -->
    <n-drawer v-model:show="showBurndown" :width="640" placement="right">
      <n-drawer-content :title="`燃尽图 - ${currentSprint?.name || ''}`" closable>
        <n-spin :show="burndownLoading">
          <n-empty v-if="!burndownLoading && burndownItems.length === 0" description="暂无燃尽图数据（Sprint 内没有任务或日期无效）" />
          <div v-show="!burndownLoading && burndownItems.length > 0" ref="chartRef" style="width: 100%; height: 420px"></div>
        </n-spin>
      </n-drawer-content>
    </n-drawer>
  </div>
</template>

<script setup lang="ts">
  import { SPRINT_STATUS } from '@/enums/entities';
  import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue';
  import { useRoute } from 'vue-router';
  import { useMessage } from 'naive-ui';
  import echarts from '@/utils/lib/echarts';
  import { getSprints, createSprint, getSprintBurndown } from '@/api/project';
  import type { SprintItem, BurndownItem } from '@/api/project';

  const route = useRoute();
  const message = useMessage();
  const projectId = computed(() => Number(route.params.projectId));

  const loading = ref(false);
  const submitting = ref(false);
  const sprints = ref<SprintItem[]>([]);
  const showCreate = ref(false);
  const formRef = ref();

  // 燃尽图
  const showBurndown = ref(false);
  const burndownLoading = ref(false);
  const burndownItems = ref<BurndownItem[]>([]);
  const currentSprint = ref<SprintItem | null>(null);
  const chartRef = ref<HTMLElement | null>(null);
  let chart: echarts.ECharts | null = null;

  function disposeChart() {
    window.removeEventListener('resize', handleChartResize);
    chart?.dispose();
    chart = null;
  }

  function handleChartResize() {
    chart?.resize();
  }

  async function openBurndown(sprint: SprintItem) {
    currentSprint.value = sprint;
    showBurndown.value = true;
    burndownLoading.value = true;
    burndownItems.value = [];
    try {
      const res = await getSprintBurndown(sprint.id);
      burndownItems.value = res?.items || [];
    } catch (e: any) {
      message.error(e.message || '获取燃尽图失败');
      return;
    } finally {
      burndownLoading.value = false;
    }
    if (burndownItems.value.length === 0) return;
    await nextTick();
    // 等抽屉进场动画结束再初始化，否则容器宽度被测成动画中的中间值
    setTimeout(renderChart, 350);
  }

  function renderChart() {
    if (!chartRef.value) return;
    disposeChart();
    chart = echarts.init(chartRef.value);

    const items = burndownItems.value;
    const dates = items.map((i) => i.date);
    const total = items.length > 0 ? items[0].remaining : 0;
    // 理想线：从总量线性递减到 0
    const n = items.length;
    const ideal = items.map((_, idx) => (n > 1 ? Math.round((total * (n - 1 - idx)) / (n - 1)) : 0));

    chart.setOption({
      tooltip: { trigger: 'axis' },
      legend: { data: ['剩余任务', '理想线'] },
      grid: { left: 48, right: 24, top: 48, bottom: 32 },
      xAxis: { type: 'category', data: dates, boundaryGap: false },
      yAxis: { type: 'value', name: '任务数', minInterval: 1 },
      series: [
        {
          name: '剩余任务',
          type: 'line',
          data: items.map((i) => i.remaining),
          smooth: true,
          symbolSize: 6,
          itemStyle: { color: '#2080f0' },
          areaStyle: { color: 'rgba(32,128,240,0.12)' },
        },
        {
          name: '理想线',
          type: 'line',
          data: ideal,
          smooth: true,
          symbol: 'none',
          lineStyle: { type: 'dashed', color: '#f0a020' },
          itemStyle: { color: '#f0a020' },
        },
      ],
    });
    window.addEventListener('resize', handleChartResize);
  }

  onUnmounted(disposeChart);

  const form = ref({ name: '', goal: '', startDate: '', endDate: '' });
  const rules = {
    name: { required: true, message: '请输入名称' },
  };

  // 字典统一出口：enums/entities.ts（与 test/plans.vue 共用同一套 Sprint 状态）
  const statusMap = SPRINT_STATUS;

  async function loadSprints() {
    loading.value = true;
    try {
      const res = await getSprints(projectId.value);
      sprints.value = res?.list || [];
    } catch {
      // ignore
    } finally {
      loading.value = false;
    }
  }

  async function handleCreate() {
    submitting.value = true;
    try {
      await createSprint(projectId.value, { ...form.value });
      message.success('创建成功');
      showCreate.value = false;
      form.value = { name: '', goal: '', startDate: '', endDate: '' };
      loadSprints();
    } catch (e: any) {
      message.error(e.message || '创建失败');
    } finally {
      submitting.value = false;
    }
  }

  onMounted(loadSprints);
</script>
