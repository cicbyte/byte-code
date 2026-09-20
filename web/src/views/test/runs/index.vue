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
            <tr
              v-for="(item, __ix) in runList"
              :key="item.id"
              class="run-row"
              :data-test-id="`runs.row-${item.id}`"
              @click="goDetail(item.id)"
            >
              <td class="col-idx">{{ __ix + 1 }}</td>
              <td>
                <n-tag :type="TEST_RUN_SOURCE.tagType(item.source)" size="small">
                  {{ TEST_RUN_SOURCE.label(item.source) }}
                </n-tag>
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
                  <n-button
                    text
                    type="info"
                    @click.stop="goDetail(item.id)"
                    :data-test-id="`runs.detail-btn-${item.id}`"
                  >
                    详情
                  </n-button>
                  <n-button text type="error" @click.stop="handleDelete(item)">删除</n-button>
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
    </n-card>
  </div>
</template>

<script lang="ts" setup>
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { TEST_RUN_SOURCE } from '@/enums/test';
  import { ref, reactive, computed, onMounted } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { useMessage, useDialog } from 'naive-ui';
  import { getTestRuns, deleteTestRun } from '@/api/test/index';
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

  // ---------- 执行记录列表（瘦身：总览/详情已独立成页） ----------
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
    loadData();
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
