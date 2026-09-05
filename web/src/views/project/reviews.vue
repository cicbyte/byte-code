<template>
  <div>
    <n-card :bordered="false" title="待审核" class="proCard">
      <template #header-extra>
        <n-tag v-if="pending.length" type="warning" size="small">{{ pending.length }} 条待审</n-tag>
      </template>

      <n-spin :show="loading">
        <EmptyState
          v-if="!loading && pending.length === 0"
          type="task"
          title="没有待审核的任务"
          description="Agent 或成员完成任务后会进入这里等待验收"
        />
        <n-table v-else :bordered="false" :single-line="false" size="small">
          <thead>
            <tr>
              <th style="width: 34%">标题</th>
              <th>提交人</th>
              <th>优先级</th>
              <th>完成时间</th>
              <th style="width: 160px">操作</th>
            </tr>
          </thead>
          <tbody>
            <template v-for="t in pending" :key="t.id">
              <tr>
                <td>
                  <n-button text type="info" @click="toggle(t.id)">
                    {{ expanded === t.id ? '▾' : '▸' }} {{ t.title }}
                  </n-button>
                </td>
                <td>
                  <n-space :size="4" align="center">
                    <span>{{ t.assigneeName || '-' }}</span>
                    <n-tag v-if="t.source === 'agent'" size="tiny" :bordered="false" type="info">Agent</n-tag>
                  </n-space>
                </td>
                <td><n-tag :type="priorityTagType(t.priority)" size="small">P{{ t.priority }}</n-tag></td>
                <td class="text-xs">{{ (t.updatedAt || '').slice(0, 16) }}</td>
                <td>
                  <n-space :size="8">
                    <n-popconfirm @positive-click="approve(t)">
                      <template #trigger>
                        <n-button size="tiny" type="success">通过</n-button>
                      </template>
                      通过后任务转入已完成（done）。
                    </n-popconfirm>
                    <n-button size="tiny" type="error" ghost @click="openReject(t)">驳回</n-button>
                  </n-space>
                </td>
              </tr>
              <tr v-if="expanded === t.id">
                <td colspan="5" class="review-detail">
                  <div v-if="t.description" class="mb-2">
                    <div class="detail-label">任务描述</div>
                    <div class="text-sm">{{ t.description }}</div>
                  </div>
                  <div>
                    <div class="detail-label">Agent 产出</div>
                    <MdPreview
                      v-if="t.artifacts"
                      :id="`md-review-${t.id}`"
                      :model-value="safeHtml(t.artifacts)"
                    />
                    <div v-else class="text-xs text-gray-400">（无产出内容）</div>
                  </div>
                </td>
              </tr>
            </template>
          </tbody>
        </n-table>
      </n-spin>
    </n-card>

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
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { useMessage } from 'naive-ui';
  import DOMPurify from 'dompurify';
  import { MdPreview } from 'md-editor-v3';
  import 'md-editor-v3/lib/preview.css';
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { getTasks, reviewTask } from '@/api/project/index';
  import type { TaskItem } from '@/api/project/index';
  import { priorityTagType } from '@/enums/task';

  const route = useRoute();
  const message = useMessage();
  const projectId = computed(() => Number(route.params.projectId));

  const loading = ref(false);
  // 待审 = 完成进 review 且 requiresHumanReview=1 / humanReviewStatus=pending
  // （CompleteTask 统一置位；审完 hrs 变化自然出队）
  const list = ref<TaskItem[]>([]);
  const pending = computed(() =>
    list.value.filter((t) => t.requiresHumanReview === 1 && t.humanReviewStatus === 'pending')
  );

  const expanded = ref<number | null>(null);
  const showReject = ref(false);
  const rejectComment = ref('');
  const rejecting = ref<TaskItem | null>(null);

  function toggle(id: number) {
    expanded.value = expanded.value === id ? null : id;
  }

  // artifacts 来自 agent 产出，渲染前消毒（与任务详情抽屉同口径）
  function safeHtml(md: string): string {
    return DOMPurify.sanitize(md);
  }

  async function load() {
    loading.value = true;
    try {
      const res = await getTasks(projectId.value, { status: 'review', size: 200 });
      list.value = res?.list || [];
    } catch {
      message.error('加载待审任务失败');
    } finally {
      loading.value = false;
    }
  }

  function drop(id: number) {
    list.value = list.value.filter((t) => t.id !== id);
    if (expanded.value === id) expanded.value = null;
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

  onMounted(load);
</script>

<style lang="less" scoped>
  .review-detail {
    background: var(--hover-bg, #fafafa);
    padding: 12px 16px;
  }
  .detail-label {
    font-size: 12px;
    color: var(--text-color-3, #999);
    margin-bottom: 4px;
  }
</style>
