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
            <n-tag size="small" :type="statusTagType(task.status)">{{ statusLabel(task.status) }}</n-tag>
          </n-descriptions-item>
          <n-descriptions-item label="类型">
            <n-tag size="small" type="info">{{ task.type }}</n-tag>
          </n-descriptions-item>
          <n-descriptions-item label="优先级">
            <n-tag size="small" :type="priorityType">P{{ task.priority }}</n-tag>
          </n-descriptions-item>
          <n-descriptions-item label="指派人">{{ task.assigneeName || '-' }}</n-descriptions-item>
          <n-descriptions-item label="截止日期">
            <n-space :size="4" align="center">
              <n-tag
                v-if="task.dueDate"
                size="small"
                :type="dueTagType(task.dueDate, task.status) || 'default'"
              >{{ dueLabel(task.dueDate) }}</n-tag>
              <n-date-picker
                :formatted-value="task.dueDate || null"
                type="date"
                value-format="yyyy-MM-dd"
                size="tiny"
                clearable
                placeholder="设置截止"
                style="width: 140px"
                @update:formatted-value="onDueChange"
              />
            </n-space>
          </n-descriptions-item>
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

        <!-- 关联文档 -->
        <n-card title="关联文档" size="small" class="mb-4" :bordered="true" :segmented="{ content: true }">
          <n-empty v-if="linkedDocs.length === 0" description="暂无关联文档（文档 frontmatter linked 指向本任务时出现）" size="small" />
          <n-space v-else vertical :size="4">
            <n-space
              v-for="d in linkedDocs"
              :key="d.path"
              align="center"
              :size="8"
              class="linked-doc"
              @click="gotoDoc(d)"
            >
              <n-tag size="tiny" :type="d.space === 'knowledge' ? 'success' : 'info'" :bordered="false">
                {{ d.space === 'knowledge' ? '知识库' : '文档' }}
              </n-tag>
              <n-icon size="14" class="text-gray-400"><FileTextOutlined /></n-icon>
              <span class="text-sm">{{ d.title || d.path }}</span>
              <span class="text-xs text-gray-400">{{ d.path }}</span>
            </n-space>
          </n-space>
        </n-card>

        <!-- AI 执行日志时间线 -->
        <n-card title="AI 执行日志" size="small" class="mb-4" :bordered="true" :segmented="{ content: true }">
          <n-empty v-if="aiLogs.length === 0" description="暂无 AI 执行记录" size="small" />
          <n-timeline v-else>
            <n-timeline-item
              v-for="log in aiLogs"
              :key="log.id"
              :type="logTimelineType(log)"
              :title="`${logTitle(log)}${log.aiUsername ? ' · ' + log.aiUsername : ''}`"
              :time="log.createdAt"
            >
              <template #default>
                <div v-if="log.detail">
                  <div class="text-xs text-gray-600 whitespace-pre-wrap">{{ shortDetail(log) }}</div>
                  <n-button
                    v-if="log.detail.length > 200"
                    text
                    type="primary"
                    size="tiny"
                    @click="expandedLogs.has(log.id) ? expandedLogs.delete(log.id) : expandedLogs.add(log.id)"
                  >{{ expandedLogs.has(log.id) ? '收起' : '展开全文' }}</n-button>
                </div>
              </template>
            </n-timeline-item>
          </n-timeline>
        </n-card>

        <!-- AI 产出（artifacts，markdown 渲染） -->
        <n-card
          v-if="task.artifacts"
          title="AI 产出"
          size="small"
          class="mb-4"
          :bordered="true"
          :segmented="{ content: true }"
        >
          <MdPreview :id="mdPreviewId" :model-value="task.artifacts" :sanitize="sanitizeHtml" />
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
                <n-space :size="4" align="center">
                  <span class="text-xs text-gray-400">{{ comment.createdAt }}</span>
                  <template v-if="isMyComment(comment)">
                    <n-button text type="primary" size="tiny" @click="startEditComment(comment)">编辑</n-button>
                    <n-button text type="error" size="tiny" @click="handleDeleteComment(comment)">删除</n-button>
                  </template>
                </n-space>
              </n-space>
              <template v-if="editingCommentId === comment.id">
                <n-input v-model:value="editCommentContent" type="textarea" :rows="2" />
                <n-space :size="4" class="mt-1" justify="end">
                  <n-button size="tiny" @click="cancelEditComment">取消</n-button>
                  <n-button size="tiny" type="primary" :loading="savingComment" @click="saveEditComment">保存</n-button>
                </n-space>
              </template>
              <p v-else class="mt-1 text-sm text-gray-600 comment-body" v-html="renderComment(comment.content)"></p>
            </div>
          </n-space>
          <n-divider />
          <n-space vertical :size="4">
            <n-mention
              v-model:value="newComment"
              type="textarea"
              placeholder="输入评论内容，@ 可提及成员（将发送通知）"
              :rows="2"
              :options="mentionOptions"
              :prefix="['@']"
              style="flex: 1"
            />
            <n-space justify="end">
              <n-button type="primary" @click="handleAddComment" :loading="submitting">发送</n-button>
            </n-space>
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
  import { MdPreview } from 'md-editor-v3';
  import 'md-editor-v3/lib/preview.css';
  import { getTags, attachTag, detachTag, createTag } from '@/api/platform/index';
  import { getAttachments, uploadAttachment, downloadAttachment, deleteAttachment } from '@/api/attachment/index';
  import { getLinkedDocs } from '@/api/docs/index';
  import type { DocsSearchItem } from '@/api/docs/index';
  import { PaperClipOutlined, FileTextOutlined } from '@vicons/antd';
  import { ref, computed } from 'vue';
  import { useRouter } from 'vue-router';
  import { useMessage, useDialog } from 'naive-ui';
  import {
    getTask,
    getComments,
    createComment,
    reviewTask,
    getAiLogs,
    updateTask,
    getMembers,
    updateComment,
    deleteComment,
  } from '@/api/project/index';
  import type { TaskItem, CommentItem, AiLogItem } from '@/api/project/index';
  import { useUserStore } from '@/store/modules/user';
  import { dueTagType, dueLabel } from '@/utils/taskDue';
  import { priorityTagType, statusLabel, statusTagType } from '@/enums/task';


  // description 来自用户输入，渲染前消毒（历史遗留的裸 v-html 注入面）
  const safeDescription = computed(() => DOMPurify.sanitize(task.value?.description || ''));

  // 评论正文：纯文本渲染但高亮 @提及——先整体 HTML 转义，再把 @词 包上样式 span，
  // 最后 DOMPurify 兜底（转义后内容理论无标签，消毒是纵深防御）
  function renderComment(content: string): string {
    const escaped = content
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
    const highlighted = escaped.replace(
      /(@[\w\u4e00-\u9fa5-]+)/g,
      '<span class="mention">$1</span>'
    );
    return DOMPurify.sanitize(highlighted);
  }
  const message = useMessage();
  const dialog = useDialog();
  const userStore = useUserStore();
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

  // @提及候选：项目成员中的 human（AI 成员由引擎自调度，@提及不产生通知，
  // 列出只会造成"提及了却没通知"的困惑）；插入值为 username（与后端
  // notifyMentions 的 @用户名 精确匹配口径一致——显示名仅在候选列表里辅助识别）
  const mentionOptions = ref<Array<{ label: string; value: string }>>([]);
  async function loadMentionOptions(projectId: number) {
    try {
      const res = await getMembers(projectId);
      mentionOptions.value = (res?.list || [])
        .filter((m) => !m.userType || m.userType === 'human')
        .map((m) => ({
          label: `${m.realName || m.username}（${m.username}）`,
          value: m.username,
        }));
    } catch {
      mentionOptions.value = [];
    }
  }

  const router = useRouter();

  // ==================== 关联文档 / AI 日志 / AI 产出 ====================
  const linkedDocs = ref<DocsSearchItem[]>([]);
  const aiLogs = ref<AiLogItem[]>([]);
  const expandedLogs = ref(new Set<number>());
  const mdPreviewId = 'task-artifacts-preview';

  // md-editor-v3 默认 html:true 不消毒——artifacts 是 AI 产出文本，渲染前必须过 DOMPurify
  const sanitizeHtml = (html: string) => DOMPurify.sanitize(html);

  async function loadLinkedDocs(projectId: number, taskId: number) {
    try {
      const res = await getLinkedDocs(projectId, `task:${taskId}`);
      linkedDocs.value = res?.items || [];
    } catch {
      linkedDocs.value = [];
    }
  }

  async function loadAiLogs(taskId: number) {
    try {
      const res = await getAiLogs(taskId);
      aiLogs.value = res?.list || [];
    } catch {
      aiLogs.value = [];
    }
  }

  function gotoDoc(d: DocsSearchItem) {
    if (!task.value) return;
    // 知识库与工作区是两个视图；按文档所属空间跳对应页面，query.path 由 docs.vue 直达打开
    const view = d.space === 'knowledge' ? 'knowledge' : 'docs';
    visible.value = false;
    router.push({ path: `/project/${task.value.projectId}/${view}`, query: { path: d.path } });
  }

  const actionLabels: Record<string, string> = {
    claim: '认领任务',
    complete: '完成任务',
    error: '执行出错',
    dead: '转入死信',
  };

  function logTitle(log: AiLogItem) {
    return actionLabels[log.action] || log.action;
  }

  function logTimelineType(log: AiLogItem): 'default' | 'info' | 'success' | 'error' | 'warning' {
    if (log.status === 'failed') return 'error';
    switch (log.action) {
      case 'claim': return 'info';
      case 'complete': return 'success';
      default: return 'default';
    }
  }

  function shortDetail(log: AiLogItem) {
    if (expandedLogs.value.has(log.id) || log.detail.length <= 200) return log.detail;
    return log.detail.slice(0, 200) + '…';
  }

  const priorityType = computed(() => priorityTagType(task.value?.priority));

  async function openModal(taskId: number) {
    visible.value = true;
    loading.value = true;
    expandedLogs.value = new Set();
    linkedDocs.value = [];
    aiLogs.value = [];
    loadAllTags();
    loadAttachments(taskId);
    try {
      const [taskRes, commentRes] = await Promise.all([
        getTask(taskId),
        getComments(taskId),
      ]);
      task.value = taskRes || null;
      comments.value = commentRes?.list || [];
      // 关联文档/AI 日志依赖任务归属项目，getTask 返回后再加载
      if (task.value) {
        loadLinkedDocs(task.value.projectId, taskId);
        loadAiLogs(taskId);
        loadMentionOptions(task.value.projectId);
      }
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

  // ==================== 评论编辑/删除（仅作者本人） ====================
  const myUsername = computed(() => userStore.getUserInfo?.username || userStore.username || '');

  function isMyComment(c: CommentItem) {
    return c.userType !== 'ai' && !!c.username && c.username === myUsername.value;
  }

  const editingCommentId = ref<number | null>(null);
  const editCommentContent = ref('');
  const savingComment = ref(false);

  function startEditComment(c: CommentItem) {
    editingCommentId.value = c.id;
    editCommentContent.value = c.content;
  }

  function cancelEditComment() {
    editingCommentId.value = null;
    editCommentContent.value = '';
  }

  async function saveEditComment() {
    if (!editingCommentId.value || !editCommentContent.value.trim()) {
      message.warning('评论内容不能为空');
      return;
    }
    savingComment.value = true;
    try {
      await updateComment(editingCommentId.value, editCommentContent.value);
      message.success('评论已更新');
      cancelEditComment();
      if (task.value) {
        const res = await getComments(task.value.id);
        comments.value = res?.list || [];
      }
    } catch {
      message.error('更新评论失败');
    } finally {
      savingComment.value = false;
    }
  }

  function handleDeleteComment(c: CommentItem) {
    dialog.warning({
      title: '确认删除评论',
      content: '删除后不可恢复，其下所有回复将一并删除。',
      positiveText: '删除',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await deleteComment(c.id);
          message.success('已删除');
          comments.value = comments.value.filter((x) => x.id !== c.id);
        } catch {
          message.error('删除评论失败');
        }
      },
    });
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

  // 截止日期内联修改：null=清除（传空串），失败回读旧值
  async function onDueChange(val: string | null) {
    if (!task.value) return;
    const old = task.value.dueDate;
    try {
      await updateTask(task.value.id, { dueDate: val || '' });
      task.value.dueDate = val || '';
      message.success(val ? `截止日期已设为 ${val}` : '已清除截止日期');
    } catch {
      task.value.dueDate = old;
      message.error('更新截止日期失败');
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

<style lang="less" scoped>
  .linked-doc {
    padding: 4px 8px;
    margin: 0 -8px;
    border-radius: 4px;
    cursor: pointer;

    &:hover {
      background: var(--hover-bg, #f5f5f5);
    }
  }

  :deep(.mention) {
    color: #18a058;
    font-weight: 500;
    background: rgb(24 160 88 / 8%);
    border-radius: 3px;
    padding: 0 3px;
  }
</style>
