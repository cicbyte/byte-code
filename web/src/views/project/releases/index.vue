<template>
  <div>
    <n-card :bordered="false" class="proCard">
      <n-space size="small" align="center" justify="space-between" class="mb-2">
        <n-space size="small" align="center">
          <n-select
            v-model:value="channelFilter"
            :options="channelFilterOptions"
            size="small"
            style="width: 110px"
            @update:value="reload"
          />
          <span class="text-xs text-gray-400">本地打包上传，团队内直接下载，免IM传文件</span>
        </n-space>
        <n-button type="primary" size="small" @click="openCreate" data-test-id="releases.create-btn">
          <template #icon>
            <n-icon><PlusOutlined /></n-icon>
          </template>
          新建发布
        </n-button>
      </n-space>

      <n-spin :show="loading">
        <EmptyState
          type="data"
          title="暂无发布"
          description="本地构建后新建发布并上传安装包，团队成员即可在此下载"
          v-if="!loading && list.length === 0"
        />
        <n-table v-else :bordered="false" :single-line="false" size="small">
          <thead>
            <tr>
              <th>版本</th>
              <th>标题</th>
              <th>文件</th>
              <th>发布人</th>
              <th>时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <template v-for="r in list" :key="r.id">
              <tr :data-test-id="`releases.row-${r.id}`">
                <td>
                  <n-space size="small" :wrap="false" align="center">
                    <n-tag size="small" type="success">{{ r.version }}</n-tag>
                    <n-tag size="small" :bordered="false" :type="channelTagType(r.channel)">
                      {{ channelLabel(r.channel) }}
                    </n-tag>
                  </n-space>
                </td>
                <td>
                  <span :title="r.notes">{{ r.title }}</span>
                </td>
                <td>
                  <n-space size="small" :wrap="false" align="center">
                    <n-tag
                      v-for="f in filesOf(r.id)"
                      :key="f.id"
                      size="small"
                      class="cursor-pointer"
                      :title="`${f.originalName}（${fmtSize(f.fileSize)}）点击下载`"
                      :data-test-id="`releases.file-${f.id}`"
                      @click="download(f.id)"
                    >
                      {{ shortName(f.originalName) }} · {{ fmtSize(f.fileSize) }}
                    </n-tag>
                    <n-upload
                      :show-file-list="false"
                      :custom-request="(o: any) => uploadFile(r.id, o)"
                      multiple
                    >
                      <n-button text type="info" size="small" :data-test-id="`releases.upload-${r.id}`">
                        + 上传
                      </n-button>
                    </n-upload>
                    <span v-if="r.fileCount === 0" class="text-gray-400 text-xs">未上传文件</span>
                  </n-space>
                </td>
                <td>{{ r.createdByName || '-' }}</td>
                <td>{{ r.createdAt }}</td>
                <td>
                  <n-space size="small">
                    <n-button text type="info" @click="openEdit(r)" :data-test-id="`releases.edit-btn-${r.id}`">
                      编辑
                    </n-button>
                    <n-button text type="error" @click="handleDelete(r)" :data-test-id="`releases.del-btn-${r.id}`">
                      删除
                    </n-button>
                  </n-space>
                </td>
              </tr>
              <tr v-if="r.notes">
                <td colspan="6" class="notes-cell" @click="toggleNotes(r.id)">
                  <pre class="notes-text" :class="{ clamp: !expandedNotes.has(r.id) }">{{ r.notes }}</pre>
                </td>
              </tr>
            </template>
          </tbody>
        </n-table>
      </n-spin>

      <div class="mt-4 flex justify-end" v-if="total > pageSize">
        <n-pagination v-model:page="page" :page-size="pageSize" :item-count="total" @update:page="loadData" />
      </div>
    </n-card>

    <!-- 新建/编辑发布 -->
    <n-modal
      v-model:show="showModal"
      preset="dialog"
      :title="isEdit ? '编辑发布' : '新建发布'"
      positive-text="确定"
      negative-text="取消"
      @positive-click="submit"
      style="width: 620px"
    >
      <n-form :model="form" label-placement="left" :label-width="72" class="py-4" :rules="formRules" ref="formRef">
        <n-form-item label="版本号" path="version" v-if="!isEdit">
          <n-input v-model:value="form.version" placeholder="如 v1.2.0（项目内唯一）" data-test-id="releases.version-input" />
        </n-form-item>
        <n-form-item label="标题">
          <n-input v-model:value="form.title" placeholder="缺省同版本号" />
        </n-form-item>
        <n-form-item label="渠道">
          <n-select v-model:value="form.channel" :options="channelOptions" />
        </n-form-item>
        <n-form-item label="发布说明">
          <n-input v-model:value="form.notes" type="textarea" placeholder="changelog / 安装说明（点击列表行可展开）" :rows="6" />
        </n-form-item>
      </n-form>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { ref, reactive, computed, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { useMessage, useDialog } from 'naive-ui';
  import { PlusOutlined } from '@vicons/antd';
  import { getReleases, createRelease, updateRelease, deleteRelease } from '@/api/project/release';
  import type { ReleaseItem } from '@/api/project/release';
  import {
    getAttachments,
    uploadAttachment,
    downloadAttachment,
    deleteAttachment,
  } from '@/api/attachment/index';
  import type { AttachmentItem } from '@/api/attachment/index';

  const message = useMessage();
  const dialog = useDialog();
  const route = useRoute();
  const projectId = computed(() => Number(route.params.projectId));

  const channelOptions = [
    { label: '正式（stable）', value: 'stable' },
    { label: '内测（beta）', value: 'beta' },
    { label: '每日（nightly）', value: 'nightly' },
  ];
  const channelFilterOptions = [
    { label: '全部渠道', value: '' },
    ...channelOptions,
  ];
  function channelLabel(c: string) {
    return channelOptions.find((o) => o.value === c)?.label || c;
  }
  function channelTagType(c: string): 'success' | 'warning' | 'info' {
    if (c === 'stable') return 'success';
    if (c === 'beta') return 'warning';
    return 'info';
  }
  function fmtSize(bytes: number): string {
    if (!bytes || bytes <= 0) return '-';
    if (bytes < 1024) return `${bytes}B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)}KB`;
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)}MB`;
    return `${(bytes / 1024 / 1024 / 1024).toFixed(2)}GB`;
  }
  function shortName(name: string): string {
    return name.length > 18 ? name.slice(0, 16) + '…' : name;
  }

  // ---------- 列表 ----------
  const loading = ref(false);
  const list = ref<ReleaseItem[]>([]);
  const total = ref(0);
  const page = ref(1);
  const pageSize = 20;
  const channelFilter = ref<string | null>(null);
  const expandedNotes = ref(new Set<number>());

  async function loadData() {
    loading.value = true;
    try {
      const res = await getReleases(projectId.value, {
        pageNum: page.value,
        pageSize,
        channel: channelFilter.value || undefined,
      });
      list.value = res?.list || [];
      total.value = res?.total || 0;
      expandedNotes.value = new Set();
      loadFiles();
    } catch (e) {
      // ignore
    } finally {
      loading.value = false;
    }
  }
  function reload() {
    page.value = 1;
    loadData();
  }
  function toggleNotes(id: number) {
    if (expandedNotes.value.has(id)) expandedNotes.value.delete(id);
    else expandedNotes.value.add(id);
    expandedNotes.value = new Set(expandedNotes.value);
  }

  // ---------- 文件（复用附件通道 entityType=release） ----------
  const filesMap = ref(new Map<number, AttachmentItem[]>());
  function filesOf(rid: number): AttachmentItem[] {
    return filesMap.value.get(rid) || [];
  }
  async function loadFiles() {
    if (list.value.length === 0) return;
    const results = await Promise.all(
      list.value.map((r) =>
        getAttachments('release', r.id)
          .then((res) => ({ rid: r.id, files: res?.list || [] }))
          .catch(() => ({ rid: r.id, files: [] }))
      )
    );
    const m = new Map<number, AttachmentItem[]>();
    for (const it of results) {
      if (it.files.length) m.set(it.rid, it.files);
    }
    filesMap.value = m;
  }
  async function uploadFile(rid: number, o: { file: { file: File } }) {
    try {
      await uploadAttachment({ entityType: 'release', entityId: rid, file: o.file.file });
      message.success('上传成功');
      loadData();
    } catch (e) {
      message.error('上传失败');
    }
  }
  async function download(attId: number) {
    try {
      const res = await downloadAttachment(attId);
      if (res?.url) window.open(res.url, '_blank');
    } catch (e) {
      message.error('获取下载链接失败');
    }
  }

  // ---------- 新建/编辑 ----------
  const showModal = ref(false);
  const isEdit = ref(false);
  const editId = ref<number | null>(null);
  const formRef = ref<any>(null);
  const form = reactive({ version: '', title: '', notes: '', channel: 'stable' });
  const formRules = {
    version: { required: true, message: '请输入版本号', trigger: 'blur' },
  };

  function openCreate() {
    isEdit.value = false;
    editId.value = null;
    form.version = '';
    form.title = '';
    form.notes = '';
    form.channel = 'stable';
    showModal.value = true;
  }
  function openEdit(r: ReleaseItem) {
    isEdit.value = true;
    editId.value = r.id;
    form.title = r.title;
    form.notes = r.notes;
    form.channel = r.channel;
    showModal.value = true;
  }
  async function submit() {
    if (!isEdit.value) {
      try {
        await formRef.value?.validate();
      } catch {
        return false;
      }
    }
    try {
      if (isEdit.value && editId.value) {
        await updateRelease(editId.value, {
          title: form.title,
          notes: form.notes,
          channel: form.channel,
        });
        message.success('更新成功');
      } else {
        await createRelease(projectId.value, { ...form });
        message.success('发布已创建，可上传安装包');
      }
      showModal.value = false;
      loadData();
    } catch (e: any) {
      message.error(e?.message || '操作失败');
      return false;
    }
  }

  function handleDelete(r: ReleaseItem) {
    dialog.warning({
      title: '删除发布',
      content: `确定删除发布 ${r.version}（含 ${r.fileCount} 个文件记录）？该操作不可恢复。`,
      positiveText: '删除',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await deleteRelease(r.id);
          message.success('已删除');
          loadData();
        } catch (e) {
          message.error('删除失败');
        }
      },
    });
  }

  onMounted(() => {
    loadData();
  });
</script>

<style scoped>
  .notes-cell {
    cursor: pointer;
    background: rgba(128, 128, 128, 0.04);
  }
  .notes-text {
    margin: 0;
    font-size: 12px;
    line-height: 1.7;
    white-space: pre-wrap;
    word-break: break-all;
    font-family: inherit;
  }
  .notes-text.clamp {
    max-height: 22px;
    overflow: hidden;
  }
</style>
