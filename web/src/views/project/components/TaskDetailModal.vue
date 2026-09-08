<template>
  <n-drawer
    v-model:show="visible"
    class="task-detail-drawer"
    placement="right"
    :width="drawerWidth"
    :mask="false"
    :close-on-esc="true"
  >
    <n-drawer-content title="任务详情" closable>
    <n-spin :show="loading">
      <template v-if="task">
        <!-- 吸顶锚点：长抽屉快速跳转（胶囊分段风格，点击滚动到对应卡片） -->
        <div class="section-nav">
          <button
            v-for="sec in navSections"
            :key="sec.id"
            class="nav-chip"
            :class="{ active: activeSec === sec.id }"
            @click="scrollToSection(sec.id)"
          >
            {{ sec.label }}
          </button>
        </div>

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

        <!-- 描述：markdown 渲染（agent 侧描述习惯 md；与 artifacts 同款
             MdPreview + DOMPurify 消毒，替换原裸 v-html） -->
        <n-card id="sec-desc" title="描述" size="small" class="mb-4" :bordered="true" :segmented="{ content: true }">
          <MdPreview
            v-if="task.description"
            :id="`md-desc-${task.id}`"
            :model-value="task.description"
            :sanitize="sanitizeHtml"
          />
          <EmptyState type="doc" title="暂无描述" v-else compact />
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
              @create="onTagCreate"
            />
          </n-space>
        </n-card>

        <!-- 附件 -->
        <n-card title="附件" size="small" class="mb-4" :bordered="true" :segmented="{ content: true }">
          <n-space vertical :size="6">
            <EmptyState type="doc" title="暂无附件" v-if="attachments.length === 0" compact />
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
          <EmptyState type="doc" title="暂无关联文档" v-if="linkedDocs.length === 0" description="在文档 frontmatter 的 linked 中指向本任务即可关联" compact />
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

        <!-- 跨项目引用：bug 的发现地与修复地分离；点击跳转被引任务 -->
        <n-card id="sec-refs" title="跨项目引用" size="small" class="mb-4" :bordered="true">
          <template #header-extra>
            <n-button text size="tiny" type="primary" @click="showRefModal = true">+ 添加引用</n-button>
          </template>
          <div v-if="relatedRefs.length === 0" class="text-xs text-gray-400">
            未关联其他项目的实体
          </div>
          <div
            v-for="(r, i) in relatedRefs"
            :key="i"
            class="ref-row"
            @click="openRef(r)"
          >
            <svg class="ic xs"><use href="#i-switch"/></svg>
            <span class="ref-title">{{ r.title }}</span>
            <n-tag size="tiny" :bordered="false">{{ refProjectLabel(r.projectId) }}</n-tag>
            <n-button text size="tiny" type="error" @click.stop="removeRef(i)">移除</n-button>
          </div>
        </n-card>

        <!-- AI 执行日志时间线 -->
        <n-card id="sec-logs" title="Agent 执行日志" size="small" class="mb-4" :bordered="true" :segmented="{ content: true }">
          <EmptyState type="generic" title="暂无 Agent 执行记录" v-if="aiLogs.length === 0" description="Agent 认领与执行的过程会留痕在这里" compact />
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
          id="sec-artifacts"
          v-if="task.artifacts"
          title="Agent 产出"
          size="small"
          class="mb-4"
          :bordered="true"
          :segmented="{ content: true }"
        >
          <MdPreview :id="mdPreviewId" :model-value="task.artifacts" :sanitize="sanitizeHtml" />
        </n-card>

        <!-- 步骤清单：长任务工作流的执行步骤；打勾即进展（触发 updated_at
             构成租约心跳），agent 恢复上下文时据此知道做到第几步 -->
        <n-card id="sec-checklist" title="步骤清单" size="small" class="mb-4" :bordered="true">
          <template #header-extra>
            <span v-if="checklist.length" class="text-xs text-gray-400">
              {{ checklist.filter((c) => c.done).length }}/{{ checklist.length }}
            </span>
          </template>
          <div v-if="checklist.length === 0" class="text-xs text-gray-400">
            未设置步骤（勾选即进展，可跨会话接续）
          </div>
          <div v-for="(item, i) in checklist" :key="i" class="checklist-row">
            <n-checkbox
              :checked="item.done"
              @update:checked="(v: boolean) => toggleCheck(i, v)"
            >
              <span :class="{ 'checklist-done': item.done }">{{ item.text }}</span>
            </n-checkbox>
            <n-button text size="tiny" type="error" @click="removeCheck(i)">删</n-button>
          </div>
          <n-input
            v-model:value="newCheckText"
            size="small"
            placeholder="添加步骤，回车确认"
            @keydown.enter.exact.prevent="addCheck"
          />
        </n-card>

        <!-- 子任务：需求导入的父子血缘（详情回填一级），点行打开对应详情 -->
        <n-card v-if="(task.subTasks || []).length" title="子任务" size="small" class="mb-4" :bordered="true">
          <div v-for="st in task.subTasks" :key="st.id" class="subtask-row" @click="openTask(st.id)">
            <n-tag size="small" :type="statusTagType(st.status)">{{ statusLabel(st.status) }}</n-tag>
            <span class="subtask-title">#{{ st.id }} {{ st.title }}</span>
            <span class="subtask-assignee">{{ st.assigneeName || '未指派' }}</span>
          </div>
        </n-card>

        <!-- 阻塞操作：in_progress 可上报（等信息/环境问题），blocked 可解除。
             显示从宽（后端门禁把关 assignee/owner），阻塞原因在下方 Agent
             执行日志时间线留痕可见 -->
        <n-card v-if="task.status === 'in_progress' || task.status === 'blocked'" title="阻塞" size="small" class="mb-4" :bordered="true">
          <n-space>
            <n-button v-if="task.status === 'in_progress'" size="small" type="warning" ghost @click="showBlockModal = true">
              上报阻塞
            </n-button>
            <n-popconfirm v-else @positive-click="handleUnblock">
              <template #trigger>
                <n-button size="small" type="success">解除阻塞</n-button>
              </template>
              解除后任务回到进行中，由负责人继续执行。
            </n-popconfirm>
          </n-space>
        </n-card>

        <!-- 终态操作：done 可关闭清账；done/closed 实测发现问题均可重开回炉 -->
        <n-card
          v-if="task.status === 'done' || task.status === 'closed'"
          title="收尾" size="small" class="mb-4" :bordered="true"
        >
          <n-space>
            <n-button size="small" type="error" ghost @click="showReopenModal = true">
              重新打开
            </n-button>
            <n-popconfirm v-if="task.status === 'done'" @positive-click="handleClose">
              <template #trigger>
                <n-button size="small" type="warning" ghost>关闭任务</n-button>
              </template>
              关闭后任务转入终态（closed），不再出现在活跃列表；实测有问题可随时重新打开。
            </n-popconfirm>
          </n-space>
        </n-card>

    <!-- 审核操作 -->
        <n-card
          id="sec-review"
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

        <!-- 评论区（IM 对话流：自己右对齐，他人/Agent 左对齐带身份标识） -->
        <n-card id="sec-comments" title="评论" size="small" :bordered="true" :segmented="{ content: true }">
          <div ref="chatListRef" class="chat-list">
            <EmptyState type="comment" title="暂无评论" v-if="comments.length === 0" description="@成员 或 @Agent 可实时送达通知" compact />
            <div
              v-for="comment in comments"
              :key="comment.id"
              class="chat-row"
              :class="{ mine: isMyComment(comment) }"
            >
              <!-- 头像位：Agent 用 sparkle，人用首字母圆 -->
              <div v-if="!isMyComment(comment)" class="chat-avatar" :class="{ agent: comment.userType === 'ai' }">
                <AgentSparkleIcon v-if="comment.userType === 'ai'" />
                <template v-else>{{ avatarChar(comment) }}</template>
              </div>
              <div class="chat-main">
                <div class="chat-meta">
                  <span class="chat-name">{{ comment.realName || comment.username }}</span>
                  <n-tag size="tiny" :bordered="false" :type="comment.userType === 'ai' ? 'warning' : 'info'">
                    {{ comment.userType === 'ai' ? 'Agent' : '用户' }}
                  </n-tag>
                  <span class="chat-time">{{ fmtChatTime(comment.createdAt) }}</span>
                  <template v-if="isMyComment(comment)">
                    <n-button text type="primary" size="tiny" @click="startEditComment(comment)">编辑</n-button>
                    <n-button text type="error" size="tiny" @click="handleDeleteComment(comment)">删除</n-button>
                  </template>
                </div>
                <template v-if="editingCommentId === comment.id">
                  <n-input v-model:value="editCommentContent" type="textarea" :rows="2" />
                  <n-space :size="4" class="mt-1" justify="end">
                    <n-button size="tiny" @click="cancelEditComment">取消</n-button>
                    <n-button size="tiny" type="primary" :loading="savingComment" @click="saveEditComment">保存</n-button>
                  </n-space>
                </template>
                <div v-else class="chat-bubble comment-body" v-html="renderComment(comment.content)"></div>
              </div>
            </div>
          </div>
          <n-divider style="margin: 8px 0" />
          <n-space vertical :size="4">
            <n-mention
              v-model:value="newComment"
              type="textarea"
              placeholder="输入评论，@ 可提及成员或 Agent（实时送达通知）"
              :rows="2"
              :options="mentionOptions"
              :prefix="['@']"
              style="flex: 1"
              @keydown.enter.exact.prevent="handleAddComment"
            />
            <n-space justify="end">
              <n-button type="primary" @click="handleAddComment" :loading="submitting">发送</n-button>
            </n-space>
          </n-space>
        </n-card>
      </template>
    </n-spin>
    <input ref="attInputRef" type="file" multiple style="display: none" @change="onAttFiles" />

    </n-drawer-content>
  </n-drawer>

  <!-- 添加跨项目引用：项目下拉 + 该项目任务选择 -->
  <n-modal v-model:show="showRefModal" preset="dialog" title="添加跨项目引用" :show-icon="false">
    <n-space vertical :size="10" class="py-2">
      <n-select
        v-model:value="refForm.projectId"
        :options="refProjectOptions"
        placeholder="选择项目"
        size="small"
        @update:value="loadRefTasks"
      />
      <n-select
        v-model:value="refForm.taskId"
        :options="refTaskOptions"
        placeholder="选择任务"
        size="small"
        filterable
        :loading="refTasksLoading"
        :disabled="!refForm.projectId"
      />
    </n-space>
    <template #action>
      <n-space>
        <n-button size="small" @click="showRefModal = false">取消</n-button>
        <n-button size="small" type="primary" :disabled="!refForm.taskId" @click="addRef">添加</n-button>
      </n-space>
    </template>
  </n-modal>

  <!-- 重开任务：原因必填（评论留痕 + 通知原执行者） -->
  <n-modal v-model:show="showReopenModal" preset="dialog" title="重新打开任务" :show-icon="false">
    <n-alert type="warning" :bordered="false" class="mb-2">
      任务将回到「待认领」重新入池，完成时间与审核结果复位；原执行者会收到通知。
    </n-alert>
    <n-input
      v-model:value="reopenReason"
      type="textarea"
      placeholder="重开原因（必填）：说明实测发现的问题，将作为评论留痕"
      :rows="3"
    />
    <template #action>
      <n-space>
        <n-button size="small" @click="showReopenModal = false">取消</n-button>
        <n-button size="small" type="error" :disabled="!reopenReason.trim()" :loading="reopening" @click="handleReopen">
          确认重开
        </n-button>
      </n-space>
    </template>
  </n-modal>

  <!-- 上报阻塞：原因必填，作为通知与执行日志留痕 -->
  <n-modal v-model:show="showBlockModal" preset="dialog" title="上报阻塞" :show-icon="false">
    <n-input
      v-model:value="blockReason"
      type="textarea"
      placeholder="阻塞原因（必填）：说明卡在哪、需要什么信息/环境，将通知任务创建者"
      :rows="3"
    />
    <template #action>
      <n-space>
        <n-button size="small" @click="showBlockModal = false">取消</n-button>
        <n-button size="small" type="warning" :disabled="!blockReason.trim()" :loading="blocking" @click="handleBlock">
          确认阻塞
        </n-button>
      </n-space>
    </template>
  </n-modal>
</template>

<script lang="ts" setup>
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { useTaskComments } from '../composables/useTaskComments';
  import DOMPurify from 'dompurify';
  import { MdPreview } from 'md-editor-v3';
  import 'md-editor-v3/lib/preview.css';
  import { getTags, attachTag, detachTag, createTag } from '@/api/platform/index';
  import { getAttachments, uploadAttachment, downloadAttachment, deleteAttachment } from '@/api/attachment/index';
  import { getLinkedDocs } from '@/api/docs/index';
  import type { DocsSearchItem } from '@/api/docs/index';
  import { PaperClipOutlined, FileTextOutlined } from '@vicons/antd';
  import { ref, reactive, computed, watch, onMounted, onUnmounted, nextTick } from 'vue';
  import { useRouter } from 'vue-router';
  import { useMessage, useDialog } from 'naive-ui';
  import {
    getTask,
    getProjects,
    getComments,
    createComment,
    reviewTask,
    blockTask,
    unblockTask,
    reopenTask,
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
  import { AgentSparkleIcon } from '@/components/Icons/AgentSparkle';


  // 抽屉内的写操作（关闭/审核/改期）会改变宿主列表展示的数据，
  // 成功后通知宿主重拉——defineExpose 只有 openModal，宿主无从感知
  const emit = defineEmits<{ (e: 'updated'): void }>();

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

  // tag 模式回车新建：naive 以字符串 label 触发 create——先建标签拿到 id，
  // 再挂到任务并返回选项供 select 立即选中（直接把字符串当 id 传 attachTag 必失败）
  async function onTagCreate(label: string) {
    if (!task.value) return { label, value: label };
    try {
      const res = await createTag({ name: label });
      const t = { id: res.id, name: label };
      allTags.value.push(t);
      await attachTag(t.id, { entityType: 'task', entityId: task.value.id });
      task.value.tags = [...(task.value.tags || []), label];
      return { label, value: t.id };
    } catch {
      message.error('新建标签失败');
      return { label, value: label };
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

  // @提及候选：项目成员（human 与 Agent 都可被 @，Agent 头像标识区分）；
  // 插入值为 username（与后端 notifyMentions 的 @用户名 精确匹配口径
  // 一致——显示名仅在候选列表里辅助识别）
  const mentionOptions = ref<Array<{ label: string; value: string }>>([]);
  async function loadMentionOptions(projectId: number) {
    try {
      const res = await getMembers(projectId);
      // human 与 Agent 都可被 @：Agent 头像标识区分，选中即发通知
      mentionOptions.value = (res?.list || [])
        .filter((m) => !m.userType || m.userType === 'human' || m.userType === 'ai')
        .map((m) => ({
          label: m.userType === 'ai'
            ? `✦ ${m.realName || m.username}（${m.username}·Agent）`
            : `${m.realName || m.username}（${m.username}）`,
          value: m.username,
        }));
    } catch {
      mentionOptions.value = [];
    }
  }

  const router = useRouter();

  // 宽屏大面板（AI 产出 markdown/长评论阅读），窄屏留 8% 呼吸
  const drawerWidth = typeof window !== 'undefined' ? Math.min(880, window.innerWidth * 0.92) : 720;

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
    // 全量重置：切任务时旧任务内容会在 spinner 下闪现，
    // getTask 失败时更会被永久当作当前任务展示
    task.value = null;
    comments.value = [];
    attachments.value = [];
    cancelEditComment();
    newComment.value = '';
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

  // 子任务跳转：关闭自身，广播给宿主（tasks 列表页监听 open-task 后调 openModal）
  function openTask(id: number) {
    visible.value = false;
    window.dispatchEvent(new CustomEvent('bc-open-task', { detail: id }));
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

  // ==================== IM 对话流工具 ====================
  const chatListRef = ref<HTMLElement | null>(null);

  function avatarChar(c: CommentItem): string {
    return (c.realName || c.username || '?').trim().charAt(0).toUpperCase();
  }

  // IM 时间：今天只显示时分，更早显示月-日 时分
  function fmtChatTime(ts: string): string {
    if (!ts) return '';
    const d = ts.slice(0, 10);
    const hm = ts.slice(11, 16);
    const today = new Date();
    const t = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`;
    return d === t ? hm : `${d.slice(5)} ${hm}`;
  }

  async function scrollChatBottom() {
    await nextTick();
    if (chatListRef.value) chatListRef.value.scrollTop = chatListRef.value.scrollHeight;
  }

  // 评论区加载与 SSE 实时刷新已抽至 useTaskComments composable（含已知
  // 脆弱点注释：通知标题精确匹配，平台换结构化字段时只改 composable 一处）
  const { comments, refreshComments } = useTaskComments(
    () => (visible.value && task.value ? task.value.id : 0),
    () => scrollChatBottom()
  );

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

  async function handleAddComment(e?: KeyboardEvent) {
    // 输入法组合态的 Enter（确认候选词）不是发送意图；keyCode 229 兼容
    // Safari 等不置 isComposing 的旧实现
    if (e && (e.isComposing || e.keyCode === 229)) return;
    if (!newComment.value.trim()) {
      message.warning('请输入评论内容');
      return;
    }
    if (!task.value) return;
    submitting.value = true;
    try {
      await createComment(task.value.id, { content: newComment.value });
      newComment.value = '';
      // IM 式上屏：静默刷新后滚到底（新评论即见）
      await refreshComments(true);
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
      emit('updated');
    } catch {
      task.value.dueDate = old;
      message.error('更新截止日期失败');
    }
  }

  // 快速跳转：锚点随卡片存在性动态显示；scroll-margin-top 防 sticky 条遮挡
  const activeSec = ref('');
  const navSections = computed(() => {
    const secs = [
      { id: 'sec-desc', label: '描述' },
      { id: 'sec-refs', label: '引用' },
      { id: 'sec-checklist', label: '清单' },
      { id: 'sec-logs', label: '日志' },
    ];
    if (task.value?.artifacts) secs.push({ id: 'sec-artifacts', label: '产出' });
    if (task.value?.requiresHumanReview === 1 && task.value?.humanReviewStatus === 'pending') {
      secs.push({ id: 'sec-review', label: '审核' });
    }
    secs.push({ id: 'sec-comments', label: '评论' });
    return secs;
  });
  function scrollToSection(id: string) {
    activeSec.value = id;
    document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' });
  }

  // 步骤清单：JSON [{text,done}]；变更走 updateTask（updated_at 随之刷新，
  // 打勾即租约心跳）。本地乐观更新，失败静默回读
  interface CheckItem { text: string; done: boolean }
  const checklist = computed<CheckItem[]>(() => {
    try {
      return JSON.parse(task.value?.checklist || '[]');
    } catch {
      return [];
    }
  });
  const newCheckText = ref('');

  async function saveChecklist(items: CheckItem[]) {
    if (!task.value) return;
    try {
      await updateTask(task.value.id, { checklist: JSON.stringify(items) });
      // 回写本地（computed 依赖 task.value.checklist，不回写则 UI 不动）
      task.value = { ...task.value, checklist: JSON.stringify(items) };
    } catch {
      message.error('保存清单失败');
      const res = await getTask(task.value.id);
      task.value = res || null;
    }
  }
  function toggleCheck(i: number, done: boolean) {
    const items = checklist.value.map((c, idx) => (idx === i ? { ...c, done } : c));
    saveChecklist(items);
  }
  function addCheck() {
    const text = newCheckText.value.trim();
    if (!text) return;
    newCheckText.value = '';
    saveChecklist([...checklist.value, { text, done: false }]);
  }
  function removeCheck(i: number) {
    saveChecklist(checklist.value.filter((_, idx) => idx !== i));
  }

  // 跨项目引用：解析/增删/跳转；写走 updateTask（后端校验权限与结构，
  // 新增项通知被引任务创建者）
  interface RelatedRef { type: string; projectId: number; id: number; title: string }
  const relatedRefs = computed<RelatedRef[]>(() => {
    try {
      return JSON.parse(task.value?.relatedRefs || '[]');
    } catch {
      return [];
    }
  });
  const showRefModal = ref(false);
  const refForm = reactive({ projectId: null as number | null, taskId: null as number | null });
  const refProjectOptions = ref<Array<{ label: string; value: number }>>([]);
  const refTaskOptions = ref<Array<{ label: string; value: number }>>([]);
  const refTasksLoading = ref(false);

  async function loadRefProjects() {
    try {
      const res = await getProjects({ size: 100 });
      refProjectOptions.value = (res?.list || [])
        .filter((p: any) => p.id !== task.value?.projectId)
        .map((p: any) => ({ label: p.name, value: p.id }));
    } catch { /* ignore */ }
  }
  async function loadRefTasks(pid: number) {
    refForm.taskId = null;
    refTaskOptions.value = [];
    refTasksLoading.value = true;
    try {
      const res = await getTasks(pid, { size: 100 });
      refTaskOptions.value = (res?.list || []).map((t: TaskItem) => ({ label: t.title, value: t.id }));
    } catch {
      message.error('加载该项目任务失败（可能无访问权限）');
    } finally {
      refTasksLoading.value = false;
    }
  }
  function refProjectLabel(pid: number): string {
    const hit = refProjectOptions.value.find((p) => p.value === pid);
    return hit?.label || `项目 ${pid}`;
  }
  async function saveRefs(list: RelatedRef[]) {
    if (!task.value) return;
    try {
      await updateTask(task.value.id, { relatedRefs: JSON.stringify(list) });
      task.value = { ...task.value, relatedRefs: JSON.stringify(list) };
    } catch (e: any) {
      message.error(e?.message || '保存引用失败');
    }
  }
  async function addRef() {
    if (!refForm.projectId || !refForm.taskId) return;
    const opt = refTaskOptions.value.find((t) => t.value === refForm.taskId);
    const next = [...relatedRefs.value, {
      type: 'task', projectId: refForm.projectId, id: refForm.taskId, title: opt?.label || String(refForm.taskId),
    }];
    await saveRefs(next);
    if (task.value && JSON.parse(task.value.relatedRefs || '[]').length === next.length) {
      message.success('引用已添加，对方创建者将收到通知');
      showRefModal.value = false;
    }
  }
  function removeRef(i: number) {
    saveRefs(relatedRefs.value.filter((_, idx) => idx !== i));
  }
  function openRef(r: RelatedRef) {
    if (r.type === 'task') {
      router.push(`/project/${r.projectId}/tasks?task=${r.id}`);
    }
  }
  watch(showRefModal, (v) => { if (v) loadRefProjects(); });

  // 阻塞流转：原因必填；成功后刷新任务并通知宿主列表
  const showBlockModal = ref(false);
  const blockReason = ref('');
  const blocking = ref(false);

  async function handleBlock() {
    if (!task.value || !blockReason.value.trim()) return;
    blocking.value = true;
    try {
      await blockTask(task.value.id, blockReason.value.trim());
      message.success('已上报阻塞，任务创建者会收到通知');
      showBlockModal.value = false;
      blockReason.value = '';
      const res = await getTask(task.value.id);
      task.value = res || null;
      emit('updated');
    } catch (e: any) {
      message.error(e?.message || '上报阻塞失败');
    } finally {
      blocking.value = false;
    }
  }

  // 终态重开：原因必填；回 open 入池，弹窗内就地刷新任务状态
  const showReopenModal = ref(false);
  const reopenReason = ref('');
  const reopening = ref(false);

  async function handleReopen() {
    if (!task.value || !reopenReason.value.trim()) return;
    reopening.value = true;
    try {
      await reopenTask(task.value.id, reopenReason.value.trim());
      message.success('任务已重开，回到待认领');
      showReopenModal.value = false;
      reopenReason.value = '';
      const res = await getTask(task.value.id);
      task.value = res || null;
      const cr = await getComments(task.value.id);
      comments.value = cr?.list || [];
      emit('updated');
    } catch (e: any) {
      message.error(e?.message || '重开失败');
    } finally {
      reopening.value = false;
    }
  }

  async function handleUnblock() {
    if (!task.value) return;
    try {
      await unblockTask(task.value.id);
      message.success('已解除阻塞，任务恢复进行中');
      const res = await getTask(task.value.id);
      task.value = res || null;
      emit('updated');
    } catch (e: any) {
      message.error(e?.message || '解除阻塞失败');
    }
  }

  // 关闭任务（done→closed 清账；重开=updateTask status open）
  async function handleClose() {
    if (!task.value) return;
    try {
      await updateTask(task.value.id, { status: 'closed' });
      message.success('任务已关闭');
      visible.value = false;
      emit('updated');
    } catch {
      message.error('关闭失败');
    }
  }

  async function handleReview(status: 'approved' | 'rejected') {
    if (!task.value) return;
    try {
      await reviewTask(task.value.id, { status });
      message.success(status === 'approved' ? '已通过' : '已驳回');
      const taskRes = await getTask(task.value.id);
      task.value = taskRes || null;
      emit('updated');
    } catch (e) {
      message.error('审核操作失败');
    }
  }

  defineExpose({ openModal });
</script>

<style lang="less" scoped>
  // 吸顶锚点条：胶囊分段风格（轻量、无边框按钮感）
  .section-nav {
    position: sticky;
    top: -16px; // 抵消抽屉内容区 padding
    z-index: 10;
    display: flex;
    align-items: center;
    gap: 4px;
    flex-wrap: wrap;
    padding: 8px 0;
    margin-bottom: 4px;
    // 显式白底：sticky 需不透明遮底；勿用 --n-color（项目主题 primaryColor
    // 为绿色，该变量在此上下文解析为主题绿——实测踩坑）。暗色主题适配
    // 与 notifications.vue 一并处理（见 status 文档 F 项）
    background: #fff;
    border-bottom: 1px solid var(--border-color, #efeff5);

    .nav-chip {
      border: none;
      background: transparent;
      color: var(--n-text-color-3, #999);
      font-size: 12px;
      line-height: 1;
      padding: 5px 12px;
      border-radius: 999px;
      cursor: pointer;
      transition: color 0.2s, background-color 0.2s;

      &:hover {
        color: var(--n-primary-color, #2080f0);
        background: var(--n-color-hover, rgba(32, 128, 240, 0.08));
      }

      &.active {
        color: #fff;
        background: var(--n-primary-color, #2080f0);
      }
    }
  }
  // 各节卡片滚动定位时避开 sticky 条高度
  ::v-deep([id^='sec-']) {
    scroll-margin-top: 52px;
  }

  .ref-row {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 5px 0;
    cursor: pointer;

    .ref-title {
      flex: 1;
      min-width: 0;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .ic {
      stroke: var(--primary);
      flex: none;
    }
  }
  .checklist-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 2px 0;
  }
  .checklist-done {
    text-decoration: line-through;
    color: var(--text-color-3, #999);
  }

  .linked-doc {
    padding: 4px 8px;
    margin: 0 -8px;
    border-radius: 4px;
    cursor: pointer;

    &:hover {
      background: var(--hover-bg, #f5f5f5);
    }
  }

  // ==================== IM 对话流 ====================
  .chat-list {
    max-height: 320px;
    overflow-y: auto;
    padding: 4px 2px;
  }

  .chat-row {
    display: flex;
    gap: 8px;
    margin-bottom: 12px;

    &.mine {
      flex-direction: row-reverse;

      .chat-main {
        align-items: flex-end;
      }

      .chat-bubble {
        background: #36ad6a;
        color: #fff;
        border-radius: 12px 2px 12px 12px;

        :deep(.mention) {
          color: #fff;
          background: rgb(255 255 255 / 20%);
        }
      }
    }
  }

  .chat-avatar {
    flex: none;
    width: 28px;
    height: 28px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 13px;
    font-weight: 600;
    color: #fff;
    background: #5a9cf8;
    overflow: hidden;

    &.agent {
      background: #f0a020;
      font-size: 16px;
    }
  }

  .chat-main {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    min-width: 0;
    max-width: 86%;
  }

  .chat-meta {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-bottom: 2px;
    font-size: 12px;
  }

  .chat-name {
    font-weight: 500;
    color: #333;
  }

  .chat-time {
    color: #999;
  }

  .chat-bubble {
    background: #f2f3f5;
    color: #333;
    border-radius: 2px 12px 12px 12px;
    padding: 6px 10px;
    font-size: 13px;
    line-height: 1.55;
    word-break: break-word;
    white-space: pre-wrap;
  }

  :deep(.mention) {
    color: #18a058;
    font-weight: 500;
    background: rgb(24 160 88 / 8%);
    border-radius: 3px;
    padding: 0 3px;
  }
</style>

<style lang="less">
// 非 scoped：modal teleport 到 body 后脱离组件 DOM 树，scoped/:deep 选不中根卡。
// card 型弹窗在 style max-height 约束下内容区内部滚动——默认 overflow
// visible 会让长内容穿透弹窗边界显示在下方（无 max-height 的弹窗不受影响）
.n-modal.n-card > .n-card__content {
  overflow-y: auto;
  min-height: 0;
}

// 抽屉 body 滚动（n-drawer-content 自带，这里只确保评论区输入区不随内容滚动走丢）
:global(.task-detail-drawer .n-drawer-body-content-wrapper) {
  padding-bottom: 8px;
}

  .subtask-row {
    display: flex;
    gap: 8px;
    align-items: center;
    padding: 5px 8px;
    border-radius: 6px;
    cursor: pointer;

    &:hover {
      background: var(--hover-bg);
    }

    .subtask-title {
      flex: 1;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      font-size: 13px;
    }

    .subtask-assignee {
      flex: none;
      color: var(--text-3, #8b949e);
      font-size: 12px;
    }
  }
</style>
