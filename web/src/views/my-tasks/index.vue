<template>
  <div>
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

    <TaskDetailModal ref="taskDetailRef" />
  </div>
</template>

<script lang="ts" setup>
  import { ref, reactive, computed, h, onMounted } from 'vue';
  import { useRouter } from 'vue-router';
  import { NButton, NTag, NSpace } from 'naive-ui';
  import type { DataTableColumns } from 'naive-ui';
  import { getMyTasks, getProjects } from '@/api/project/index';
  import type { MyTaskItem } from '@/api/project/index';
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
  });
</script>

<style lang="less" scoped>
  .header-hint {
    margin-left: 8px;
    font-size: 12px;
    color: #999;
    font-weight: normal;
  }
</style>
