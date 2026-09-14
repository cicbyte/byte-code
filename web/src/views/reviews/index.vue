<template>
  <div>
    <n-card :bordered="false" class="proCard">
      <template #header>
        审核中心
        <n-tag v-if="total" type="warning" size="small" class="ml-2">{{ total }} 条待审</n-tag>
      </template>
      <template #header-extra>
        <n-space :size="8">
          <span v-if="checkedKeys.length" class="text-xs text-gray-400">已选 {{ checkedKeys.length }} 条</span>
          <n-button size="small" type="success" :disabled="checkedKeys.length === 0" :loading="batching" @click="batchApprove">
            批量通过{{ checkedKeys.length ? ` (${checkedKeys.length})` : '' }}
          </n-button>
          <n-button size="small" type="warning" ghost :disabled="checkedKeys.length === 0" @click="openBatchReject">
            批量驳回
          </n-button>
          <n-button size="small" @click="load">刷新</n-button>
        </n-space>
      </template>

      <n-data-table
        v-model:checked-row-keys="checkedKeys"
        :columns="columns"
        :data="list"
        :loading="loading"
        :row-key="(row: PendingReviewItem) => row.id"
        :max-height="tableMaxHeight"
        size="small"
      />
    </n-card>

    <!-- 单条驳回：理由必填 -->
    <n-modal v-model:show="showReject" preset="dialog" title="驳回任务" :show-icon="false" style="width: 460px">
      <n-space vertical :size="8" class="py-2">
        <span class="text-sm">「{{ rejecting?.title }}」将回到进行中，理由将通知执行者</span>
        <n-input v-model:value="rejectComment" type="textarea" placeholder="驳回理由（必填）" :rows="3" />
      </n-space>
      <template #action>
        <n-space>
          <n-button size="small" @click="showReject = false">取消</n-button>
          <n-button size="small" type="error" :disabled="!rejectComment.trim()" :loading="rejectingLoading" @click="confirmReject">确认驳回</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 批量驳回：统一理由逐条落评论 -->
    <n-modal v-model:show="showBatchReject" preset="dialog" title="批量驳回" :show-icon="false" style="width: 460px">
      <n-space vertical :size="8" class="py-2">
        <span class="text-sm">已选 {{ checkedKeys.length }} 条将全部回到进行中，统一理由逐条落为任务评论并通知执行者</span>
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
  import { useRouter } from 'vue-router';
  import { NButton, NSpace, NTag, useMessage } from 'naive-ui';
  import type { DataTableColumns } from 'naive-ui';
  import { getPendingReviews, batchReview } from '@/api/platform/index';
  import type { PendingReviewItem } from '@/api/platform/index';
  import { reviewTask } from '@/api/project/index';
  import { priorityTagType, typeLabel, typeTagType } from '@/enums/task';

  const router = useRouter();
  const message = useMessage();

  const loading = ref(false);
  const list = ref<PendingReviewItem[]>([]);
  const total = ref(0);
  const checkedKeys = ref<number[]>([]);
  // 表格高度自适应视口（审核队列常超一屏，页内滚动而非整页滚动）
  const tableMaxHeight = computed(() => Math.max(320, window.innerHeight - 300));

  async function load() {
    loading.value = true;
    try {
      const res = await getPendingReviews();
      list.value = res?.list || [];
      total.value = res?.total || 0;
      // 刷新后清掉已不在队列中的勾选
      const ids = new Set(list.value.map((x) => x.id));
      checkedKeys.value = checkedKeys.value.filter((id) => ids.has(id));
    } catch {
      message.error('加载待审任务失败');
    } finally {
      loading.value = false;
    }
  }

  function dropIds(ids: number[]) {
    list.value = list.value.filter((t) => !ids.includes(t.id));
    total.value = Math.max(0, total.value - ids.length);
    checkedKeys.value = checkedKeys.value.filter((id) => !ids.includes(id));
  }

  function reportFailed(failed: Array<{ id: number; title: string; error: string }>) {
    if (failed.length) {
      message.warning(`${failed.length} 条失败：${failed.map((f) => `#${f.id} ${f.error}`).join('；')}`);
    }
  }

  // ---- 单条审核 ----
  async function approve(row: PendingReviewItem) {
    try {
      await reviewTask(row.id, { status: 'approved' });
      message.success(`「${row.title}」已通过`);
      dropIds([row.id]);
    } catch {
      message.error('审核操作失败');
    }
  }

  const showReject = ref(false);
  const rejecting = ref<PendingReviewItem | null>(null);
  const rejectComment = ref('');
  const rejectingLoading = ref(false);

  function openReject(row: PendingReviewItem) {
    rejecting.value = row;
    rejectComment.value = '';
    showReject.value = true;
  }

  async function confirmReject() {
    if (!rejecting.value || !rejectComment.value.trim()) return;
    rejectingLoading.value = true;
    try {
      await reviewTask(rejecting.value.id, { status: 'rejected', comment: rejectComment.value.trim() });
      message.success(`「${rejecting.value.title}」已驳回`);
      showReject.value = false;
      dropIds([rejecting.value.id]);
    } catch {
      message.error('驳回失败');
    } finally {
      rejectingLoading.value = false;
    }
  }

  // ---- 批量审核 ----
  const batching = ref(false);

  async function batchApprove() {
    if (!checkedKeys.value.length) return;
    batching.value = true;
    try {
      const res = await batchReview({ ids: [...checkedKeys.value], status: 'approved' });
      if (res?.succeeded) message.success(`已通过 ${res.succeeded} 条`);
      reportFailed(res?.failed || []);
      dropIds([...checkedKeys.value]);
      load();
    } catch {
      message.error('批量通过失败');
    } finally {
      batching.value = false;
    }
  }

  const showBatchReject = ref(false);
  const batchRejectComment = ref('');

  function openBatchReject() {
    batchRejectComment.value = '';
    showBatchReject.value = true;
  }

  async function confirmBatchReject() {
    if (!checkedKeys.value.length || !batchRejectComment.value.trim()) return;
    batching.value = true;
    try {
      const res = await batchReview({
        ids: [...checkedKeys.value],
        status: 'rejected',
        comment: batchRejectComment.value.trim(),
      });
      if (res?.succeeded) message.success(`已驳回 ${res.succeeded} 条（回到进行中）`);
      reportFailed(res?.failed || []);
      showBatchReject.value = false;
      dropIds([...checkedKeys.value]);
      load();
    } catch {
      message.error('批量驳回失败');
    } finally {
      batching.value = false;
    }
  }

  const columns = computed<DataTableColumns<PendingReviewItem>>(() => [
    { type: 'selection' },
    {
      type: 'expand',
      renderExpand: (row) =>
        h('div', { class: 'review-expand' }, [
          h('div', { class: 'review-expand-label' }, '任务描述'),
          h('div', { class: 'review-expand-desc' }, row.description || '（无描述）'),
        ]),
    },
    {
      title: '任务',
      key: 'title',
      render: (row) =>
        h(NSpace, { size: 6, align: 'center' }, () => [
          h(
            NButton,
            { text: true, type: 'info', onClick: () => gotoTask(row) },
            () => row.title,
          ),
        ]),
    },
    {
      title: '项目',
      key: 'projectName',
      width: 150,
      render: (row) =>
        h(NTag, { size: 'small', bordered: false }, () => row.projectName || `#${row.projectId}`),
    },
    {
      title: '类型',
      key: 'type',
      width: 90,
      render: (row) => h(NTag, { size: 'small', type: typeTagType(row.type) }, () => typeLabel(row.type)),
    },
    {
      title: '优先级',
      key: 'priority',
      width: 80,
      render: (row) => h(NTag, { size: 'small', type: priorityTagType(row.priority) }, () => `P${row.priority}`),
    },
    { title: '提交人', key: 'assigneeName', width: 110, render: (row) => row.assigneeName || '-' },
    { title: '更新时间', key: 'updatedAt', width: 150, render: (row) => (row.updatedAt || '').slice(0, 16) },
    {
      title: '操作',
      key: 'actions',
      width: 130,
      render: (row) =>
        h(NSpace, { size: 4 }, () => [
          h(NButton, { size: 'tiny', type: 'success', onClick: () => approve(row) }, () => '通过'),
          h(NButton, { size: 'tiny', type: 'warning', ghost: true, onClick: () => openReject(row) }, () => '驳回'),
        ]),
    },
  ]);

  // 需要完整上下文（artifacts/评论/AI 产出）时跳任务详情
  function gotoTask(row: PendingReviewItem) {
    router.push(`/project/${row.projectId}/tasks?task=${row.id}`);
  }

  onMounted(load);
</script>

<style lang="less" scoped>
  :deep(.review-expand) {
    padding: 4px 8px;

    .review-expand-label {
      font-size: 12px;
      font-weight: 600;
      color: var(--text-color-2, #5c6470);
      margin-bottom: 4px;
    }

    .review-expand-desc {
      font-size: 13px;
      color: var(--text-color-2, #5c6470);
      white-space: pre-wrap;
      line-height: 1.7;
      max-width: 720px;
    }
  }
</style>
