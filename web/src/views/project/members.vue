<template>
  <div>
    <n-card :bordered="false" title="项目成员" class="proCard">
      <template #header-extra>
        <n-button type="primary" @click="showAddModal = true">添加成员</n-button>
      </template>

      <n-spin :show="loading">
        <n-empty v-if="!loading && memberList.length === 0" description="暂无成员" />
        <n-table v-else :bordered="false" :single-line="false" size="small">
          <thead>
            <tr>
              <th>用户名</th>
              <th>姓名</th>
              <th>角色</th>
              <th>加入时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="member in memberList" :key="member.id">
              <td>{{ member.username }}</td>
              <td>{{ member.realName }}</td>
              <td>{{ member.role }}</td>
              <td>{{ member.joinedAt }}</td>
              <td>
                <n-button text type="error" @click="handleRemove(member)">移除</n-button>
              </td>
            </tr>
          </tbody>
        </n-table>
      </n-spin>
    </n-card>

    <!-- 添加成员弹窗 -->
    <n-modal
      v-model:show="showAddModal"
      preset="dialog"
      title="添加成员"
      positive-text="确定"
      negative-text="取消"
      @positive-click="handleAdd"
      style="width: 420px"
    >
      <n-form ref="formRef" :model="formData" :rules="formRules" label-placement="left" :label-width="80" class="py-4">
        <n-form-item label="用户ID" path="userId">
          <n-input-number v-model:value="formData.userId" placeholder="请输入用户ID" style="width: 100%" />
        </n-form-item>
        <n-form-item label="角色" path="role">
          <n-input v-model:value="formData.role" placeholder="请输入角色" />
        </n-form-item>
      </n-form>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import { ref, reactive, computed, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { useMessage, useDialog } from 'naive-ui';
  import { getMembers, addMember, removeMember } from '@/api/project/index';
  import type { MemberItem } from '@/api/project/index';

  const route = useRoute();
  const message = useMessage();
  const dialog = useDialog();
  const projectId = computed(() => Number(route.params.projectId));

  const loading = ref(false);
  const memberList = ref<MemberItem[]>([]);

  const showAddModal = ref(false);
  const formRef = ref<any>(null);
  const formData = reactive({ userId: null as number | null, role: '' });
  const formRules = {
    userId: { required: true, type: 'number', message: '请输入用户ID', trigger: 'blur' },
    role: { required: true, message: '请输入角色', trigger: 'blur' },
  };

  async function loadMembers() {
    loading.value = true;
    try {
      const res = await getMembers(projectId.value);
      memberList.value = res?.list || [];
    } catch { /* ignore */ } finally { loading.value = false; }
  }

  async function handleAdd() {
    try { await formRef.value?.validate(); } catch { return false; }
    try {
      await addMember(projectId.value, { userId: formData.userId!, role: formData.role });
      message.success('成员添加成功');
      showAddModal.value = false;
      formData.userId = null; formData.role = '';
      loadMembers();
    } catch { message.error('添加失败'); return false; }
  }

  function handleRemove(member: MemberItem) {
    dialog.warning({
      title: '确认移除',
      content: `确定要移除成员「${member.username}」吗？`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        try { await removeMember(projectId.value, member.userId); message.success('移除成功'); loadMembers(); }
        catch { message.error('移除失败'); }
      },
    });
  }

  onMounted(loadMembers);
</script>
