<template>
  <div>
    <n-card :bordered="false" title="待审核" class="proCard">
      <template #header-extra>
        <n-space :size="8" align="center">
          <n-tag v-if="pending.length" type="warning" size="small">{{ pending.length }} 条待审</n-tag>
          <span v-if="checkedIds.length" class="text-xs text-gray-400">已选 {{ checkedIds.length }} 条</span>
          <n-button
            size="small"
            type="success"
            :disabled="checkedIds.length === 0"
            :loading="batching"
            @click="batchApprove"
            data-test-id="project-reviews.batch-approve-btn"
          >
            批量通过{{ checkedIds.length ? ` (${checkedIds.length})` : '' }}
          </n-button>
          <n-button
            size="small"
            type="warning"
            ghost
            :disabled="checkedIds.length === 0"
            @click="openBatchReject"
            data-test-id="project-reviews.batch-reject-btn"
          >
            批量驳回
          </n-button>
        </n-space>
      </template>

      <ProTable
        ref="tableRef"
        table-key="project-reviews"
        :columns="columns"
        :request="tableRequest"
        :row-key="(row: TaskItem) => row.id"
        :row-props="rowProps"
        :pagination="false"
        :striped="true"
      >
        <template #toolbar>
          <input
            type="checkbox"
            :checked="allChecked"
            :indeterminate.prop="someChecked"
            data-test-id="project-reviews.check-all"
            @change="toggleAll"
          />
          <span class="text-xs text-gray-400 ml-1">全选本页</span>
        </template>
        <template #empty>
          <EmptyState
            type="task"
            title="没有待审核的任务"
            description="Agent 或成员完成任务后会进入这里等待验收"
          />
        </template>
      </ProTable>
    </n-card>

    <!-- 详情抽屉：描述 + Agent 产出，审核动作常驻 footer -->
    <n-drawer v-model:show="showDetail" :width="720" placement="right" :auto-focus="false">
      <n-drawer-content :title="detail?.title || '任务详情'" closable>
        <div v-if="detail" class="rv-detail">
          <n-space size="small" align="center" class="rv-meta">
            <n-tag :type="priorityTagType(detail.priority)" size="small">P{{ detail.priority }}</n-tag>
            <span class="rv-meta-text">
              提交人 {{ detail.assigneeName || '-' }}
              <n-tag v-if="detail.source === 'agent'" size="tiny" :bordered="false" type="info">Agent</n-tag>
              · 完成于 {{ (detail.updatedAt || '').slice(0, 16) }}
            </span>
          </n-space>
          <div class="rv-sec">
            <div class="rv-label">任务描述</div>
            <div class="rv-text">{{ detail.description || '（无描述）' }}</div>
          </div>
          <div class="rv-sec">
            <div class="rv-label">Agent 产出</div>
            <MdPreview
              v-if="detail.artifacts"
              :id="`md-review-${detail.id}`"
              :model-value="safeHtml(detail.artifacts)"
            />
            <div v-else class="rv-empty">（无产出内容）</div>
          </div>
        </div>
        <template #footer>
          <n-space :size="8" v-if="detail">
            <n-button size="small" type="error" ghost @click="openReject(detail)">驳回</n-button>
            <n-popconfirm @positive-click="approve(detail)">
              <template #trigger>
                <n-button size="small" type="success">通过</n-button>
              </template>
              通过后任务转入已完成（done）。
            </n-popconfirm>
          </n-space>
        </template>
      </n-drawer-content>
    </n-drawer>

    <!-- 驳回理由：作为评论通知提交人 -->
    <n-modal v-model:show="showReject" preset="dialog" title="驳回理由" :show-icon="false">
      <n-input
        v-model:value="rejectComment"
        type="textarea"
        placeholder="驳回原因（必填，将作为评论通知提交人）"
        :rows="3"
      />
      <template #action>
        <n-space>
          <n-button size="small" @click="showReject = false">取消</n-button>
          <n-button size="small" type="error" :disabled="!rejectComment.trim()" @click="confirmReject">确认驳回</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 批量驳回：统一理由逐条落评论（与审核中心同口径） -->
    <n-modal v-model:show="showBatchReject" preset="dialog" title="批量驳回" :show-icon="false" style="width: 460px">
      <n-space vertical :size="8" class="py-2">
        <span class="text-sm">已选 {{ checkedIds.length }} 条将全部回到进行中，统一理由逐条落为任务评论并通知执行者</span>
        <n-input v-model:value="batchRejectComment" type="textarea" placeholder="驳回理由（必填）" :rows="3" />
      </n-space>
      <template #action>
        <n-space>
          <n-button size="small" @click="showBatchReject = false">取消</n-button>
          <n-button size="small" type="error" :disabled="!batchRejectComment.trim()" :loading="batching" @click="confirmBatchReject">确认驳回</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, h, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { useMessage } from 'naive-ui';
  import DOMPurify from 'dompurify';
  import { MdPreview } from 'md-editor-v3';
  import 'md-editor-v3/lib/preview.css';
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { getTasks, reviewTask } from '@/api/project/index';
  import { batchReview } from '@/api/platform/index';
  import type { TaskItem } from '@/api/project/index';
  import { priorityTagType } from '@/enums/task';
  import ProTable from '@/components/ProTable/index.vue';
  import { NButton, NPopconfirm, NSpace, NTag } from 'naive-ui';
  import type { DataTableColumns } from 'naive-ui';

  const route = useRoute();
  const message = useMessage();
  const projectId = computed(() => Number(route.params.projectId));

  const tableRef = ref();
  // 待审 = 完成进 review 且 requiresHumanReview=1 / humanReviewStatus=pending
  // （CompleteTask 统一置位；审完 hrs 变化自然出队）
  // 请求不分页（队列一次拉全量），pending 由原始列表过滤派生
  const list = ref<TaskItem[]>([]);

  async function tableRequest() {
    const res = await getTasks(projectId.value, { status: 'review', size: 200 });
    list.value = res?.list || [];
    return { list: list.value, itemCount: list.value.length };
  }

  const pending = computed(() =>
    list.value.filter((t) => t.requiresHumanReview === 1 && t.humanReviewStatus === 'pending')
  );

  function rowProps(row: TaskItem) {
    return { 'data-test-id': `reviews.item-${row.id}` };
  }

  const columns = computed<DataTableColumns<TaskItem>>(() => [
    {
      title: '',
      key: 'check',
      width: 40,
      render: (t) =>
        h('input', {
          type: 'checkbox',
          value: t.id,
          checked: checkedIds.value.includes(t.id),
          'data-test-id': `project-reviews.check-${t.id}`,
          onChange: () => toggleCheck(t.id),
        }),
    },
    {
      title: '标题',
      key: 'title',
      render: (t) =>
        h(
          NButton,
          { text: true, type: 'info', onClick: () => openDetail(t), 'data-test-id': `reviews.title-btn-${t.id}` },
          () => t.title
        ),
    },
    {
      title: '提交人',
      key: 'assigneeName',
      width: 150,
      render: (t) =>
        h(NSpace, { size: 4, align: 'center' }, () => [
          t.assigneeName || '-',
          t.source === 'agent' ? h(NTag, { size: 'tiny', bordered: false, type: 'info' }, () => 'Agent') : null,
        ]),
    },
    {
      title: '优先级',
      key: 'priority',
      width: 80,
      render: (t) => h(NTag, { type: priorityTagType(t.priority), size: 'small' }, () => `P${t.priority}`),
    },
    { title: '完成时间', key: 'updatedAt', width: 150, render: (t) => (t.updatedAt || '').slice(0, 16) },
    {
      title: '操作',
      key: 'actions',
      width: 150,
      render: (t) =>
        h(NSpace, { size: 8 }, () => [
          h(
            NPopconfirm,
            { onPositiveClick: () => approve(t) },
            {
              trigger: () =>
                h(NButton, { size: 'tiny', type: 'success', 'data-test-id': `reviews.approve-btn-${t.id}` }, () => '通过'),
              default: () => '通过后任务转入已完成（done）。',
            }
          ),
          h(NButton, { size: 'tiny', type: 'error', ghost: true, onClick: () => openReject(t) }, () => '驳回'),
        ]),
    },
  ]);

  const showReject = ref(false);
  const rejectComment = ref('');
  const rejecting = ref<TaskItem | null>(null);

  // ---- 详情抽屉 ----
  const showDetail = ref(false);
  const detail = ref<TaskItem | null>(null);

  function openDetail(t: TaskItem) {
    detail.value = t;
    showDetail.value = true;
  }

  // artifacts 来自 agent 产出，渲染前消毒（与任务详情抽屉同口径）
  function safeHtml(md: string): string {
    return DOMPurify.sanitize(md);
  }

  function load() {
    tableRef.value?.reload();
  }

  function drop(id: number) {
    list.value = list.value.filter((t) => t.id !== id);
    // 打开中的详情被审掉（含批量）时同步收起抽屉
    if (detail.value?.id === id) {
      detail.value = null;
      showDetail.value = false;
    }
  }

  async function approve(t: TaskItem) {
    try {
      await reviewTask(t.id, { status: 'approved' });
      message.success(`「${t.title}」已通过`);
      drop(t.id);
    } catch {
      message.error('审核操作失败');
    }
  }

  function openReject(t: TaskItem) {
    rejecting.value = t;
    rejectComment.value = '';
    showReject.value = true;
  }

  async function confirmReject() {
    const t = rejecting.value;
    if (!t || !rejectComment.value.trim()) return;
    try {
      await reviewTask(t.id, { status: 'rejected', comment: rejectComment.value.trim() });
      message.success(`「${t.title}」已驳回（回到进行中）`);
      showReject.value = false;
      drop(t.id);
    } catch {
      message.error('驳回失败');
    }
  }

  // ---- 多选批量（复用审核中心 batchReview 端点） ----
  const checkedIds = ref<number[]>([]);
  const batching = ref(false);
  const showBatchReject = ref(false);
  const batchRejectComment = ref('');

  const allChecked = computed(() => pending.value.length > 0 && checkedIds.value.length === pending.value.length);
  const someChecked = computed(() => checkedIds.value.length > 0 && !allChecked.value);

  function toggleCheck(id: number) {
    checkedIds.value = checkedIds.value.includes(id)
      ? checkedIds.value.filter((x) => x !== id)
      : [...checkedIds.value, id];
  }
  function toggleAll() {
    checkedIds.value = allChecked.value ? [] : pending.value.map((t) => t.id);
  }

  function reportFailed(failed: Array<{ id: number; error: string }>) {
    if (failed?.length) {
      message.warning(`${failed.length} 条失败：${failed.map((f) => `#${f.id} ${f.error}`).join('；')}`);
    }
  }

  async function batchApprove() {
    if (!checkedIds.value.length) return;
    batching.value = true;
    try {
      const res = await batchReview({ ids: [...checkedIds.value], status: 'approved' });
      if (res?.succeeded) message.success(`已通过 ${res.succeeded} 条`);
      reportFailed(res?.failed || []);
      checkedIds.value.forEach(drop);
      checkedIds.value = [];
      load();
    } catch {
      message.error('批量通过失败');
    } finally {
      batching.value = false;
    }
  }

  function openBatchReject() {
    batchRejectComment.value = '';
    showBatchReject.value = true;
  }

  async function confirmBatchReject() {
    if (!checkedIds.value.length || !batchRejectComment.value.trim()) return;
    batching.value = true;
    try {
      const res = await batchReview({
        ids: [...checkedIds.value],
        status: 'rejected',
        comment: batchRejectComment.value.trim(),
      });
      if (res?.succeeded) message.success(`已驳回 ${res.succeeded} 条（回到进行中）`);
      reportFailed(res?.failed || []);
      showBatchReject.value = false;
      checkedIds.value.forEach(drop);
      checkedIds.value = [];
      load();
    } catch {
      message.error('批量驳回失败');
    } finally {
      batching.value = false;
    }
  }

  // ProTable 挂载即自动请求
</script>

<style lang="less" scoped>
  .col-check {
    width: 34px;
    text-align: center;

    input[type='checkbox'] {
      cursor: pointer;
      accent-color: var(--primary-color, #16a34a);
    }
  }

  .rv-meta-text {
    font-size: 12px;
    color: var(--text-color-3, #999);
  }

  .rv-sec {
    margin-top: 14px;
  }

  .rv-label {
    font-size: 12px;
    color: var(--text-color-3, #999);
    margin-bottom: 4px;
  }

  .rv-text {
    font-size: 13px;
    color: var(--text-color-2, #57606a);
    white-space: pre-wrap;
    word-break: break-word;
    line-height: 1.7;
  }

  .rv-empty {
    font-size: 12px;
    color: var(--text-color-3, #999);
  }
</style>
