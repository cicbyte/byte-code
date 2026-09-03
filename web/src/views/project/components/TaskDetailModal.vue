<template>
  <n-modal
    v-model:show="visible"
    preset="card"
    title="任务详情"
    style="width: 720px; max-height: 80vh"
    :mask-closable="true"
  >
    <n-spin :show="loading">
      <template v-if="task">
        <!-- 基本信息 -->
        <n-descriptions label-placement="left" :column="2" bordered size="small" class="mb-4">
          <n-descriptions-item label="标题" :span="2">
            <span class="font-medium">{{ task.title }}</span>
          </n-descriptions-item>
          <n-descriptions-item label="状态">
            <n-tag size="small">{{ task.status }}</n-tag>
          </n-descriptions-item>
          <n-descriptions-item label="类型">
            <n-tag size="small" type="info">{{ task.type }}</n-tag>
          </n-descriptions-item>
          <n-descriptions-item label="优先级">
            <n-tag size="small" :type="priorityType">P{{ task.priority }}</n-tag>
          </n-descriptions-item>
          <n-descriptions-item label="指派人">{{ task.assigneeName || '-' }}</n-descriptions-item>
          <n-descriptions-item label="创建人">{{ task.creatorName || '-' }}</n-descriptions-item>
          <n-descriptions-item label="创建时间">{{ task.createdAt }}</n-descriptions-item>
        </n-descriptions>

        <!-- 描述 -->
        <n-card title="描述" size="small" class="mb-4" :bordered="true" :segmented="{ content: true }">
          <div v-if="task.description" v-html="safeDescription" class="prose max-w-none"></div>
          <n-empty v-else description="暂无描述" size="small" />
        </n-card>

        <!-- 审核操作 -->
        <n-card
          v-if="task.requiresHumanReview && task.humanReviewStatus === 'pending'"
          title="审核操作"
          size="small"
          class="mb-4"
          :bordered="true"
        >
          <n-space>
            <n-button type="success" @click="handleReview('approved')">通过</n-button>
            <n-button type="error" @click="handleReview('rejected')">驳回</n-button>
          </n-space>
        </n-card>

        <!-- 评论区 -->
        <n-card title="评论" size="small" :bordered="true" :segmented="{ content: true }">
          <n-space vertical :size="12">
            <n-empty v-if="comments.length === 0" description="暂无评论" size="small" />
            <div v-for="comment in comments" :key="comment.id" class="border-b border-gray-100 pb-2 last:border-0">
              <n-space justify="space-between" align="center">
                <n-space size="small" align="center">
                  <n-tag size="tiny" :type="comment.userType === 'ai' ? 'warning' : 'info'">
                    {{ comment.userType === 'ai' ? 'AI' : '用户' }}
                  </n-tag>
                  <span class="text-sm font-medium">{{ comment.realName || comment.username }}</span>
                </n-space>
                <span class="text-xs text-gray-400">{{ comment.createdAt }}</span>
              </n-space>
              <p class="mt-1 text-sm text-gray-600">{{ comment.content }}</p>
            </div>
          </n-space>
          <n-divider />
          <n-space>
            <n-input
              v-model:value="newComment"
              type="textarea"
              placeholder="输入评论内容"
              :rows="2"
              style="flex: 1"
            />
            <n-button type="primary" @click="handleAddComment" :loading="submitting">发送</n-button>
          </n-space>
        </n-card>
      </template>
    </n-spin>

    <template #footer>
      <n-space justify="end">
        <n-button @click="visible = false">关闭</n-button>
      </n-space>
    </template>
  </n-modal>
</template>

<script lang="ts" setup>
  import DOMPurify from 'dompurify';
  import { ref, computed } from 'vue';
  import { useMessage } from 'naive-ui';
  import {
    getTask,
    getComments,
    createComment,
    reviewTask,
  } from '@/api/project/index';
  import type { TaskItem, CommentItem } from '@/api/project/index';


  // description 来自用户输入，渲染前消毒（历史遗留的裸 v-html 注入面）
  const safeDescription = computed(() => DOMPurify.sanitize(task.value?.description || ''));
  const message = useMessage();
  const visible = ref(false);
  const loading = ref(false);
  const submitting = ref(false);
  const task = ref<TaskItem | null>(null);
  const comments = ref<CommentItem[]>([]);
  const newComment = ref('');

  const priorityType = computed(() => {
    const p = task.value?.priority;
    if (p === undefined || p === null) return 'default';
    if (p >= 4) return 'error';
    if (p >= 3) return 'warning';
    if (p >= 2) return 'info';
    return 'default';
  });

  async function openModal(taskId: number) {
    visible.value = true;
    loading.value = true;
    try {
      const [taskRes, commentRes] = await Promise.all([
        getTask(taskId),
        getComments(taskId),
      ]);
      task.value = taskRes || null;
      comments.value = commentRes?.list || [];
    } catch (e) {
      message.error('加载任务详情失败');
    } finally {
      loading.value = false;
    }
  }

  async function handleAddComment() {
    if (!newComment.value.trim()) {
      message.warning('请输入评论内容');
      return;
    }
    if (!task.value) return;
    submitting.value = true;
    try {
      await createComment(task.value.id, { content: newComment.value });
      message.success('评论发送成功');
      newComment.value = '';
      const commentRes = await getComments(task.value.id);
      comments.value = commentRes?.list || [];
    } catch (e) {
      message.error('评论发送失败');
    } finally {
      submitting.value = false;
    }
  }

  async function handleReview(status: 'approved' | 'rejected') {
    if (!task.value) return;
    try {
      await reviewTask(task.value.id, { status });
      message.success(status === 'approved' ? '已通过' : '已驳回');
      const taskRes = await getTask(task.value.id);
      task.value = taskRes || null;
    } catch (e) {
      message.error('审核操作失败');
    }
  }

  defineExpose({ openModal });
</script>
