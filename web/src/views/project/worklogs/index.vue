<template>
  <div>
    <n-card :bordered="false" class="proCard">
      <template #header>
        工作日志
        <span class="text-xs text-gray-400 ml-2 font-normal">项目演化叙事——做了什么、结果如何</span>
      </template>
      <template #header-extra>
        <n-space :size="8">
          <n-button size="small" @click="openGen" data-test-id="project-worklog.gen-btn">
            <template #icon>
              <n-icon><RobotOutlined /></n-icon>
            </template>
            从任务生成草稿
          </n-button>
          <n-button size="small" type="primary" @click="openCreate" data-test-id="project-worklog.add-btn">
            <template #icon>
              <n-icon><PlusOutlined /></n-icon>
            </template>
            写日志
          </n-button>
        </n-space>
      </template>

      <n-spin :show="loading">
        <EmptyState
          v-if="!loading && groups.length === 0"
          type="generic"
          title="还没有工作日志"
          description="记录做了什么、结果如何；或从已完成的任务一键生成草稿"
        />
        <template v-else>
          <div v-for="g in groups" :key="g.date" class="wl-group" :data-test-id="`project-worklog.group-${g.date}`">
            <div class="wl-group-head">
              <span class="wl-date">{{ g.label }}</span>
              <span class="wl-count">{{ g.items.length }} 条</span>
            </div>
            <div class="wl-list">
              <div
                v-for="item in g.items"
                :key="item.id"
                class="wl-item"
                :data-test-id="`project-worklog.item-${item.id}`"
              >
                <div class="wl-item-head">
                  <span class="wl-time">{{ (item.createdAt || '').slice(11, 16) }}</span>
                  <span class="wl-author">{{ item.authorName || `#${item.authorId}` }}</span>
                  <n-tag v-if="item.authorType === 'ai'" size="tiny" type="info" :bordered="false">AI</n-tag>
                  <n-tag v-if="item.source === 'tasks'" size="tiny" :bordered="false" class="wl-src">任务草稿</n-tag>
                  <span class="wl-actions">
                    <n-button text size="tiny" type="primary" :data-test-id="`project-worklog.edit-${item.id}`" @click="openEdit(item)">编辑</n-button>
                    <n-button text size="tiny" type="error" :data-test-id="`project-worklog.del-${item.id}`" @click="confirmDelete(item)">删除</n-button>
                  </span>
                </div>
                <div class="wl-content">
                  <MdPreview :model-value="item.content" :sanitize="safeHtml" />
                </div>
              </div>
            </div>
          </div>
          <div class="mt-4 flex justify-center" v-if="hasMore">
            <n-button size="small" quaternary type="primary" :loading="loadingMore" @click="loadMore" data-test-id="project-worklog.load-more">
              加载更多
            </n-button>
          </div>
        </template>
      </n-spin>
    </n-card>

    <!-- 记录 / 编辑 -->
    <n-modal
      v-model:show="showEditor"
      preset="card"
      :title="editing ? '编辑工作日志' : '记录工作日志'"
      class="wl-editor-modal"
      :mask-closable="false"
      data-test-id="project-worklog.editor-modal"
    >
      <n-input
        v-model:value="editorContent"
        type="textarea"
        :rows="14"
        placeholder="做了什么、结果如何、验收情况……（支持 markdown）"
        data-test-id="project-worklog.content-input"
      />
      <div v-if="editorSource === 'tasks'" class="text-xs text-gray-400 mt-2">
        草稿由已完成任务/发布汇总生成，润色后发布
      </div>
      <template #footer>
        <n-space justify="end">
          <n-button size="small" @click="showEditor = false">取消</n-button>
          <n-button
            size="small"
            type="primary"
            :disabled="!editorContent.trim()"
            :loading="saving"
            data-test-id="project-worklog.submit-btn"
            @click="save"
          >
            {{ editing ? '保存' : '发布' }}
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 从任务生成草稿：选时间范围 -->
    <n-modal
      v-model:show="showGen"
      preset="dialog"
      title="从任务生成草稿"
      :show-icon="false"
      style="width: 480px"
      data-test-id="project-worklog.gen-modal"
    >
      <n-space vertical :size="10" class="py-2">
        <span class="text-sm">汇总时间段内已完成的任务（标题+执行者+产物）与发布，生成草稿进编辑器润色</span>
        <n-date-picker
          v-model:value="genRange"
          type="daterange"
          clearable
          data-test-id="project-worklog.gen-range"
        />
      </n-space>
      <template #action>
        <n-space>
          <n-button size="small" @click="showGen = false">取消</n-button>
          <n-button
            size="small"
            type="primary"
            :disabled="!genRange || genRange.length !== 2"
            :loading="generating"
            data-test-id="project-worklog.gen-confirm"
            @click="generate"
          >
            生成草稿
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { useMessage, useDialog } from 'naive-ui';
  import { PlusOutlined, RobotOutlined } from '@vicons/antd';
  import { MdPreview } from 'md-editor-v3';
  import 'md-editor-v3/lib/style.css';
  import DOMPurify from 'dompurify';
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import {
    getWorklogs,
    createWorklog,
    updateWorklog,
    deleteWorklog,
    getWorklogDraft,
  } from '@/api/project/worklog';
  import type { WorklogItem } from '@/api/project/worklog';

  const message = useMessage();
  const dialog = useDialog();
  const route = useRoute();
  const projectId = Number(route.params.projectId);

  const safeHtml = (html: string) => DOMPurify.sanitize(html);

  const loading = ref(false);
  const loadingMore = ref(false);
  const list = ref<WorklogItem[]>([]);
  const total = ref(0);
  const pageNum = ref(1);
  const pageSize = 20;
  const hasMore = computed(() => list.value.length < total.value);

  async function load(reset = true) {
    if (reset) pageNum.value = 1;
    loading.value = true;
    try {
      const res = await getWorklogs(projectId, { pageNum: pageNum.value, pageSize });
      if (res) {
        list.value = res.list || [];
        total.value = res.total || 0;
      }
    } catch {
      message.error('加载工作日志失败');
    } finally {
      loading.value = false;
    }
  }

  async function loadMore() {
    if (!hasMore.value) return;
    loadingMore.value = true;
    try {
      pageNum.value += 1;
      const res = await getWorklogs(projectId, { pageNum: pageNum.value, pageSize });
      if (res) {
        list.value = [...list.value, ...(res.list || [])];
        total.value = res.total || 0;
      }
    } catch {
      pageNum.value -= 1;
      message.error('加载更多失败');
    } finally {
      loadingMore.value = false;
    }
  }

  // 按日期分组（列表本身倒序，组内即时间倒序）
  const groups = computed(() => {
    const map = new Map<string, WorklogItem[]>();
    for (const it of list.value) {
      const d = (it.createdAt || '').slice(0, 10) || '未知日期';
      if (!map.has(d)) map.set(d, []);
      map.get(d)!.push(it);
    }
    return [...map.entries()].map(([date, items]) => ({ date, items, label: dateLabel(date) }));
  });

  function dateLabel(date: string): string {
    const today = new Date();
    const fmt = (d: Date) =>
      `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
    if (date === fmt(today)) return '今天';
    const yest = new Date(today.getTime() - 86400000);
    if (date === fmt(yest)) return '昨天';
    const weekdays = ['日', '一', '二', '三', '四', '五', '六'];
    const d = new Date(`${date}T00:00:00`);
    if (!Number.isNaN(d.getTime())) {
      const base = `${d.getMonth() + 1}月${d.getDate()}日 · 周${weekdays[d.getDay()]}`;
      return d.getFullYear() === today.getFullYear() ? base : `${d.getFullYear()}年${base}`;
    }
    return date;
  }

  // ---- 记录 / 编辑 ----
  const showEditor = ref(false);
  const editorContent = ref('');
  const editorSource = ref<'manual' | 'tasks'>('manual');
  const editing = ref<WorklogItem | null>(null);
  const saving = ref(false);

  function openCreate() {
    editing.value = null;
    editorContent.value = '';
    editorSource.value = 'manual';
    showEditor.value = true;
  }

  function openEdit(item: WorklogItem) {
    editing.value = item;
    editorContent.value = item.content;
    editorSource.value = item.source;
    showEditor.value = true;
  }

  async function save() {
    const content = editorContent.value.trim();
    if (!content) return;
    saving.value = true;
    try {
      if (editing.value) {
        await updateWorklog(editing.value.id, content);
        message.success('已保存');
        showEditor.value = false;
        load();
      } else {
        await createWorklog(projectId, { content, source: editorSource.value });
        message.success('已发布');
        showEditor.value = false;
        load();
      }
    } catch (e: any) {
      message.error(e?.message || '保存失败');
    } finally {
      saving.value = false;
    }
  }

  function confirmDelete(item: WorklogItem) {
    dialog.warning({
      title: '删除工作日志',
      content: `确定删除 ${item.createdAt.slice(0, 10)} ${item.createdAt.slice(11, 16)} 的这条日志？不可恢复。`,
      positiveText: '删除',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await deleteWorklog(item.id);
          message.success('已删除');
          list.value = list.value.filter((x) => x.id !== item.id);
          total.value = Math.max(0, total.value - 1);
        } catch (e: any) {
          message.error(e?.message || '删除失败');
        }
      },
    });
  }

  // ---- 从任务生成草稿 ----
  const showGen = ref(false);
  const genRange = ref<[number, number] | null>(null);
  const generating = ref(false);

  function openGen() {
    // 默认最近 7 天（含今天）
    const today = new Date();
    const start = new Date(today.getTime() - 6 * 86400000);
    const day = (d: Date) => new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime();
    genRange.value = [day(start), day(today)];
    showGen.value = true;
  }

  async function generate() {
    if (!genRange.value || genRange.value.length !== 2) return;
    generating.value = true;
    try {
      const fmt = (ts: number) => {
        const d = new Date(ts);
        return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
      };
      const res = await getWorklogDraft(projectId, fmt(genRange.value[0]), fmt(genRange.value[1]));
      showGen.value = false;
      editing.value = null;
      editorContent.value = res?.content || '';
      editorSource.value = 'tasks';
      showEditor.value = true;
      if (res) {
        message.info(`已汇总 ${res.taskCount} 个任务、${res.releaseCount} 个发布，润色后发布`);
      }
    } catch (e: any) {
      message.error(e?.message || '生成草稿失败');
    } finally {
      generating.value = false;
    }
  }

  onMounted(() => load());
</script>

<style lang="less" scoped>
  .wl-group {
    margin-bottom: 8px;
  }

  .wl-group-head {
    display: flex;
    align-items: baseline;
    gap: 10px;
    padding: 8px 0 6px;
    border-bottom: 1px solid var(--line, #e9e9e7);
    position: sticky;
    top: 0;
    z-index: 1;
    /* 跟 n-card 的背景走（--n-color 由 naive-ui 注入并继承），
       暗色主题下 sticky 头不会变成白条 */
    background: var(--n-color, var(--body-color, #fff));
  }

  .wl-date {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-1, #24292f);
  }

  .wl-count {
    font-size: 12px;
    color: var(--text-3, #8b949e);
  }

  .wl-list {
    position: relative;
  }

  .wl-list::before {
    content: '';
    position: absolute;
    left: 43px;
    top: 18px;
    bottom: 18px;
    width: 2px;
    border-radius: 1px;
    background: var(--line, #e9e9e7);
  }

  .wl-item {
    position: relative;
    padding: 10px 8px 10px 62px;
    border-radius: 6px;
  }

  .wl-item:hover {
    background: var(--hover-bg, rgba(0, 0, 0, 0.035));
  }

  .wl-item::before {
    content: '';
    position: absolute;
    left: 38px;
    top: 17px;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--primary-color, #18a058);
    box-shadow: 0 0 0 3px rgba(24, 160, 88, 0.15);
  }

  .wl-item-head {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
  }

  .wl-time {
    position: absolute;
    left: 8px;
    top: 14px;
    font-size: 12px;
    color: var(--text-3, #8b949e);
    font-variant-numeric: tabular-nums;
  }

  .wl-author {
    font-weight: 600;
    color: var(--text-1, #24292f);
  }

  .wl-src {
    margin-left: -2px;
  }

  .wl-actions {
    margin-left: auto;
    opacity: 0;
    transition: opacity 0.15s;
    display: flex;
    gap: 8px;
  }

  .wl-item:hover .wl-actions {
    opacity: 1;
  }

  .wl-content {
    margin-top: 4px;
    font-size: 13px;

    :deep(.md-editor-preview-wrapper) {
      padding: 0;
    }

    :deep(.md-editor-preview) {
      font-size: 13px;
      color: var(--text-2, #57606a);
    }
  }
</style>
