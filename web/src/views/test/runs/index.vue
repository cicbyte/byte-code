<template>
  <div>
    <n-card :bordered="false" class="proCard">
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

      <!-- 空态判断交给 ProTable 的 empty 插槽；外壳不再自旋（ProTable 内含 loading） -->
      <ProTable
        ref="tableRef"
        table-key="test-runs"
        :columns="columns"
        :request="tableRequest"
        :row-key="(row: TestRunItem) => row.id"
        :row-props="rowProps"
        :striped="true"
      >
        <template #empty>
          <EmptyState
            type="doc"
            title="暂无执行记录"
            description="通过 byte-code-pytest 插件（pytest --bcode）、bcode test --upload 或 CI 上报到此处"
          />
        </template>
      </ProTable>
    </n-card>
  </div>
</template>

<script lang="ts" setup>
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { TEST_RUN_SOURCE } from '@/enums/test';
  import { ref, reactive, computed, h, onMounted } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { useMessage, useDialog } from 'naive-ui';
  import { getTestRuns, deleteTestRun } from '@/api/test/index';
  import ProTable from '@/components/ProTable/index.vue';
  import { NButton, NSpace, NTag } from 'naive-ui';
  import type { DataTableColumns } from 'naive-ui';
  import type { TestRunItem } from '@/api/test/index';

  const message = useMessage();
  const dialog = useDialog();
  const route = useRoute();
  const router = useRouter();
  const projectId = computed(() => Number(route.params.projectId));

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

  function fmtDuration(ms: number): string {
    if (!ms || ms <= 0) return '-';
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
    const m = Math.floor(ms / 60000);
    const s = Math.round((ms % 60000) / 1000);
    return `${m}m${String(s).padStart(2, '0')}s`;
  }

  // ---------- 执行记录列表（ProTable 托管：工具栏/分页） ----------
  const tableRef = ref();
  const filters = reactive({ source: null as string | null, status: null as string | null, branch: '' });

  async function tableRequest(p: { page: number; pageSize: number }) {
    const res = await getTestRuns(projectId.value, {
      pageNum: p.page,
      pageSize: p.pageSize,
      source: filters.source || undefined,
      status: filters.status || undefined,
      branch: filters.branch || undefined,
    });
    return { list: res?.list || [], itemCount: res?.total || 0 };
  }

  function reload() {
    tableRef.value?.reload();
  }

  function rowProps(row: TestRunItem) {
    return {
      style: 'cursor: pointer;',
      'data-test-id': `runs.row-${row.id}`,
      onClick: () => goDetail(row.id),
    };
  }

  const columns: DataTableColumns<TestRunItem> = [
    {
      title: '来源',
      key: 'source',
      width: 90,
      render: (row) =>
        h(NTag, { type: TEST_RUN_SOURCE.tagType(row.source), size: 'small' }, () => TEST_RUN_SOURCE.label(row.source)),
    },
    {
      title: '分支 / Commit',
      key: 'branch',
      render: (row) =>
        row.branch || row.gitSha
          ? h('span', null, [
              row.branch || '-',
              row.gitSha ? h('span', { class: 'text-gray-400' }, ` @${row.gitSha.slice(0, 8)}`) : null,
            ])
          : h('span', null, '-'),
    },
    { title: '环境', key: 'env', width: 90, render: (row) => row.env || '-' },
    { title: '触发者', key: 'triggeredByName', width: 110, render: (row) => row.triggeredByName || '-' },
    {
      title: '结果',
      key: 'result',
      width: 230,
      render: (row) =>
        h(NSpace, { size: 'small', wrap: false }, () => [
          h(NTag, { type: 'success', size: 'small' }, () => `${row.passed} 通过`),
          row.failed ? h(NTag, { type: 'error', size: 'small' }, () => `${row.failed} 失败`) : null,
          row.errors ? h(NTag, { type: 'warning', size: 'small' }, () => `${row.errors} 错误`) : null,
          row.skipped ? h(NTag, { size: 'small' }, () => `${row.skipped} 跳过`) : null,
        ]),
    },
    { title: '耗时', key: 'durationMs', width: 80, render: (row) => fmtDuration(row.durationMs) },
    { title: '结束时间', key: 'finishedAt', width: 160, render: (row) => row.finishedAt || row.createdAt },
    {
      title: '操作',
      key: 'actions',
      width: 110,
      render: (row) =>
        h(NSpace, { size: 'small' }, () => [
          h(
            NButton,
            {
              text: true,
              type: 'info',
              onClick: (e: Event) => {
                e.stopPropagation();
                goDetail(row.id);
              },
              'data-test-id': `runs.detail-btn-${row.id}`,
            },
            () => '详情'
          ),
          h(
            NButton,
            {
              text: true,
              type: 'error',
              onClick: (e: Event) => {
                e.stopPropagation();
                handleDelete(row);
              },
            },
            () => '删除'
          ),
        ]),
    },
  ];

  function goDetail(runId: number) {
    router.push(`/project/${projectId.value}/test-runs/${runId}`);
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
        } catch (e) {
          message.error('删除失败');
        }
      },
    });
  }

  onMounted(() => {
    // ProTable 挂载即自动请求
  });
</script>

<style scoped>
  .run-row {
    cursor: pointer;
  }
  .run-row:hover {
    background: rgba(128, 128, 128, 0.06);
  }
</style>
