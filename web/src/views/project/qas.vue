<template>
  <div>
    <n-card :bordered="false" class="proCard">
      <template #header>
        QA 库 <n-text depth="3" style="font-size: 12px; font-weight: 400">常见问答沉淀，Agent 遇到问题先查这里，解决后沉淀新条目</n-text>
      </template>
      <template #header-extra>
        <n-space :size="8" align="center">
          <n-input v-model:value="keyword" size="small" placeholder="搜问题 / 答案 / 标签" clearable style="width: 200px" @keyup.enter="load" @clear="load" />
          <n-button size="small" @click="load">查询</n-button>
          <n-button size="small" type="primary" ghost @click="openEditor()">+ 沉淀 QA</n-button>
        </n-space>
      </template>

      <n-spin :show="loading">
        <EmptyState v-if="!loading && list.length === 0" type="search" title="没有匹配的 QA"
          description="解决问题后把「坑 + 解法」沉淀进来，人和 Agent 都能复用" />
        <n-space v-else vertical :size="14">
          <div v-for="qa in list" :key="qa.id" class="qa-card" @click="toggle(qa.id)">
            <div class="qa-q">
              <span class="qa-hit">{{ qa.hits }}</span>
              <span class="qa-q-text">{{ qa.question }}</span>
              <n-tag v-for="tg in splitTags(qa.tags)" :key="tg" size="tiny" :bordered="false" class="ml-2">{{ tg }}</n-tag>
              <span class="qa-meta">{{ qa.updater }} · {{ (qa.updatedAt || '').slice(0, 10) }}</span>
            </div>
            <div v-if="expanded === qa.id" class="qa-a">
              <MdPreview :id="'md-qa-' + qa.id" :model-value="qa.answer" :sanitize="safeHtml" />
              <n-space :size="8" justify="end" class="mt-2">
                <n-button size="tiny" @click.stop="openEditor(qa)">编辑</n-button>
                <n-popconfirm @positive-click="archive(qa)">
                  <template #trigger><n-button size="tiny" type="error" ghost>归档</n-button></template>
                  归档后不再出现在检索与开工包中。
                </n-popconfirm>
              </n-space>
            </div>
          </div>
        </n-space>
      </n-spin>
    </n-card>

    <!-- 沉淀/编辑：问题 + 答案（markdown）+ 标签 -->
    <n-modal v-model:show="showEditor" preset="dialog" :title="editForm.id ? '编辑 QA' : '沉淀 QA'" :show-icon="false" style="width: 640px">
      <n-space vertical :size="12" class="py-2">
        <n-input v-model:value="editForm.question" placeholder="问题（同问题再次沉淀会更新答案，不堆积）" />
        <n-input v-model:value="editForm.tags" placeholder="标签（逗号分隔，如 部署,数据库）" />
        <div>
          <div class="pd-label" style="margin-bottom: 6px">答案（markdown）</div>
          <MdEditor
            :model-value="editForm.answer"
            editor-id="qa-editor"
            :sanitize="safeHtml"
            :toolbars="mdToolbars"
            :footers="[]"
            placeholder="解法 / 原因 / 参考链接"
            style="height: 240px"
            @update:model-value="(v: string) => (editForm.answer = v)"
          />
        </div>
      </n-space>
      <template #action>
        <n-space>
          <n-button size="small" @click="showEditor = false">取消</n-button>
          <n-button size="small" type="primary" :disabled="!editForm.question.trim() || !editForm.answer.trim()" :loading="saving" @click="save">保存</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, reactive, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { useMessage } from 'naive-ui';
  import DOMPurify from 'dompurify';
  import { MdPreview, MdEditor } from 'md-editor-v3';
  import 'md-editor-v3/lib/style.css';
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { getQas, upsertQa, archiveQa } from '@/api/project/index';
  import type { QaItem } from '@/api/project/index';

  const route = useRoute();
  const message = useMessage();
  const projectId = computed(() => Number(route.params.projectId));

  const loading = ref(false);
  const list = ref<QaItem[]>([]);
  const keyword = ref('');
  const expanded = ref<number | null>(null);

  function safeHtml(md: string): string {
    return DOMPurify.sanitize(md);
  }
  function toggle(id: number) {
    expanded.value = expanded.value === id ? null : id;
  }
  function splitTags(tags: string): string[] {
    return (tags || '').split(',').map((x) => x.trim()).filter(Boolean);
  }

  async function load() {
    loading.value = true;
    try {
      const res = await getQas(projectId.value, { keyword: keyword.value || undefined });
      list.value = res?.list || [];
    } catch {
      message.error('加载 QA 失败');
    } finally {
      loading.value = false;
    }
  }

  const showEditor = ref(false);
  const saving = ref(false);
  const editForm = reactive({ id: 0, question: '', answer: '', tags: '' });

  function openEditor(qa?: QaItem) {
    Object.assign(editForm, qa
      ? { id: qa.id, question: qa.question, answer: qa.answer, tags: qa.tags }
      : { id: 0, question: '', answer: '', tags: '' });
    showEditor.value = true;
  }
  async function save() {
    if (!editForm.question.trim() || !editForm.answer.trim()) return;
    saving.value = true;
    try {
      const res = await upsertQa(projectId.value, {
        question: editForm.question.trim(), answer: editForm.answer, tags: editForm.tags,
      });
      message.success(res?.updated ? '已有同问条目，答案已更新' : 'QA 已沉淀');
      showEditor.value = false;
      load();
    } catch (e: any) {
      message.error(e?.message || '保存失败');
    } finally {
      saving.value = false;
    }
  }
  async function archive(qa: QaItem) {
    try {
      await archiveQa(projectId.value, qa.id);
      message.success('已归档');
      load();
    } catch (e: any) { message.error(e?.message || '归档失败'); }
  }

  const mdToolbars = [
    'bold', 'italic', 'title', 'quote',
    'unorderedList', 'orderedList', 'task',
    'codeRow', 'codeBlock', 'link', 'image', 'table',
    '-',
    'preview', 'pageFullscreen', 'fullscreen',
  ] as const;

  onMounted(load);
</script>

<style lang="less" scoped>
  .qa-card {
    border: 1px solid var(--border-color, #eef0f3);
    border-radius: 14px;
    padding: 13px 15px;
    cursor: pointer;
    transition: border-color 0.15s;

    &:hover { border-color: var(--primary-color, #16a34a); }
  }
  .qa-q { display: flex; align-items: center; gap: 8px; }
  .qa-hit {
    flex: none; width: 34px; height: 22px; border-radius: 7px;
    background: var(--primary-weak, rgba(22, 163, 74, 0.08));
    color: var(--primary-color, #16a34a);
    font-size: 12px; font-weight: 700; text-align: center; line-height: 22px;
  }
  .qa-q-text { font-weight: 600; color: var(--text-color-1, #1f2329); flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .qa-meta { margin-left: auto; flex: none; font-size: 11px; color: var(--text-color-3, #9aa1ab); }
  .qa-a { margin-top: 10px; border-top: 1px dashed var(--border-color, #eef0f3); padding-top: 10px; cursor: default; }
  .pd-label { font-size: 12px; font-weight: 700; color: var(--text-color-2, #5c6470); }
</style>
