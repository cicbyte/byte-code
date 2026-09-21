<template>
  <div>
    <n-card :bordered="false" class="proCard">
      <template #header>
        <n-space size="small" align="center">
          <n-button text type="info" @click="backToList" data-test-id="runs.detail.back">
            <template #icon>
              <n-icon><ArrowLeftOutlined /></n-icon>
            </template>
            执行记录
          </n-button>
          <span class="text-gray-400">/</span>
          <span data-test-id="runs.detail.title">执行记录 #{{ runId }}</span>
        </n-space>
      </template>

      <n-spin :show="detailLoading">
        <template v-if="runDetail">
          <!-- 该次执行整体情况：汇总统计 -->
          <n-space class="mb-3" :size="24" align="center" data-test-id="runs.detail.stats">
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
            <n-statistic label="通过率" :value="passRateText" />
          </n-space>
          <n-space class="mb-3" size="small" align="center">
            <n-tag size="small" :type="TEST_RUN_SOURCE.tagType(runDetail.source)">
              {{ TEST_RUN_SOURCE.label(runDetail.source) }}
            </n-tag>
            <span v-if="runDetail.branch" class="text-gray-500 text-xs">{{ runDetail.branch }}</span>
            <span v-if="runDetail.gitSha" class="text-gray-400 text-xs">{{ runDetail.gitSha.slice(0, 8) }}</span>
            <span v-if="runDetail.env" class="text-gray-400 text-xs">{{ runDetail.env }}</span>
            <span v-if="runDetail.startedAt" class="text-gray-400 text-xs">
              {{ runDetail.startedAt }} ~ {{ runDetail.finishedAt }}
            </span>
            <span v-if="runDetail.triggeredByName" class="text-gray-400 text-xs">
              触发：{{ runDetail.triggeredByName }}
            </span>
            <n-select
              v-model:value="caseFilter"
              :options="caseFilterOptions"
              size="small"
              style="width: 110px"
            />
          </n-space>

          <EmptyState
            v-if="filteredCases.length === 0 && !detailLoading"
            type="search"
            title="无匹配用例"
            :description="caseFilter === 'bad' ? '该次执行没有失败/错误用例' : '该次执行没有用例记录'"
          />
          <n-table v-else :bordered="false" :single-line="false" size="small">
            <thead>
              <tr>
                <th class="num-col">#</th>
                <th style="width: 32%">用例</th>
                <th>映射平台用例</th>
                <th>状态</th>
                <th>耗时</th>
                <th>失败信息 / 缺陷</th>
                <th>附件</th>
              </tr>
            </thead>
            <tbody>
              <template v-for="(c, idx) in filteredCases" :key="c.id">
                <tr :data-test-id="`runs.detail.case-${c.id}`">
                  <td class="num-col">{{ idx + 1 }}</td>
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
                    <n-space size="small" :wrap="false">
                      <n-tag :type="RUN_CASE_STATUS.tagType(c.status)" size="small">
                        {{ RUN_CASE_STATUS.label(c.status) }}
                      </n-tag>
                      <n-tag v-if="c.flaky" type="warning" size="small" :bordered="false">flaky</n-tag>
                    </n-space>
                  </td>
                  <td>{{ fmtDuration(c.durationMs) }}</td>
                  <td>
                    <n-space size="small" align="center" :wrap="false">
                      <n-button v-if="c.message || c.code" text type="info" size="small" @click="toggleExpand(c.id)">
                        {{ expanded.has(c.id) ? '收起' : '查看' }}
                      </n-button>
                      <router-link
                        v-if="c.bugTaskId"
                        :to="{ name: 'project_tasks' }"
                        class="text-xs"
                        style="text-decoration: none"
                      >
                        <n-tag type="error" size="small" :bordered="false">缺陷 #{{ c.bugTaskId }}</n-tag>
                      </router-link>
                      <n-button
                        v-else-if="c.status === 'fail' || c.status === 'error'"
                        text
                        type="error"
                        size="small"
                        @click="openBugDialog(c)"
                        :data-test-id="`runs.bug-btn-${c.id}`"
                      >
                        转缺陷
                      </n-button>
                      <span v-if="!c.message && !c.bugTaskId && c.status !== 'fail' && c.status !== 'error'" class="text-gray-400">-</span>
                    </n-space>
                  </td>
                  <td>
                    <n-space size="small" :wrap="false">
                      <n-tag
                        v-for="a in caseAttachments(c.id)"
                        :key="a.id"
                        size="small"
                        :bordered="false"
                        class="cursor-pointer"
                        :title="a.originalName"
                        @click="openAttachment(a.id)"
                      >
                        {{ attachLabel(a.originalName) }}
                      </n-tag>
                      <span v-if="caseAttachments(c.id).length === 0" class="text-gray-400">-</span>
                    </n-space>
                  </td>
                </tr>
                <tr v-if="expanded.has(c.id)">
                  <td colspan="6">
                    <pre class="fail-msg">{{ c.message }}</pre>
                    <template v-if="c.code">
                      <div class="text-xs text-gray-400 mb-1">源码快照</div>
                      <pre class="code-snap">{{ c.code }}</pre>
                    </template>
                  </td>
                </tr>
              </template>
            </tbody>
          </n-table>
        </template>
        <EmptyState
          v-else-if="!detailLoading"
          type="search"
          title="执行记录不存在"
          description="可能已被删除，返回列表看看"
        />
      </n-spin>
    </n-card>

    <!-- 转缺陷弹窗：建 bug 任务并挂接（未指派任务会广播给绑定 agent 认领） -->
    <n-modal
      v-model:show="showBugModal"
      preset="dialog"
      title="失败用例转缺陷任务"
      positive-text="创建并挂接"
      negative-text="取消"
      @positive-click="submitBug"
      style="width: 560px"
    >
      <n-form :model="bugForm" label-placement="left" :label-width="72" class="py-4">
        <n-form-item label="任务标题">
          <n-input v-model:value="bugForm.title" placeholder="缺省按用例标题生成" />
        </n-form-item>
        <n-form-item label="优先级">
          <n-select v-model:value="bugForm.priority" :options="prioOptions" size="small" />
        </n-form-item>
        <n-alert type="info" :bordered="false">
          任务描述将自动附带失败信息与执行上下文（分支/commit/run 入口）；未指派任务会广播给项目内已接入的
          agent 认领，修复后 bcode test --run 重跑即形成验证闭环。
        </n-alert>
      </n-form>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { RUN_CASE_STATUS, TEST_RUN_SOURCE } from '@/enums/test';
  import { ref, reactive, computed, watch } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { useMessage } from 'naive-ui';
  import { ArrowLeftOutlined } from '@vicons/antd';
  import { getTestRunDetail, caseToBug } from '@/api/test/index';
  import { getAttachments, downloadAttachment } from '@/api/attachment/index';
  import type { AttachmentItem } from '@/api/attachment/index';
  import type { TestRunDetail } from '@/api/test/index';

  const message = useMessage();
  const route = useRoute();
  const router = useRouter();
  const projectId = computed(() => Number(route.params.projectId));
  const runId = computed(() => Number(route.params.runId));

  const prioOptions = [
    { label: 'P1 紧急', value: 1 },
    { label: 'P2 高', value: 2 },
    { label: 'P3 中', value: 3 },
    { label: 'P4 低', value: 4 },
  ];

  function fmtDuration(ms: number): string {
    if (!ms || ms <= 0) return '-';
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
    const m = Math.floor(ms / 60000);
    const s = Math.round((ms % 60000) / 1000);
    return `${m}m${String(s).padStart(2, '0')}s`;
  }

  // ---------- 详情（URL 可深链：/project/:pid/test-runs/:runId） ----------
  const detailLoading = ref(false);
  const runDetail = ref<TestRunDetail | null>(null);
  const expanded = ref(new Set<number>());
  const caseFilter = ref('all');

  const caseFilterOptions = [
    { label: '全部用例', value: 'all' },
    { label: '仅失败/错误', value: 'bad' },
    { label: '仅通过', value: 'pass' },
  ];

  const passRateText = computed(() => {
    const d = runDetail.value;
    if (!d || d.total <= 0) return '-';
    return `${Math.round((d.passed / d.total) * 100)}%`;
  });

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

  function backToList() {
    router.push(`/project/${projectId.value}/test-runs`);
  }

  // 执行用例附件（失败截图等，entityType=test_run_case）：详情加载时对失败/错误行并行拉取
  const attachmentsMap = ref<Record<number, AttachmentItem[]>>({});
  function caseAttachments(caseId: number): AttachmentItem[] {
    return attachmentsMap.value[caseId] || [];
  }
  async function openAttachment(id: number) {
    try {
      const res = await downloadAttachment(id);
      if (res?.url) window.open(res.url, '_blank');
    } catch (e) {
      message.error('获取附件失败');
    }
  }
  function attachLabel(filename: string): string {
    if (/\.(png|jpe?g|gif|webp)$/i.test(filename)) return '截图';
    if (/\.log$/i.test(filename)) return '日志';
    return filename.length > 12 ? filename.slice(0, 10) + '…' : filename;
  }
  async function loadCaseAttachments(cases: { id: number; status: string }[]) {
    const targets = cases.filter((c) => c.status === 'fail' || c.status === 'error');
    const results = await Promise.all(
      targets.map((c) =>
        getAttachments('test_run_case', c.id)
          .then((r) => ({ id: c.id, list: r?.list || [] }))
          .catch(() => ({ id: c.id, list: [] }))
      )
    );
    const map: Record<number, AttachmentItem[]> = {};
    for (const r of results) {
      if (r.list.length) map[r.id] = r.list;
    }
    attachmentsMap.value = map;
  }

  async function loadDetail() {
    detailLoading.value = true;
    runDetail.value = null;
    expanded.value = new Set();
    try {
      const det = await getTestRunDetail(runId.value);
      runDetail.value = det;
      // 有失败默认聚焦失败（总览视角先看坏的）
      caseFilter.value = det && det.failed + det.errors > 0 ? 'bad' : 'all';
      loadCaseAttachments(det ? det.cases : []);
    } catch (e) {
      message.error('加载执行详情失败');
    } finally {
      detailLoading.value = false;
    }
  }

  watch(
    runId,
    () => {
      if (runId.value) loadDetail();
    },
    { immediate: true }
  );

  // ---------- 转缺陷 ----------
  const showBugModal = ref(false);
  const bugSubmitting = ref(false);
  const bugForm = reactive({ runCaseId: 0, title: '', priority: 2 });

  function openBugDialog(c: { id: number; title: string }) {
    bugForm.runCaseId = c.id;
    bugForm.title = '';
    bugForm.priority = 2;
    showBugModal.value = true;
  }

  async function submitBug() {
    if (bugSubmitting.value) return false;
    bugSubmitting.value = true;
    try {
      const res = await caseToBug(bugForm.runCaseId, {
        title: bugForm.title || undefined,
        priority: bugForm.priority,
      });
      message.success(`已创建缺陷任务 #${res.taskId}（已广播绑定 agent）`);
      // 详情刷新拿挂钩回显
      if (runDetail.value) {
        await loadDetail();
      }
      return true;
    } catch (e: any) {
      message.error(e?.message || '转缺陷失败');
      return false;
    } finally {
      bugSubmitting.value = false;
    }
  }
</script>

<style scoped>
  /* 用例序号列：窄列灰字，不抢内容视觉 */
  .num-col {
    width: 36px;
    color: var(--text-3, #8b949e);
    font-size: 12px;
    text-align: center;
    white-space: nowrap;
  }
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
  .code-snap {
    margin: 8px 0 0;
    padding: 10px 14px;
    max-height: 320px;
    overflow: auto;
    background: rgba(24, 160, 88, 0.05);
    border-left: 3px solid rgba(24, 160, 88, 0.4);
    border-radius: 4px;
    font-size: 12px;
    line-height: 1.7;
    font-family: 'JetBrains Mono', Consolas, monospace;
    white-space: pre;
  }
</style>
