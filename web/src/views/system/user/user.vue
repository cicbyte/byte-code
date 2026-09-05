<template>
  <div>
    <n-card :bordered="false" title="用户管理" class="proCard">
      <template #header-extra>
        <n-button type="primary" @click="openCreate">新建用户</n-button>
      </template>

      <n-space class="mb-4" align="center">
        <n-input
          v-model:value="keyword"
          placeholder="用户名 / 姓名"
          clearable
          style="width: 220px"
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        />
        <n-button type="primary" @click="handleSearch">查询</n-button>
        <n-button @click="handleReset">重置</n-button>
      </n-space>

      <n-spin :show="loading">
        <EmptyState v-if="!loading && users.length === 0" type="member" title="暂无用户" description="邀请新用户后将在这里展示" compact />
        <n-table v-else :bordered="false" :single-line="false" size="small">
          <thead>
            <tr>
              <th>用户名</th>
              <th>姓名</th>
              <th>邮箱</th>
              <th>状态</th>
              <th>创建时间</th>
              <th style="width: 160px">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in users" :key="item.id">
              <td class="font-medium">{{ item.username }}</td>
              <td>{{ item.realName || '-' }}</td>
              <td>{{ item.email || '-' }}</td>
              <td>
                <n-tag :type="USER_STATUS.tagType(item.status)" size="small">
                  {{ USER_STATUS.label(item.status) }}
                </n-tag>
              </td>
              <td class="text-xs text-gray-400">{{ (item.createdAt || '').slice(0, 10) }}</td>
              <td>
                <n-space size="small">
                  <n-button text type="primary" size="small" @click="openEdit(item)">编辑</n-button>
                  <n-button text type="warning" size="small" @click="openResetPwd(item)">重置密码</n-button>
                  <n-button
                    v-if="item.id !== 1"
                    text
                    :type="item.status === 1 ? 'error' : 'success'"
                    size="small"
                    @click="toggleStatus(item)"
                  >
                    {{ item.status === 1 ? '禁用' : '启用' }}
                  </n-button>
                  <n-button
                    v-if="item.id !== 1"
                    text
                    type="error"
                    size="small"
                    @click="handleDelete(item)"
                  >
                    删除
                  </n-button>
                </n-space>
              </td>
            </tr>
          </tbody>
        </n-table>

        <div class="mt-4 flex justify-end">
          <n-pagination
            v-model:page="pagination.page"
            :page-size="pagination.size"
            :item-count="total"
            @update:page="loadUsers"
          />
        </div>
      </n-spin>
    </n-card>

    <!-- 新建/编辑弹窗 -->
    <n-modal v-model:show="showModal" :title="isEdit ? '编辑用户' : '新建用户'" preset="card" style="width: 480px">
      <n-form ref="formRef" :model="form" :rules="formRules" label-placement="left" label-width="80">
        <n-form-item v-if="!isEdit" label="用户名" path="username">
          <n-input v-model:value="form.username" placeholder="3-30位，创建后不可修改" :disabled="isEdit" />
        </n-form-item>
        <n-form-item v-if="!isEdit" label="密码" path="password">
          <n-input
            v-model:value="form.password"
            type="password"
            placeholder="8-20位，须包含字母和数字"
            show-password-on="click"
          />
        </n-form-item>
        <n-form-item label="姓名" path="realName">
          <n-input v-model:value="form.realName" placeholder="请输入姓名" />
        </n-form-item>
        <n-form-item label="邮箱" path="email">
          <n-input v-model:value="form.email" placeholder="请输入邮箱" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showModal = false">取消</n-button>
          <n-button type="primary" :loading="submitting" @click="handleSubmit">
            {{ isEdit ? '保存' : '创建' }}
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 重置密码弹窗 -->
    <n-modal v-model:show="showResetPwd" title="重置密码" preset="card" style="width: 420px">
      <n-alert type="warning" :bordered="false" class="mb-4">
        重置后该用户的所有登录会话将被踢出，需使用新密码重新登录。
      </n-alert>
      <n-form label-placement="left" label-width="80">
        <n-form-item label="新密码">
          <n-input
            v-model:value="newPassword"
            type="password"
            placeholder="8-20位，须包含字母和数字"
            show-password-on="click"
          />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showResetPwd = false">取消</n-button>
          <n-button type="primary" :loading="submitting" @click="handleResetPwd">确认重置</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { USER_STATUS } from '@/enums/entities';
  import { ref, reactive, onMounted } from 'vue';
  import { useMessage, useDialog } from 'naive-ui';
  import {
    getUserList,
    createUser,
    updateUser,
    resetUserPassword,
    deleteUser,
  } from '@/api/system/userManage';
  import type { UserItem } from '@/api/system/userManage';

  const message = useMessage();
  const dialog = useDialog();

  const loading = ref(false);
  const submitting = ref(false);
  const users = ref<UserItem[]>([]);
  const total = ref(0);
  const keyword = ref('');
  const pagination = reactive({ page: 1, size: 20 });

  const showModal = ref(false);
  const isEdit = ref(false);
  const editId = ref<number | null>(null);
  const formRef = ref();
  const form = ref({ username: '', password: '', realName: '', email: '' });

  const showResetPwd = ref(false);
  const resetPwdId = ref<number | null>(null);
  const newPassword = ref('');

  const formRules = {
    username: [
      { required: true, message: '请输入用户名', trigger: 'blur' },
      { min: 3, max: 30, message: '用户名长度3-30位', trigger: 'blur' },
    ],
    password: [
      { required: true, message: '请输入密码', trigger: 'blur' },
      {
        validator: (_r: unknown, v: string) =>
          v.length >= 8 && v.length <= 20 && /[A-Za-z]/.test(v) && /\d/.test(v),
        message: '密码8-20位且须包含字母和数字',
        trigger: 'blur',
      },
    ],
  };

  async function loadUsers() {
    loading.value = true;
    try {
      const res = await getUserList({
        keyword: keyword.value || undefined,
        page: pagination.page,
        size: pagination.size,
      });
      users.value = res?.list || [];
      total.value = res?.total || 0;
    } catch {
      // ignore
    } finally {
      loading.value = false;
    }
  }

  function handleSearch() {
    pagination.page = 1;
    loadUsers();
  }

  function handleReset() {
    keyword.value = '';
    pagination.page = 1;
    loadUsers();
  }

  function openCreate() {
    isEdit.value = false;
    editId.value = null;
    form.value = { username: '', password: '', realName: '', email: '' };
    showModal.value = true;
  }

  function openEdit(item: UserItem) {
    isEdit.value = true;
    editId.value = item.id;
    form.value = {
      username: item.username,
      password: '',
      realName: item.realName || '',
      email: item.email || '',
    };
    showModal.value = true;
  }

  async function handleSubmit() {
    if (!isEdit.value) {
      try {
        await formRef.value?.validate();
      } catch {
        return;
      }
    }
    submitting.value = true;
    try {
      if (isEdit.value && editId.value) {
        const data: Record<string, string> = {};
        if (form.value.realName !== '') data.realName = form.value.realName;
        if (form.value.email !== '') data.email = form.value.email;
        await updateUser(editId.value, data);
        message.success('更新成功');
      } else {
        await createUser({
          username: form.value.username,
          password: form.value.password,
          realName: form.value.realName || undefined,
          email: form.value.email || undefined,
        });
        message.success('创建成功');
      }
      showModal.value = false;
      loadUsers();
    } catch (e: any) {
      message.error(e.message || (isEdit.value ? '更新失败' : '创建失败'));
    } finally {
      submitting.value = false;
    }
  }

  function openResetPwd(item: UserItem) {
    resetPwdId.value = item.id;
    newPassword.value = '';
    showResetPwd.value = true;
  }

  async function handleResetPwd() {
    if (!newPassword.value || newPassword.value.length < 8 || !/[A-Za-z]/.test(newPassword.value) || !/\d/.test(newPassword.value)) {
      message.warning('密码须8-20位且包含字母和数字');
      return;
    }
    submitting.value = true;
    try {
      await resetUserPassword(resetPwdId.value!, newPassword.value);
      message.success('密码重置成功');
      showResetPwd.value = false;
    } catch (e: any) {
      message.error(e.message || '重置失败');
    } finally {
      submitting.value = false;
    }
  }

  async function toggleStatus(item: UserItem) {
    const newStatus = item.status === 1 ? 0 : 1;
    try {
      await updateUser(item.id, { status: newStatus });
      message.success(newStatus === 1 ? '已启用' : '已禁用');
      loadUsers();
    } catch (e: any) {
      message.error(e.message || '操作失败');
    }
  }

  function handleDelete(item: UserItem) {
    dialog.warning({
      title: '删除用户',
      content: `确定要删除「${item.username}」吗？该用户的项目成员身份和登录会话将一并清除。`,
      positiveText: '删除',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await deleteUser(item.id);
          message.success('删除成功');
          loadUsers();
        } catch (e: any) {
          message.error(e.message || '删除失败');
        }
      },
    });
  }

  onMounted(loadUsers);
</script>
