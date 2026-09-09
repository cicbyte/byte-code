<template>
  <div>
    <n-card :bordered="false" :segmented="{ content: true }">
      <template #header>
        全局记忆
        <n-tag size="small" :bordered="false" style="margin-left: 8px">project_id = 0</n-tag>
      </template>
      <template #header-extra>
        <span class="header-hint">跨项目通用约定，对所有项目的 AI 上下文生效</span>
      </template>

      <n-alert type="info" :show-icon="true" class="mb-4" closable>
        全局记忆是跨项目共享的 KV 约定（如公司代码风格、通用部署规范）。AI 开工时
        <code>kb_get_conventions</code> 会自动注入 <code>conventions.*</code> 前缀的全局记忆，
        <code>mem_list</code> / <code>mem_get</code> 亦会带出全局条目（同名 key 项目记忆优先）。
        仅管理员可修改，全员可读。
      </n-alert>

      <!-- 过滤栏：左过滤右操作，单行排布 -->
      <n-space class="mb-4" align="center" justify="space-between">
        <n-space align="center">
          <n-input
            v-model:value="prefix"
            size="small"
            style="width: 220px"
            placeholder="key 前缀过滤，例：conventions."
            clearable
            @keyup.enter="load"
            @clear="load"
          />
          <n-checkbox-group v-model:value="includeStates" size="small" @update:value="load">
            <n-space>
              <n-checkbox value="stale" label="含腐化 (stale)" />
              <n-checkbox value="expired" label="含过期 (expired)" />
            </n-space>
          </n-checkbox-group>
          <n-button size="small" @click="load">查询</n-button>
        </n-space>
        <n-button v-if="isAdmin" type="primary" size="small" @click="openCreate">新增记忆</n-button>
      </n-space>

      <n-data-table
        :columns="columns"
        :data="list"
        :loading="loading"
        :row-key="(row: MemoryItem) => row.key"
        size="small"
      />
    </n-card>

    <!-- 新增/编辑抽屉（大文本 + markdown 编辑器适合抽屉；宽度与任务详情对齐） -->
    <n-drawer v-model:show="showModal" :width="drawerWidth" placement="right">
      <n-drawer-content :title="editingKey ? `编辑全局记忆：${editingKey}` : '新增全局记忆'" closable>
        <n-form label-placement="top" class="px-1">
          <n-form-item v-if="!editingKey" label="key" required>
            <n-input v-model:value="form.key" placeholder="点分层级，例：conventions.code-style / deploy.default" />
          </n-form-item>
          <n-form-item label="value" required>
            <MdEditor
              v-model="form.value"
              editor-id="global-memory-editor"
              :theme="isDark ? 'dark' : 'light'"
              placeholder="支持 markdown（代码块/列表/表格），上限 64KB"
              :toolbars="mdToolbars"
              :footers="[]"
              style="height: 360px"
            />
          </n-form-item>
          <div class="form-grid">
            <n-form-item label="状态">
              <n-radio-group v-model:value="form.status">
                <n-radio value="active">active（已验证）</n-radio>
                <n-radio value="pending">pending（待验证）</n-radio>
              </n-radio-group>
            </n-form-item>
            <n-form-item label="有效期">
              <n-select
                v-model:value="form.ttl"
                :options="ttlOptions"
                placeholder="缺省永不过期"
                clearable
              />
            </n-form-item>
          </div>
        </n-form>
        <template #footer>
          <n-space>
            <n-button @click="showModal = false">取消</n-button>
            <n-button type="primary" @click="handleSubmit">确定</n-button>
          </n-space>
        </template>
      </n-drawer-content>
    </n-drawer>
  </div>
</template>

<script lang="ts" setup>
  import { MEMORY_STATUS } from '@/enums/entities';
  import { ref, reactive, computed, h, onMounted } from 'vue';
  import { NButton, NSpace, NTag, useMessage, useDialog } from 'naive-ui';
  import type { DataTableColumns } from 'naive-ui';
  import { MdEditor } from 'md-editor-v3';
  import 'md-editor-v3/lib/style.css';
  import { useDesignSetting } from '@/hooks/setting/useDesignSetting';
  import { mdToolbars, editDrawerWidth } from '@/utils/mdEditor';
  import {
    getGlobalMemories,
    setGlobalMemory,
    verifyGlobalMemory,
    expireGlobalMemory,
    deleteGlobalMemory,
  } from '@/api/docs/index';
  import type { MemoryItem } from '@/api/docs/index';
  import { useUserStore } from '@/store/modules/user';

  const message = useMessage();
  const { getDarkTheme: isDark } = useDesignSetting();
  const dialog = useDialog();

  // 与 asyncRoute 的菜单过滤同规则：拥有系统菜单/角色管理权限即视为管理员
  const userStore = useUserStore();
  const isAdmin = computed(() => {
    const perms = new Set((userStore.permissions || []).map((p: any) => p?.value || p));
    return perms.has('system_menu') || perms.has('system_role');
  });

  const loading = ref(false);
  const list = ref<MemoryItem[]>([]);
  const prefix = ref('');
  const includeStates = ref<string[]>([]);
  const drawerWidth = editDrawerWidth();

  const showModal = ref(false);
  const editingKey = ref('');
  const form = reactive({ key: '', value: '', status: 'active', ttl: null as string | null });

  const ttlOptions = [
    { label: '30 分钟', value: '30m' },
    { label: '12 小时', value: '12h' },
    { label: '7 天', value: '7d' },
    { label: '30 天', value: '30d' },
  ];

  // 字典统一出口：enums/entities.ts（与 memories.vue 共用）
  const statusTagType = MEMORY_STATUS.tagType;
  const statusLabel = MEMORY_STATUS.label;

  const columns = computed<DataTableColumns<MemoryItem>>(() => {
    const cols: DataTableColumns<MemoryItem> = [
      { title: 'key', key: 'key', width: 240, ellipsis: { tooltip: true } },
      { title: 'value', key: 'value', ellipsis: { tooltip: true } },
      {
        title: '状态',
        key: 'status',
        width: 110,
        render: (row) =>
          h(NSpace, { size: 4, align: 'center' }, () => [
            h(NTag, { size: 'small', type: statusTagType(row.status) }, () => statusLabel(row.status)),
            row.status === 'active' && row.staleDays >= 21
              ? h(NTag, { size: 'small', type: 'warning', bordered: false }, () => `${row.staleDays}天未验证`)
              : null,
          ]),
      },
      { title: '来源', key: 'source', width: 110, ellipsis: { tooltip: true } },
      { title: '上次验证', key: 'lastVerifiedAt', width: 160 },
    ];
    if (isAdmin.value) {
      cols.push({
        title: '操作',
        key: 'actions',
        width: 220,
        render: (row) =>
          h(NSpace, { size: 4 }, () => [
            h(NButton, { size: 'tiny', onClick: () => openEdit(row) }, () => '编辑'),
            row.status !== 'active'
              ? h(NButton, { size: 'tiny', type: 'success', ghost: true, onClick: () => handleVerify(row) }, () => '验证保鲜')
              : null,
            row.status !== 'expired'
              ? h(NButton, { size: 'tiny', type: 'warning', ghost: true, onClick: () => handleExpire(row) }, () => '废弃')
              : null,
            h(NButton, { size: 'tiny', type: 'error', ghost: true, onClick: () => handleDelete(row) }, () => '删除'),
          ]),
      });
    }
    return cols;
  });

  async function load() {
    loading.value = true;
    try {
      const include = includeStates.value.join(',');
      const res = await getGlobalMemories(prefix.value || undefined, include || undefined);
      list.value = res?.list || [];
    } catch {
      // http 层统一提示
    } finally {
      loading.value = false;
    }
  }

  function openCreate() {
    editingKey.value = '';
    form.key = '';
    form.value = '';
    form.status = 'active';
    form.ttl = null;
    showModal.value = true;
  }

  function openEdit(row: MemoryItem) {
    editingKey.value = row.key;
    form.key = row.key;
    form.value = row.value;
    form.status = row.status === 'pending' ? 'pending' : 'active';
    form.ttl = null;
    showModal.value = true;
  }

  async function handleSubmit() {
    const key = (editingKey.value || form.key).trim();
    if (!key || !form.value) {
      message.warning('key 和 value 不能为空');
      return false;
    }
    try {
      await setGlobalMemory(key, {
        value: form.value,
        status: form.status as 'active' | 'pending',
        ...(form.ttl ? { ttl: form.ttl } : {}),
      });
      message.success('已保存');
      showModal.value = false;
      load();
    } catch {
      message.error('保存失败');
      return false;
    }
  }

  async function handleVerify(row: MemoryItem) {
    try {
      await verifyGlobalMemory(row.key);
      message.success(`已验证：${row.key}`);
      load();
    } catch {
      message.error('操作失败');
    }
  }

  function handleExpire(row: MemoryItem) {
    dialog.warning({
      title: '确认废弃',
      content: `确定将「${row.key}」标记为过期吗？废弃后所有项目的 AI 查询默认不可见。`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await expireGlobalMemory(row.key);
          message.success('已废弃');
          load();
        } catch {
          message.error('操作失败');
        }
      },
    });
  }

  function handleDelete(row: MemoryItem) {
    dialog.warning({
      title: '确认删除',
      content: `确定删除「${row.key}」吗？该操作不可恢复，所有项目的 AI 均不再可见。`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await deleteGlobalMemory(row.key);
          message.success('已删除');
          load();
        } catch {
          message.error('删除失败');
        }
      },
    });
  }

  onMounted(() => {
    load();
  });
</script>

<style lang="less" scoped>
  .header-hint {
    font-size: 12px;
    color: #999;
    font-weight: normal;
  }

  .form-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0 16px;
  }
</style>
