<template>
  <n-grid cols="1" responsive="screen" class="-mt-4">
    <n-grid-item>
      <n-list>
        <n-list-item>
          <template #suffix>
            <n-button type="primary" text @click="showPasswordModal = true">修改</n-button>
          </template>
          <n-thing title="账户密码">
            <template #description>
              <span class="text-gray-400">定期更换密码，保障账户安全</span>
            </template>
          </n-thing>
        </n-list-item>
      </n-list>
    </n-grid-item>

    <n-modal
      v-model:show="showPasswordModal"
      preset="dialog"
      title="修改密码"
      positive-text="确认修改"
      negative-text="取消"
      :loading="submitting"
      @positive-click="handleSubmitPassword"
    >
      <n-form ref="pwdFormRef" :model="pwdForm" :rules="pwdRules" label-placement="left" label-width="80">
        <n-form-item label="旧密码" path="oldPassword">
          <n-input v-model:value="pwdForm.oldPassword" type="password" placeholder="请输入旧密码" show-password-on="click" />
        </n-form-item>
        <n-form-item label="新密码" path="newPassword">
          <n-input v-model:value="pwdForm.newPassword" type="password" placeholder="请输入新密码（6-20位）" show-password-on="click" />
        </n-form-item>
        <n-form-item label="确认密码" path="confirmPassword">
          <n-input v-model:value="pwdForm.confirmPassword" type="password" placeholder="请再次输入新密码" show-password-on="click" />
        </n-form-item>
      </n-form>
    </n-modal>
  </n-grid>
</template>

<script lang="ts" setup>
  import { ref, reactive } from 'vue';
  import { useMessage } from 'naive-ui';
  import { useRouter } from 'vue-router';
  import { changePassword } from '@/api/setting/profile';
  import { useUserStore } from '@/store/modules/user';

  const message = useMessage();
  const router = useRouter();
  const userStore = useUserStore();
  const showPasswordModal = ref(false);
  const submitting = ref(false);
  const pwdFormRef: any = ref(null);

  const pwdForm = reactive({
    oldPassword: '',
    newPassword: '',
    confirmPassword: '',
  });

  const pwdRules = {
    oldPassword: { required: true, message: '请输入旧密码', trigger: 'blur' },
    newPassword: { required: true, message: '请输入新密码', trigger: 'blur' },
    confirmPassword: {
      required: true,
      trigger: 'blur',
      validator: (_rule, value) => {
        if (!value) return new Error('请确认新密码');
        if (value !== pwdForm.newPassword) return new Error('两次输入的密码不一致');
        return true;
      },
    },
  };

  async function handleSubmitPassword() {
    try {
      await pwdFormRef.value?.validate();
    } catch {
      return false;
    }
    submitting.value = true;
    try {
      await changePassword({
        oldPassword: pwdForm.oldPassword,
        newPassword: pwdForm.newPassword,
      });
      // 后端已踢掉全部登录态（含当前会话），清理本地并回登录页重新登录
      message.success('密码修改成功，请重新登录');
      showPasswordModal.value = false;
      pwdForm.oldPassword = '';
      pwdForm.newPassword = '';
      pwdForm.confirmPassword = '';
      await userStore.logout();
      router.replace('/login');
    } catch (e: any) {
      message.error(e.message || '密码修改失败');
      return false;
    } finally {
      submitting.value = false;
    }
  }
</script>
