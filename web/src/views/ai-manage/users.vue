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

      <n-spin :show="loading">
        <EmptyState type="generic" title="还没有 Agent" v-if="!loading && userList.length === 0" description="外部 Agent 经注册与项目接入码加入后，会在这里出现" />
        <n-table v-else :bordered="false" :single-line="false" size="small">
          <thead>
            <tr>
              <th class="col-idx">#</th>
            <th>用户名</th>
              <th>姓名</th>
              <th>能力</th>
              <th>接入项目</th>
              <th>状态</th>
              <th>创建时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(item, __ix) in userList" :key="item.id">
              <td class="col-idx">{{ __ix + 1 }}</td>
            <td>{{ item.username }}</td>
              <td>{{ item.realName || '-' }}</td>
              <td>{{ item.capabilities || '-' }}</td>
              <td>
                <n-space :size="4">
                  <n-tag v-for="pn in item.projects || []" :key="pn" size="small" :bordered="false">{{ pn }}</n-tag>
                  <span v-if="!(item.projects || []).length" class="text-gray-400">-</span>
                </n-space>
              </td>
              <td>
                <n-tag :type="USER_STATUS.tagType(item.status)" size="small">
                  {{ USER_STATUS.label(item.status) }}
                </n-tag>
              </td>
              <td>{{ item.createdAt }}</td>
              <td>
                <n-space size="small">
                  <n-button text type="info" @click="handleEdit(item)">编辑</n-button>
                  <n-button text type="warning" @click="handleResetKey(item)">重置 Key</n-button>
                  <n-button text type="error" @click="handleDelete(item)">删除</n-button>
                </n-space>
              </td>
            </tr>
          </tbody>
        </n-table>
      </n-spin>
    </n-card>

    <!-- 创建/编辑弹窗 -->
    <n-modal
      v-model:show="showModal"
      preset="dialog"
      :title="isEdit ? '编辑 AI 用户' : '创建 Agent'"
      positive-text="确定"
      negative-text="取消"
      @positive-click="handleSubmit"
      style="width: 520px"
    >
      <n-form ref="formRef" :model="formData" :rules="formRules" label-placement="left" :label-width="80" class="py-4">
        <n-form-item label="用户名" path="username">
          <n-input v-model:value="formData.username" placeholder="请输入用户名" :disabled="isEdit" />
        </n-form-item>
        <n-form-item label="姓名" path="realName">
          <n-input v-model:value="formData.realName" placeholder="请输入姓名" />
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
  import { ref, reactive, onMounted } from 'vue';
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

  const message = useMessage();
  const dialog = useDialog();
  const loading = ref(false);
  const userList = ref<AiUserItem[]>([]);
  const showModal = ref(false);
  const showKeyModal = ref(false);
  const isEdit = ref(false);
  const editId = ref<number | null>(null);
  const formRef = ref<any>(null);
  const generatedKey = ref('');

  const formData = reactive({ username: '', realName: '', capabilities: '' });
  const formRules = {
    username: { required: true, message: '请输入用户名', trigger: 'blur' },
  };

  async function loadData() {
    loading.value = true;
    try {
      const res = await getAiUsers();
      userList.value = res?.list || [];
    } catch (e) {
      // ignore
    } finally {
      loading.value = false;
    }
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
    navigator.clipboard.writeText(generatedKey.value).then(() => {
      message.success('已复制到剪贴板');
    }).catch(() => {
      message.error('复制失败，请手动复制');
    });
  }

  onMounted(() => {
    loadData();
  });
</script>
