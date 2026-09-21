<template>
  <div>
    <n-card :bordered="false" class="proCard">
      <template #header>
        讨论区
        <span class="text-xs text-gray-400 ml-2 font-normal">想法先聊再建：讨论成熟后可一键转任务，背景不丢失</span>
      </template>
      <template #header-extra>
        <n-space :size="8">
          <n-select
            v-model:value="statusFilter"
            :options="statusOptions"
            size="small"
            style="width: 110px"
            @update:value="reload"
          />
          <n-button type="primary" size="small" @click="openCreate" data-test-id="discussions.create-btn">
            <template #icon>
              <n-icon><PlusOutlined /></n-icon>
            </template>
            发起讨论
          </n-button>
        </n-space>
      </template>

      <n-data-table
        :columns="columns"
        :data="list"
        :loading="loading"
        :row-props="rowProps"
        :max-height="tableMaxHeight"
        size="small"
      >
        <template #empty>
          <EmptyState
            type="data"
            title="暂无讨论"
            description="有了想法先发起讨论，团队成员与 AI 都可以参与"
          />
        </template>
      </n-data-table>

      <div class="mt-4 flex justify-end" v-if="total > pageSize">
        <n-pagination v-model:page="page" :page-size="pageSize" :item-count="total" @update:page="loadData" />
      </div>
    </n-card>

    <!-- 详情抽屉：大容器承载正文 + 回复流，回复框常驻底部 -->
    <n-drawer v-model:show="showDetail" :width="760" placement="right" :auto-focus="false">
      <n-drawer-content closable>
        <template #header>
          <div class="ds-drawer-head" v-if="detail">
            <div class="ds-head-line">
              <h3 class="ds-title" data-test-id="discussions.detail-title">{{ detail.title }}</h3>
              <n-tag v-if="detail.status === 'converted'" size="small" type="success" :bordered="false">
                已转任务
                <router-link
                  class="ml-1"
                  :to="`/project/${detail.projectId}/tasks`"
                  data-test-id="discussions.task-link"
                >
                  #{{ detail.convertedTaskId }}
                </router-link>
              </n-tag>
              <n-tag v-else-if="detail.status === 'archived'" size="small" :bordered="false">已归档</n-tag>
            </div>
            <div class="ds-head-meta">
              <span>{{ detail.authorName || '-' }}</span>
              <n-tag v-if="detail.authorType === 'ai'" size="tiny" type="info" :bordered="false">AI</n-tag>
              <span>· 发起于 {{ detail.createdAt }}</span>
              <n-space size="small" :wrap="false" class="ds-head-actions">
                <n-button
                  v-if="detail.status !== 'converted'"
                  text
                  type="warning"
                  size="small"
                  @click="openConvert"
                  data-test-id="discussions.convert-btn"
                >
                  转为任务
                </n-button>
                <n-button text size="small" @click="toggleArchive" data-test-id="discussions.archive-btn">
                  {{ detail.status === 'archived' ? '恢复' : '归档' }}
                </n-button>
                <n-button text type="error" size="small" @click="handleDelete" data-test-id="discussions.del-btn">
                  删除
                </n-button>
              </n-space>
            </div>
          </div>
        </template>

        <n-spin :show="detailLoading">
          <div class="ds-detail" v-if="detail">
            <div class="ds-body">
              <MdPreview :model-value="detail.body || '（无正文）'" :sanitize="safeHtml" :theme="isDark ? 'dark' : 'light'" />
            </div>

            <n-divider style="margin: 14px 0 8px" />
            <div class="text-sm font-medium mb-2">回复（{{ detail.replies.length }}）</div>
            <div class="ds-replies">
              <div v-for="r in detail.replies" :key="r.id" class="ds-reply" :data-test-id="`discussions.reply-${r.id}`">
                <div class="text-xs text-gray-400 mb-1">
                  {{ r.userName || '-' }}
                  <n-tag v-if="r.userType === 'ai'" size="tiny" type="info" :bordered="false">AI</n-tag>
                  · {{ r.createdAt }}
                </div>
                <div class="ds-reply-content">{{ r.content }}</div>
              </div>
              <div v-if="detail.replies.length === 0" class="text-xs text-gray-400">还没有回复，说说你的想法</div>
            </div>
          </div>
        </n-spin>

        <template #footer>
          <div v-if="detail && detail.status !== 'archived'">
            <n-input
              v-model:value="replyDraft"
              type="textarea"
              placeholder="回复讨论（支持多行文本；agent 参与@提及见 v2）"
              :rows="3"
              data-test-id="discussions.reply-input"
            />
            <div class="mt-2 text-right">
              <n-button
                type="primary"
                size="small"
                :loading="replying"
                @click="submitReply"
                data-test-id="discussions.reply-btn"
              >
                回复
              </n-button>
            </div>
          </div>
          <div v-else-if="detail" class="text-xs text-gray-400">已归档的讨论不再接受回复</div>
        </template>
      </n-drawer-content>
    </n-drawer>

    <!-- 发起讨论：正文 md-editor-v3 -->
    <n-modal
      v-model:show="showCreate"
      preset="dialog"
      title="发起讨论"
      positive-text="发布"
      negative-text="取消"
      @positive-click="submitCreate"
      style="width: 760px"
    >
      <n-form label-placement="top" class="py-2">
        <n-form-item label="标题" required>
          <n-input v-model:value="createForm.title" placeholder="一句话说清这个想法/议题" data-test-id="discussions.title-input" />
        </n-form-item>
        <n-form-item label="背景（markdown）">
          <MdEditor
            v-model="createForm.body"
            class="ds-create-editor"
            placeholder="背景、动机、初步想法……给参与的人（和 AI）足够的上下文"
            :theme="isDark ? 'dark' : 'light'"
            :toolbarsExclude="['github', 'save', 'htmlPreview', 'catalog']"
            :footers="[]"
            :sanitize="safeHtml"
            data-test-id="discussions.body-input"
          />
        </n-form-item>
      </n-form>
    </n-modal>

    <!-- 转任务 -->
    <n-modal
      v-model:show="showConvert"
      preset="dialog"
      title="转为任务"
      positive-text="创建任务"
      negative-text="取消"
      @positive-click="submitConvert"
      style="width: 520px"
    >
      <div class="py-2 text-xs text-gray-400 mb-2">
        讨论正文与来源链接会写入任务描述，讨论标记为「已转」并回链任务。
      </div>
      <n-form label-placement="left" label-width="72">
        <n-form-item label="任务标题">
          <n-input v-model:value="convertForm.title" data-test-id="discussions.convert-title" />
        </n-form-item>
        <n-form-item label="类型">
          <n-select v-model:value="convertForm.type" :options="taskTypeOptions" />
        </n-form-item>
      </n-form>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, h, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { useMessage, useDialog } from 'naive-ui';
  import { NTag, NSpace, NIcon } from 'naive-ui';
  import type { DataTableColumns } from 'naive-ui';
  import { PlusOutlined, MessageOutlined } from '@vicons/antd';
  import { MdPreview, MdEditor } from 'md-editor-v3';
  import 'md-editor-v3/lib/style.css';
  import DOMPurify from 'dompurify';
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { useDesignSetting } from '@/store/modules/designSetting';
  import {
    getDiscussions,
    createDiscussion,
    getDiscussionDetail,
    deleteDiscussion,
    archiveDiscussion,
    replyDiscussion,
    convertDiscussion,
  } from '@/api/project/discussion';
  import type { DiscussionItem, DiscussionDetail } from '@/api/project/discussion';

  const message = useMessage();
  const dialog = useDialog();
  const route = useRoute();
  const { getDarkTheme } = useDesignSetting();
  const isDark = computed(() => getDarkTheme.value === true);
  const projectId = Number(route.params.projectId);

  const statusOptions = [
    { label: '全部状态', value: '' },
    { label: '开放', value: 'open' },
    { label: '已转任务', value: 'converted' },
    { label: '已归档', value: 'archived' },
  ];
  const taskTypeOptions = [
    { label: '功能（feature）', value: 'feature' },
    { label: '缺陷（bug）', value: 'bug' },
    { label: '测试（test）', value: 'test' },
    { label: '文档（docs）', value: 'docs' },
    { label: '杂务（chore）', value: 'chore' },
  ];

  function relTime(dt: string): string {
    if (!dt) return '';
    const t = new Date(dt.replace(' ', 'T')).getTime();
    if (Number.isNaN(t)) return dt;
    const diff = Date.now() - t;
    if (diff < 60_000) return '刚刚';
    if (diff < 3_600_000) return `${Math.floor(diff / 60_000)} 分钟前`;
    if (diff < 86_400_000) return `${Math.floor(diff / 3_600_000)} 小时前`;
    if (diff < 30 * 86_400_000) return `${Math.floor(diff / 86_400_000)} 天前`;
    return dt;
  }
  function safeHtml(md: string): string {
    return DOMPurify.sanitize(md);
  }

  // ---------- 列表（表格） ----------
  const loading = ref(false);
  const list = ref<DiscussionItem[]>([]);
  const total = ref(0);
  const page = ref(1);
  const pageSize = 20;
  const statusFilter = ref('');
  const tableMaxHeight = computed(() => Math.max(320, window.innerHeight - 300));

  async function loadData() {
    loading.value = true;
    try {
      const res = await getDiscussions(projectId, {
        pageNum: page.value,
        pageSize,
        status: statusFilter.value || undefined,
      });
      list.value = res?.list || [];
      total.value = res?.total || 0;
    } finally {
      loading.value = false;
    }
  }
  function reload() {
    page.value = 1;
    loadData();
  }

  function statusTag(s: string, taskId: number) {
    if (s === 'converted') {
      return h(NTag, { size: 'small', type: 'success', bordered: false }, () => `已转 #${taskId}`);
    }
    if (s === 'archived') {
      return h(NTag, { size: 'small', bordered: false }, () => '已归档');
    }
    return h(NTag, { size: 'small', type: 'info', bordered: false }, () => '开放');
  }

  const columns: DataTableColumns<DiscussionItem> = [
    {
      title: '标题',
      key: 'title',
      render: (row) =>
        h('span', { class: 'ds-cell-title' }, [
          h('span', { class: `ds-status-dot st-${row.status}` }),
          row.title,
        ]),
    },
    { title: '状态', key: 'status', width: 110, render: (row) => statusTag(row.status, row.convertedTaskId) },
    {
      title: '回复',
      key: 'replyCount',
      width: 70,
      render: (row) =>
        h('span', { class: 'ds-reply-count' }, [
          h(NIcon, { size: 13 }, () => h(MessageOutlined)),
          `${row.replyCount}`,
        ]),
    },
    {
      title: '发起人',
      key: 'authorName',
      width: 140,
      render: (row) =>
        h(NSpace, { size: 4, align: 'center', wrap: false }, () => [
          row.authorName || '-',
          row.authorType === 'ai' ? h(NTag, { size: 'tiny', type: 'info', bordered: false }, () => 'AI') : null,
        ]),
    },
    { title: '发起时间', key: 'createdAt', width: 150, render: (row) => (row.createdAt || '').slice(0, 16) },
    { title: '最近活动', key: 'updatedAt', width: 110, render: (row) => relTime(row.updatedAt) },
  ];

  // 行点击开抽屉；锚点挂 tr（row-props 直落 <tr>）
  function rowProps(row: DiscussionItem) {
    return {
      style: 'cursor: pointer;',
      'data-test-id': `discussions.row-${row.id}`,
      onClick: () => openDetail(row.id),
    };
  }

  // ---------- 详情（抽屉） ----------
  const showDetail = ref(false);
  const detail = ref<DiscussionDetail | null>(null);
  const detailLoading = ref(false);
  const replyDraft = ref('');
  const replying = ref(false);

  async function openDetail(id: number) {
    showDetail.value = true;
    detailLoading.value = true;
    try {
      detail.value = await getDiscussionDetail(id);
    } finally {
      detailLoading.value = false;
    }
  }

  // ---------- 发起 ----------
  const showCreate = ref(false);
  const createForm = ref({ title: '', body: '' });
  function openCreate() {
    createForm.value = { title: '', body: '' };
    showCreate.value = true;
  }
  async function submitCreate() {
    if (!createForm.value.title.trim()) {
      message.error('请填写标题');
      return false;
    }
    try {
      await createDiscussion(projectId, { ...createForm.value });
      message.success('讨论已发起');
      showCreate.value = false;
      await loadData();
    } catch (e: any) {
      message.error(e?.message || '发起失败');
      return false;
    }
  }

  // ---------- 回复 ----------
  async function submitReply() {
    if (!detail.value || !replyDraft.value.trim()) return;
    replying.value = true;
    try {
      await replyDiscussion(detail.value.id, replyDraft.value);
      replyDraft.value = '';
      await openDetail(detail.value.id);
      loadData();
    } catch (e: any) {
      message.error(e?.message || '回复失败');
    } finally {
      replying.value = false;
    }
  }

  // ---------- 归档 / 删除 / 转任务 ----------
  async function toggleArchive() {
    if (!detail.value) return;
    try {
      const res = await archiveDiscussion(detail.value.id);
      message.success(res?.status === 'archived' ? '已归档' : '已恢复');
      await openDetail(detail.value.id);
      loadData();
    } catch (e: any) {
      message.error(e?.message || '操作失败');
    }
  }

  function handleDelete() {
    if (!detail.value) return;
    const d = detail.value;
    dialog.warning({
      title: '删除讨论',
      content: `确定删除讨论「${d.title}」及其全部回复？不可恢复。`,
      positiveText: '删除',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await deleteDiscussion(d.id);
          message.success('已删除');
          showDetail.value = false;
          detail.value = null;
          loadData();
        } catch (e: any) {
          message.error(e?.message || '删除失败');
        }
      },
    });
  }

  const showConvert = ref(false);
  const convertForm = ref({ title: '', type: 'feature' });
  function openConvert() {
    if (!detail.value) return;
    convertForm.value = { title: detail.value.title, type: 'feature' };
    showConvert.value = true;
  }
  async function submitConvert() {
    if (!detail.value) return false;
    try {
      const res = await convertDiscussion(detail.value.id, { ...convertForm.value });
      message.success(`已转为任务 #${res?.taskId}`);
      showConvert.value = false;
      await openDetail(detail.value.id);
      loadData();
    } catch (e: any) {
      message.error(e?.message || '转化失败');
      return false;
    }
  }

  onMounted(() => {
    loadData();
  });
</script>

<style scoped>
  /* 表格单元格 */
  .ds-cell-title {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-weight: 500;
    color: var(--text-1, #24292f);
    min-width: 0;
  }
  .ds-status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
    background: #18a058;
  }
  .ds-status-dot.st-converted {
    background: #2080f0;
  }
  .ds-status-dot.st-archived {
    background: var(--text-3, #8b949e);
  }
  .ds-reply-count {
    display: inline-flex;
    align-items: center;
    font-size: 12px;
    color: var(--text-3, #8b949e);
    font-variant-numeric: tabular-nums;
  }

  /* 抽屉头部 */
  .ds-drawer-head {
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-width: 0;
    padding-right: 32px;
  }
  .ds-head-line {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
  }
  .ds-title {
    margin: 0;
    font-size: 17px;
    color: var(--text-1, #24292f);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .ds-head-meta {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    color: var(--text-3, #8b949e);
  }
  .ds-head-actions {
    margin-left: auto;
  }

  /* 抽屉正文与回复 */
  .ds-detail {
    min-height: 200px;
  }
  .ds-body {
    border: 1px solid var(--line, #e9e9e7);
    border-radius: 8px;
    padding: 4px 14px;
    background: var(--panel-bg, #fff);
  }
  .ds-replies {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .ds-reply {
    border: 1px solid var(--line, #e9e9e7);
    border-radius: 8px;
    padding: 8px 12px;
    background: var(--panel-bg, #fff);
  }
  .ds-reply-content {
    font-size: 13px;
    color: var(--text-1, #24292f);
    white-space: pre-wrap;
    word-break: break-word;
  }

  /* 发起弹窗里的编辑器高度 */
  .ds-create-editor {
    height: 280px;
  }
</style>
