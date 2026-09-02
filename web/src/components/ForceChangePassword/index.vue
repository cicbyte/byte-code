<template>
  <n-modal
    v-model:show="visible"
    :mask-closable="false"
    :close-on-esc="false"
    preset="card"
    title="首次登录请修改初始密码"
    style="width: 420px"
  >
    <template #header-extra>
      <n-tag type="warning" size="small">必须完成</n-tag>
    </template>

    <n-alert type="info" :bordered="false" class="mb-4">
      当前账号仍在使用初始默认密码，修改后才能使用系统功能。修改成功后需要重新登录。
    </n-alert>

    <n-form ref="formRef" :model="form" :rules="rules" label-placement="top" size="medium">
      <n-form-item label="新密码" path="newPassword">
        <n-input
          v-model:value="form.newPassword"
          type="password"
          placeholder="8-20位，须包含字母和数字"
          show-password-on="click"
        />
      </n-form-item>
      <n-form-item label="确认新密码" path="confirmPassword">
        <n-input
          v-model:value="form.confirmPassword"
          type="password"
          placeholder="请再次输入新密码"
          show-password-on="click"
          @keyup.enter="handleSubmit"
        />
      </n-form-item>
    </n-form>

    <template #action>
      <n-button type="primary" block :loading="submitting" @click="handleSubmit">
        确认修改并重新登录
      </n-button>
    </template>
  </n-modal>
</template>

<script lang="ts" setup>
  import { reactive, ref } from 'vue';
  import { useRouter } from 'vue-router';
  import { useMessage } from 'naive-ui';
  import { changePassword } from '@/api/setting/profile';
  import { useUserStore } from '@/store/modules/user';
  import { forcePwdVisible } from '@/store/modules/forcePassword';

  const router = useRouter();
  const message = useMessage();
  const userStore = useUserStore();

  // 全局单例显隐状态（拦截器/登录页 import 后置 true）
  const visible = forcePwdVisible;

  const formRef = ref();
  const submitting = ref(false);
  const form = reactive({ newPassword: '', confirmPassword: '' });

  const rules = {
    newPassword: [
      { required: true, message: '请输入新密码', trigger: 'blur' },
      {
        validator: (_r: unknown, v: string) => v.length >= 8 && v.length <= 20,
        message: '密码长度为 8-20 位',
        trigger: 'blur',
      },
      {
        validator: (_r: unknown, v: string) => /[A-Za-z]/.test(v) && /\d/.test(v),
        message: '密码须同时包含字母和数字',
        trigger: 'blur',
      },
    ],
    confirmPassword: {
      required: true,
      trigger: 'blur',
      validator: (_r: unknown, v: string) => {
        if (!v) return new Error('请确认新密码');
        if (v !== form.newPassword) return new Error('两次输入的密码不一致');
        return true;
      },
    },
  };

  async function handleSubmit() {
    try {
      await formRef.value?.validate();
    } catch {
      return;
    }
    submitting.value = true;
    try {
      await changePassword({
        oldPassword: '',
        newPassword: form.newPassword,
      });
      visible.value = false;
      message.success('密码修改成功，请使用新密码重新登录');
      // 后端改密会踢掉全部 token，本地清理后回登录页
      await userStore.logout();
      router.replace('/login');
    } catch (e: any) {
      message.error(e.message || '密码修改失败');
    } finally {
      submitting.value = false;
    }
  }
</script>
