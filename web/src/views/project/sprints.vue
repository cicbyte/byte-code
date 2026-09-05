<template>
  <div>
    <n-card :bordered="false" title="Sprint 管理" class="proCard">
      <template #header-extra>
        <n-button type="primary" @click="openCreate">新建 Sprint</n-button>
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
                <n-tag :type="statusMap.tagType(sprint.status)" size="small">
                  {{ statusMap.label(sprint.status) }}
                </n-tag>
              </td>
              <td>
                <n-space :size="6">
                  <n-button text type="primary" size="small" @click="openBurndown(sprint)">燃尽图</n-button>
                  <n-button text type="primary" size="small" @click="openEdit(sprint)">编辑</n-button>
                  <n-button text type="info" size="small" @click="openTasks(sprint)">任务</n-button>
                  <n-button text type="error" size="small" @click="handleDelete(sprint)">删除</n-button>
                </n-space>
              </td>
            </tr>
          </tbody>
        </n-table>
      </n-spin>
    </n-card>

    <n-modal v-model:show="showCreate" :title="editingId ? '编辑 Sprint' : '新建 Sprint'" preset="card" style="width: 500px">
      <n-form ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="80">
        <n-form-item label="名称" path="name">
          <n-input v-model:value="form.name" placeholder="Sprint 1" />
        </n-form-item>
        <n-form-item label="目标" path="goal">
          <n-input v-model:value="form.goal" type="textarea" placeholder="Sprint 目标" />
        </n-form-item>
        <n-form-item label="开始日期" path="startDate">
          <n-date-picker v-model:formatted-value="form.startDate" type="date" value-format="yyyy-MM-dd" placeholder="选择开始日期" style="width: 100%" />
        </n-form-item>
        <n-form-item label="结束日期" path="endDate">
          <n-date-picker v-model:formatted-value="form.endDate" type="date" value-format="yyyy-MM-dd" placeholder="选择结束日期" style="width: 100%" />
        </n-form-item>
        <n-form-item v-if="editingId" label="状态" path="status">
          <n-select v-model:value="form.status" :options="SPRINT_STATUS.options" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showCreate = false">取消</n-button>
          <n-button type="primary" @click="handleSubmit" :loading="submitting">{{ editingId ? '保存' : '创建' }}</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- Sprint 任务绑定抽屉 -->
    <n-drawer v-model:show="showTasks" :width="480" placement="right">
      <n-drawer-content :title="`任务绑定 - ${currentSprint?.name || ''}`" closable>
        <n-spin :show="tasksLoading">
          <h4 class="mb-2">已绑定（{{ boundTasks.length }}）</h4>
          <n-empty v-if="boundTasks.length === 0" description="暂无绑定任务" size="small" />
          <n-space v-else vertical :size="4">
            <n-space v-for="t in boundTasks" :key="t.id" justify="space-between" align="center" class="w-full">
              <span class="text-sm">{{ t.title }}</span>
              <n-button text type="error" size="tiny" @click="handleUnbind(t)">移除</n-button>
            </n-space>
          </n-space>
          <n-divider />
          <h4 class="mb-2">未绑定（{{ unboundTasks.length }}）</h4>
          <n-empty v-if="unboundTasks.length === 0" description="没有可绑定的任务" size="small" />
          <n-space v-else vertical :size="4">
            <n-space v-for="t in unboundTasks" :key="t.id" justify="space-between" align="center" class="w-full">
              <n-space :size="6" align="center">
                <span class="text-sm">{{ t.title }}</span>
                <!-- 属其它 Sprint 的任务绑定即改派，显式标识防误操作 -->
                <n-tag v-if="t.sprintId" size="tiny" :bordered="false" type="warning">
                  来自：{{ sprintName(t.sprintId) }}
                </n-tag>
              </n-space>
              <n-button text type="primary" size="tiny" @click="handleBind(t)">绑定</n-button>
            </n-space>
          </n-space>
        </n-spin>
      </n-drawer-content>
    </n-drawer>

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
  import { useMessage, useDialog } from 'naive-ui';
  import echarts from '@/utils/lib/echarts';
  import {
    getSprints, createSprint, getSprintBurndown,
    updateSprint, deleteSprint, getTasks,
    addTaskToSprint, removeTaskFromSprint,
  } from '@/api/project';
  import type { SprintItem, BurndownItem, TaskItem } from '@/api/project';

  const route = useRoute();
  const message = useMessage();
  const dialog = useDialog();
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

  const form = ref({ name: '', goal: '', startDate: '', endDate: '', status: 'planning' });
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

  // ==================== 新建/编辑（弹窗复用） ====================
  const editingId = ref<number | null>(null);

  function openCreate() {
    editingId.value = null;
    form.value = { name: '', goal: '', startDate: '', endDate: '', status: 'planning' };
    showCreate.value = true;
  }

  function openEdit(s: SprintItem) {
    editingId.value = s.id;
    // DB 里的日期可能带时间部分，date-picker formatted-value 只认 yyyy-MM-dd
    const day = (v: string) => (v || '').slice(0, 10);
    form.value = {
      name: s.name,
      goal: s.goal || '',
      startDate: day(s.startDate),
      endDate: day(s.endDate),
      status: s.status || 'planning',
    };
    showCreate.value = true;
  }

  async function handleSubmit() {
    submitting.value = true;
    try {
      if (editingId.value) {
        await updateSprint(editingId.value, { ...form.value });
        message.success('已保存');
      } else {
        await createSprint(projectId.value, { ...form.value });
        message.success('创建成功');
      }
      showCreate.value = false;
      loadSprints();
    } catch (e: any) {
      message.error(e.message || (editingId.value ? '保存失败' : '创建失败'));
    } finally {
      submitting.value = false;
    }
  }

  function handleDelete(s: SprintItem) {
    dialog.warning({
      title: '确认删除 Sprint',
      content: `删除「${s.name}」后其下绑定的任务会被解绑（任务本身保留），该操作不可恢复。`,
      positiveText: '删除',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await deleteSprint(s.id);
          message.success('已删除');
          loadSprints();
        } catch (e: any) {
          message.error(e.message || '删除失败');
        }
      },
    });
  }

  // ==================== 任务绑定 ====================
  const showTasks = ref(false);
  const tasksLoading = ref(false);
  const allTasks = ref<TaskItem[]>([]);

  const boundTasks = computed(() =>
    allTasks.value.filter((t) => currentSprint.value && t.sprintId === currentSprint.value.id)
  );
  const unboundTasks = computed(() =>
    // 未绑定 = sprint_id 为空/0，或属于其它 Sprint 的都可转入（绑定即改派）
    allTasks.value.filter((t) => !currentSprint.value || t.sprintId !== currentSprint.value.id)
  );

  async function openTasks(s: SprintItem) {
    currentSprint.value = s;
    showTasks.value = true;
    tasksLoading.value = true;
    try {
      const res = await getTasks(projectId.value, { size: 200 });
      allTasks.value = res?.list || [];
    } catch {
      allTasks.value = [];
    } finally {
      tasksLoading.value = false;
    }
  }

  function sprintName(id: number): string {
    return sprints.value.find((x) => x.id === id)?.name || `#${id}`;
  }

  async function handleBind(t: TaskItem) {
    if (!currentSprint.value) return;
    const fromSprint = t.sprintId && t.sprintId !== currentSprint.value.id ? sprintName(t.sprintId) : '';
    const doBind = async () => {
      try {
        await addTaskToSprint(currentSprint.value!.id, t.id);
        t.sprintId = currentSprint.value!.id;
        message.success(`已绑定：${t.title}`);
      } catch (e: any) {
        message.error(e.message || '绑定失败');
      }
    };
    if (fromSprint) {
      dialog.warning({
        title: '确认转移任务',
        content: `「${t.title}」当前属于「${fromSprint}」，绑定到本 Sprint 会把它从原 Sprint 移出。`,
        positiveText: '转移',
        negativeText: '取消',
        onPositiveClick: doBind,
      });
      return;
    }
    await doBind();
  }

  async function handleUnbind(t: TaskItem) {
    if (!currentSprint.value) return;
    try {
      await removeTaskFromSprint(currentSprint.value.id, t.id);
      t.sprintId = 0;
      message.success(`已移除：${t.title}`);
    } catch (e: any) {
      message.error(e.message || '移除失败');
    }
  }

  onMounted(loadSprints);
</script>
