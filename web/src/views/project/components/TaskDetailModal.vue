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

        <!-- 标签 -->
        <n-card title="标签" size="small" class="mb-4" :bordered="true" :segmented="{ content: true }">
          <n-space :size="4" align="center">
            <n-tag
              v-for="t in task.tags || []"
              :key="t"
              size="small"
              round
              closable
              @close="detachByName(t)"
            >{{ t }}</n-tag>
            <n-select
              v-model:value="attachTagId"
              :options="attachableTagOptions"
              size="tiny"
              filterable
              tag
              placeholder="+ 挂标签（输入回车可新建）"
              style="width: 200px"
              :show-arrow="false"
              @update:value="onAttachSelect"
            />
          </n-space>
        </n-card>

        <!-- 附件 -->
        <n-card title="附件" size="small" class="mb-4" :bordered="true" :segmented="{ content: true }">
          <n-space vertical :size="6">
            <n-empty v-if="attachments.length === 0" description="暂无附件" size="small" />
            <n-space v-for="a in attachments" :key="a.id" justify="space-between" align="center" class="w-full">
              <n-space :size="8" align="center">
                <n-icon size="14"><PaperClipOutlined /></n-icon>
                <span class="text-sm">{{ a.originalName }}</span>
                <span class="text-xs text-gray-400">{{ fmtSize(a.fileSize) }}</span>
                <span class="text-xs text-gray-400">{{ a.uploaderName || '' }}</span>
              </n-space>
              <n-space :size="2">
                <n-button text type="info" size="tiny" @click="downloadAtt(a)">下载</n-button>
                <n-button text type="error" size="tiny" @click="removeAtt(a)">删除</n-button>
              </n-space>
            </n-space>
            <n-space>
              <n-button size="tiny" :loading="uploadingAtt" @click="attInputRef?.click()">上传附件</n-button>
            </n-space>
          </n-space>
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
    <input ref="attInputRef" type="file" multiple style="display: none" @change="onAttFiles" />

    <template #footer>
      <n-space justify="end">
        <n-button @click="visible = false">关闭</n-button>
      </n-space>
    </template>
  </n-modal>
</template>

<script lang="ts" setup>
  import DOMPurify from 'dompurify';
  import { getTags, attachTag, detachTag, createTag } from '@/api/platform/index';
  import { getAttachments, uploadAttachment, downloadAttachment, deleteAttachment } from '@/api/attachment/index';
  import { PaperClipOutlined } from '@vicons/antd';
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

  // 标签管理：attach 选择（含新建：tag 模式回车创建后挂载）
  const attachTagId = ref<number | null>(null);
  const allTags = ref<Array<{ id: number; name: string }>>([]);

  const attachableTagOptions = computed(() => {
    const attached = new Set(task.value?.tags || []);
    return allTags.value
      .filter((t) => !attached.has(t.name))
      .map((t) => ({ label: t.name, value: t.id }));
  });

  async function loadAllTags() {
    try {
      const res = await getTags();
      allTags.value = (res?.list || []).map((t: any) => ({ id: t.id, name: t.name }));
    } catch {
      // ignore
    }
  }

  async function onAttachSelect(val: number | null) {
    if (!task.value || !val) return;
    try {
      await attachTag(val, { entityType: 'task', entityId: task.value.id });
      const t = allTags.value.find((x) => x.id === val);
      if (t) task.value.tags = [...(task.value.tags || []), t.name];
      attachTagId.value = null;
    } catch {
      message.error('挂标签失败');
      attachTagId.value = null;
    }
  }

  async function detachByName(name: string) {
    if (!task.value) return;
    const t = allTags.value.find((x) => x.name === name);
    if (!t) return;
    try {
      await detachTag(t.id, 'task', task.value.id);
      task.value.tags = (task.value.tags || []).filter((x) => x !== name);
    } catch {
      message.error('移除标签失败');
    }
  }
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
    loadAllTags();
    loadAttachments(taskId);
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

  // ==================== 附件 ====================
  const attachments = ref<any[]>([]);
  const uploadingAtt = ref(false);
  const attInputRef = ref<HTMLInputElement | null>(null);

  function fmtSize(n: number) {
    if (!n) return '';
    if (n < 1024) return `${n} B`;
    if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
    return `${(n / 1024 / 1024).toFixed(1)} MB`;
  }

  async function loadAttachments(taskId: number) {
    try {
      const res = await getAttachments('task', taskId);
      attachments.value = res?.list || [];
    } catch {
      attachments.value = [];
    }
  }

  async function onAttFiles(e: Event) {
    const input = e.target as HTMLInputElement;
    const files = Array.from(input.files || []);
    if (!files.length || !task.value) return;
    uploadingAtt.value = true;
    for (const f of files) {
      try {
        await uploadAttachment({ entityType: 'task', entityId: task.value.id, file: f });
      } catch {
        message.error(`上传失败：${f.name}`);
      }
    }
    uploadingAtt.value = false;
    input.value = '';
    loadAttachments(task.value.id);
  }

  async function downloadAtt(a: any) {
    try {
      const res = await downloadAttachment(a.id);
      if (res?.url) window.open(res.url, '_blank');
    } catch {
      message.error('获取下载链接失败');
    }
  }

  async function removeAtt(a: any) {
    try {
      await deleteAttachment(a.id);
      attachments.value = attachments.value.filter((x) => x.id !== a.id);
    } catch {
      message.error('删除附件失败');
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
