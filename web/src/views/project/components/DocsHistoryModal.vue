<template>
  <n-modal
    :show="show"
    preset="card"
    title="版本历史"
    style="width: 720px"
    :bordered="false"
    @update:show="$emit('update:show', $event)"
  >
    <n-spin :show="loading">
      <EmptyState type="doc" title="暂无历史版本" v-if="!loading && items.length === 0" description="保存或删除文档时会自动快照" />
      <n-list v-else bordered clickable>
        <n-list-item v-for="it in items" :key="it.snapshot" @click="preview(it.snapshot)">
          <template #prefix>
            <n-icon size="16"><HistoryOutlined /></n-icon>
          </template>
          <n-thing>
            <template #header>{{ formatSnapshot(it.snapshot) }}</template>
            <template #description>{{ formatSize(it.size) }}</template>
          </n-thing>
          <template #suffix>
            <n-space :size="4">
              <n-button size="tiny" @click.stop="preview(it.snapshot)">查看</n-button>
              <n-button size="tiny" type="warning" ghost @click.stop="restore(it.snapshot)">恢复</n-button>
            </n-space>
          </template>
        </n-list-item>
      </n-list>
    </n-spin>

    <!-- 快照内容预览 -->
    <n-modal
      v-model:show="previewShow"
      preset="card"
      :title="`快照内容 · ${previewing}`"
      style="width: 680px; max-height: 80vh"
      :bordered="false"
    >
      <n-input :value="previewContent" type="textarea" readonly :autosize="{ minRows: 8, maxRows: 20 }" />
    </n-modal>
  </n-modal>
</template>

<script lang="ts" setup>
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { ref, watch } from 'vue';
  import { useMessage, useDialog } from 'naive-ui';
  import { HistoryOutlined } from '@vicons/antd';
  import { getDocsHistory, readDocsHistory, restoreDocsHistory } from '@/api/docs/index';
  import type { HistoryItem } from '@/api/docs/index';

  const props = defineProps<{
    show: boolean;
    projectId: number;
    path: string;
  }>();
  const emit = defineEmits(['update:show', 'restored']);

  const message = useMessage();
  const dialog = useDialog();

  const loading = ref(false);
  const items = ref<HistoryItem[]>([]);
  const previewShow = ref(false);
  const previewing = ref('');
  const previewContent = ref('');

  function formatSize(n: number) {
    if (n < 1024) return `${n} B`;
    return `${(n / 1024).toFixed(1)} KB`;
  }

  // 快照标识 20260904-112137.000(-01) → 可读时间（含序号）
  function formatSnapshot(s: string) {
    const m = s.match(/^(\d{4})(\d{2})(\d{2})-(\d{2})(\d{2})(\d{2})(?:\.(\d{3}))?(?:-(\d+))?$/);
    if (!m) return s;
    const [, y, mo, d, h, mi, se, ms, seq] = m;
    return `${y}-${mo}-${d} ${h}:${mi}:${se}${ms ? `.${ms}` : ''}${seq ? ` #${seq}` : ''}`;
  }

  async function load() {
    if (!props.path) return;
    loading.value = true;
    try {
      const res = await getDocsHistory(props.projectId, props.path);
      items.value = res?.list || [];
    } catch {
      items.value = [];
    } finally {
      loading.value = false;
    }
  }

  watch(
    () => props.show,
    (v) => v && load()
  );

  async function preview(snapshot: string) {
    try {
      const res = await readDocsHistory(props.projectId, props.path, snapshot);
      previewing.value = formatSnapshot(snapshot);
      previewContent.value = res?.content || '';
      previewShow.value = true;
    } catch {
      message.error('读取快照失败');
    }
  }

  function restore(snapshot: string) {
    dialog.warning({
      title: '恢复该版本',
      content: `将把「${formatSnapshot(snapshot)}」的内容恢复为当前版本。当前内容会自动快照保底（恢复可再撤销）。`,
      positiveText: '恢复',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await restoreDocsHistory(props.projectId, props.path, snapshot);
          message.success('已恢复（原内容已保底快照）');
          emit('restored');
          emit('update:show', false);
        } catch {
          message.error('恢复失败');
        }
      },
    });
  }
</script>
