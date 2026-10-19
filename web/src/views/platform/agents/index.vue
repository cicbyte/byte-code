<template>
  <div>
    <n-card :bordered="false" title="Agent 账号管理" class="proCard">
      <template #header-extra>
        <n-button type="primary" @click="handleCreate">
          <template #icon>
            <n-icon><PlusOutlined /></n-icon>
          </template>
          创建 Agent
        </n-button>
      </template>

      <ProTable
        ref="tableRef"
        table-key="platform-agents"
        :columns="columns"
        :request="tableRequest"
        :row-key="(row: AiUserItem) => row.id"
        :row-props="rowProps"
        :striped="true"
      >
        <template #empty>
          <EmptyState type="generic" title="还没有 Agent" description="外部 Agent 经注册与项目接入码加入后，会在这里出现" />
        </template>
      </ProTable>
    </n-card>

    <!-- 项目绑定总览（含各项目能力集；调整入口在各项目成员页） -->
    <n-modal v-model:show="showBindings" preset="dialog" :title="`项目绑定 · ${bindingsTarget?.username || ''}`" :show-icon="false" style="width: 620px">
      <n-spin :show="bindingsLoading">
        <EmptyState type="generic" title="未接入任何项目" v-if="!bindingsLoading && bindingList.length === 0" compact />
        <n-table v-else :bordered="false" :single-line="false" size="small">
          <thead><tr><th>项目</th><th>能力集</th><th>接入时间</th></tr></thead>
          <tbody>
            <tr v-for="b in bindingList" :key="b.projectId" :data-test-id="`platform-agents.row-${b.projectId}`">
              <td>{{ b.projectName }}（{{ b.projectCode }}）</td>
              <td>
                <n-tag v-if="!b.capabilities" size="small" type="success" :bordered="false">全部能力</n-tag>
                <n-space v-else :size="4" style="display: inline-flex">
                  <n-tag v-for="c in b.capabilities.split(',')" :key="c" size="small" :bordered="false" type="warning" :data-test-id="`platform-agents.item-${c}`">{{ capLabel(c) }}</n-tag>
                </n-space>
              </td>
              <td>{{ (b.joinedAt || '').slice(0, 16) }}</td>
            </tr>
          </tbody>
        </n-table>
      </n-spin>
      <template #action>
        <n-button size="small" @click="showBindings = false">关闭</n-button>
      </template>
    </n-modal>

    <!-- 创建/编辑弹窗 -->
    <n-modal
      v-model:show="showModal"
      preset="dialog"
      :title="isEdit ? '编辑 Agent' : '创建 Agent'"
      positive-text="确定"
      negative-text="取消"
      @positive-click="handleSubmit"
      style="width: 520px"
    >
      <n-form ref="formRef" :model="formData" :rules="formRules" label-placement="left" :label-width="80" class="py-4">
        <n-form-item label="账号" path="username">
          <n-input v-model:value="formData.username" placeholder="请输入账号（如 agent-zcode）" :disabled="isEdit" />
        </n-form-item>
        <n-form-item label="名称" path="realName">
          <n-input v-model:value="formData.realName" placeholder="请输入名称" />
        </n-form-item>
        <n-form-item label="能力描述" path="capabilities">
          <n-input v-model:value="formData.capabilities" type="textarea" placeholder="请输入能力描述" :rows="3" />
        </n-form-item>
      </n-form>
    </n-modal>

    <!-- API Key 展示弹窗 -->
    <n-modal v-model:show="showKeyModal" preset="dialog" title="API Key 已生成" :show-icon="false" style="width: 480px">
      <n-alert type="warning" class="mb-3">
        请立即复制保存 API Key，关闭后将无法再次查看完整内容。
      </n-alert>
      <n-input :value="generatedKey" readonly type="textarea" :rows="3" />
      <template #action>
        <n-space>
          <n-button @click="showKeyModal = false">关闭</n-button>
          <n-button type="primary" @click="copyKey">复制</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { ref, reactive, h, onMounted } from 'vue';
  import { USER_STATUS } from '@/enums/entities';
  import { useMessage, useDialog } from 'naive-ui';
  import { PlusOutlined } from '@vicons/antd';
  import {
    getAiUsers,
    createAiUser,
    updateAiUser,
    deleteAiUser,
    resetAiUserKey,
  } from '@/api/ai/index';
  import type { AiUserItem } from '@/api/ai/index';
  import { getAgentBindings } from '@/api/ai/index';
  import { AGENT_CAPS } from '@/api/agent/index';
  import { copyToClipboard } from '@/utils/clipboard';
  import ProTable from '@/components/ProTable/index.vue';
  import { NButton, NSpace, NTag } from 'naive-ui';
  import type { DataTableColumns } from 'naive-ui';

  const message = useMessage();
  const dialog = useDialog();
  const tableRef = ref();

  function rowProps(row: AiUserItem) {
    return { 'data-test-id': `platform-agents.row-${row.id}` };
  }

  async function tableRequest(p: { page: number; pageSize: number }) {
    const res = await getAiUsers({ page: p.page, size: p.pageSize });
    return { list: res?.list || [], itemCount: res?.total || 0 };
  }

  const columns: DataTableColumns<AiUserItem> = [
    { title: '账号', key: 'username', width: 150, render: (row) => row.username },
    { title: '名称', key: 'realName', width: 130, render: (row) => row.realName || '-' },
    { title: '能力描述', key: 'capabilities', ellipsis: { tooltip: true }, render: (row) => row.capabilities || '-' },
    {
      title: '接入项目',
      key: 'projects',
      render: (row) =>
        h(NSpace, { size: 4 }, () => [
          ...(row.projects || []).map((pn) =>
            h(NTag, { size: 'small', bordered: false, 'data-test-id': `platform-agents.item-${pn}` }, () => pn)
          ),
          !(row.projects || []).length ? h('span', { class: 'text-gray-400' }, '-') : null,
        ]),
    },
    {
      title: '状态',
      key: 'status',
      width: 90,
      render: (row) => h(NTag, { type: USER_STATUS.tagType(row.status), size: 'small' }, () => USER_STATUS.label(row.status)),
    },
    { title: '创建时间', key: 'createdAt', width: 160, render: (row) => row.createdAt },
    {
      title: '注册 IP',
      key: 'registerIp',
      width: 130,
      render: (row) =>
        row.registerIp
          ? h('span', { class: 'text-xs' }, row.registerIp)
          : h('span', { class: 'text-gray-400 text-xs' }, '管理端创建'),
    },
    {
      title: '操作',
      key: 'actions',
      width: 200,
      render: (row) =>
        h(NSpace, { size: 'small' }, () => [
          h(NButton, { text: true, type: 'info', onClick: () => handleBindings(row) }, () => '绑定'),
          h(NButton, { text: true, type: 'info', onClick: () => handleEdit(row) }, () => '编辑'),
          h(NButton, { text: true, type: 'warning', onClick: () => handleResetKey(row) }, () => '重置 Key'),
          h(NButton, { text: true, type: 'error', onClick: () => handleDelete(row) }, () => '删除'),
        ]),
    },
  ];

  // ==================== 项目绑定总览 ====================
  const showBindings = ref(false);
  const bindingsLoading = ref(false);
  const bindingsTarget = ref<AiUserItem | null>(null);
  const bindingList = ref<Array<{ projectId: number; projectName: string; projectCode: string; capabilities: string; joinedAt: string }>>([]);
  function capLabel(key: string) {
    return AGENT_CAPS.find((c) => c.key === key)?.label || key;
  }
  async function handleBindings(item: AiUserItem) {
    bindingsTarget.value = item;
    showBindings.value = true;
    bindingsLoading.value = true;
    try {
      const res = await getAgentBindings(item.id);
      bindingList.value = res?.list || [];
    } catch {
      bindingList.value = [];
    } finally {
      bindingsLoading.value = false;
    }
  }
  const showModal = ref(false);
  const showKeyModal = ref(false);
  const isEdit = ref(false);
  const editId = ref<number | null>(null);
  const formRef = ref<any>(null);
  const generatedKey = ref('');

  const formData = reactive({ username: '', realName: '', capabilities: '' });
  const formRules = {
    username: { required: true, message: '请输入账号', trigger: 'blur' },
  };

  function loadData() {
    tableRef.value?.reload();
  }

  function handleCreate() {
    formData.username = '';
    formData.realName = '';
    formData.capabilities = '';
    isEdit.value = false;
    editId.value = null;
    showModal.value = true;
  }

  function handleEdit(item: AiUserItem) {
    isEdit.value = true;
    editId.value = item.id;
    formData.username = item.username;
    formData.realName = item.realName;
    formData.capabilities = item.capabilities;
    showModal.value = true;
  }

  async function handleSubmit() {
    try {
      await formRef.value?.validate();
    } catch {
      return false;
    }
    try {
      if (isEdit.value && editId.value) {
        await updateAiUser(editId.value, {
          realName: formData.realName,
          capabilities: formData.capabilities,
        });
        message.success('更新成功');
      } else {
        const res = await createAiUser({
          username: formData.username,
          realName: formData.realName,
          capabilities: formData.capabilities,
        });
        message.success('创建成功');
        if (res?.apiKey) {
          generatedKey.value = res.apiKey;
          showKeyModal.value = true;
        }
      }
      showModal.value = false;
      loadData();
    } catch (e) {
      message.error('操作失败');
      return false;
    }
  }

  function handleResetKey(item: AiUserItem) {
    dialog.warning({
      title: '确认重置',
      content: `确定要重置「${item.username}」的 API Key 吗？旧 Key 将立即失效。`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          const res = await resetAiUserKey(item.id);
          message.success('Key 已重置');
          if (res?.apiKey) {
            generatedKey.value = res.apiKey;
            showKeyModal.value = true;
          }
          loadData();
        } catch (e) {
          message.error('重置失败');
        }
      },
    });
  }

  function handleDelete(item: AiUserItem) {
    dialog.warning({
      title: '确认删除',
      content: `确定要删除「${item.username}」吗？`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await deleteAiUser(item.id);
          message.success('删除成功');
          loadData();
        } catch (e) {
          message.error('删除失败');
        }
      },
    });
  }

  function copyKey() {
    copyToClipboard(generatedKey.value).then(() => {
      message.success('已复制到剪贴板');
    }).catch(() => {
      message.error('复制失败，请手动复制');
    });
  }

  onMounted(() => {
    loadData();
  });
</script>

<style lang="less" scoped>
  // 账号列走等宽字体：与项目 code 徽章同思路，标识类字段更易读
  .agent-account {
    font-family: 'JetBrains Mono', Consolas, monospace;
    font-size: 13px;
  }
</style>
