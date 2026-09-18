<template>
  <div>
    <n-card :bordered="false" title="执行记录" class="proCard">
      <template #header-extra>
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
      </template>

      <n-spin :show="loading">
        <EmptyState
          type="doc"
          title="暂无执行记录"
          description="通过 bcode-pytest 插件（pytest --bcode）或 CI 上报到此处"
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
    </n-card>

    <!-- 执行详情：汇总 + 逐用例结果 -->
    <n-modal
      v-model:show="showDetailModal"
      preset="card"
      :title="`执行记录 #${currentRun?.id || ''}`"
      style="width: 960px; max-height: 85vh"
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
                <th style="width: 40%">用例</th>
                <th>映射平台用例</th>
                <th>状态</th>
                <th>耗时</th>
                <th>失败信息</th>
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
                    <n-tag :type="caseTagType(c.status)" size="small">{{ caseLabel(c.status) }}</n-tag>
                  </td>
                  <td>{{ fmtDuration(c.durationMs) }}</td>
                  <td>
                    <n-button v-if="c.message" text type="info" size="small" @click="toggleExpand(c.id)">
                      {{ expanded.has(c.id) ? '收起' : '查看' }}
                    </n-button>
                    <span v-else class="text-gray-400">-</span>
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
  </div>
</template>

<script lang="ts" setup>
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { ref, reactive, computed, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { useMessage, useDialog } from 'naive-ui';
  import { getTestRuns, getTestRunDetail, deleteTestRun } from '@/api/test/index';
  import type { TestRunItem, TestRunDetail } from '@/api/test/index';

  const message = useMessage();
  const dialog = useDialog();
  const route = useRoute();
  const projectId = computed(() => Number(route.params.projectId));

  const loading = ref(false);
  const runList = ref<TestRunItem[]>([]);
  const total = ref(0);
  const pagination = reactive({ page: 1, pageSize: 10 });
  const filters = reactive({ source: null as string | null, status: null as string | null, branch: '' });

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
    detailLoading.value = true;
    runDetail.value = null;
    expanded.value = new Set();
    caseFilter.value = item.failed + item.errors > 0 ? 'bad' : 'all';
    try {
      runDetail.value = await getTestRunDetail(item.id);
    } catch (e) {
      message.error('加载执行详情失败');
    } finally {
      detailLoading.value = false;
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
        } catch (e) {
          message.error('删除失败');
        }
      },
    });
  }

  onMounted(loadData);
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
</style>
