<template>
  <div class="reset-page">
    <n-card :bordered="false" class="reset-card" title="设置新密码">
      <n-space vertical :size="12">
        <n-alert v-if="tokenMissing" type="error" :show-icon="false">
          缺少重置令牌：请从邮件中的重置链接进入本页
        </n-alert>
        <template v-else>
          <n-text depth="3" style="font-size: 13px">
            重置链接 30 分钟内有效、仅可使用一次；设置成功后将退出所有已登录会话
          </n-text>
          <n-form ref="formRef" :model="form" :rules="rules" label-placement="top" size="large">
            <n-form-item label="新密码" path="password">
              <n-input v-model:value="form.password" type="password" show-password-on="click" placeholder="8-20 位" @keyup.enter="handleSubmit" />
            </n-form-item>
            <n-form-item label="确认新密码" path="confirm">
              <n-input v-model:value="form.confirm" type="password" show-password-on="click" placeholder="再次输入新密码" @keyup.enter="handleSubmit" />
            </n-form-item>
          </n-form>
          <n-button type="primary" size="large" block :loading="loading" :disabled="!token" @click="handleSubmit">
            重置密码
          </n-button>
        </template>
        <n-button text type="primary" @click="router.push('/login')">返回登录</n-button>
      </n-space>
    </n-card>
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { useMessage } from 'naive-ui';
  import { resetPassword } from '@/api/system/user';

  const route = useRoute();
  const router = useRouter();
  const message = useMessage();

  const token = computed(() => String(route.query.token || ''));
  const tokenMissing = computed(() => !token.value);

  const formRef = ref<any>(null);
  const loading = ref(false);
  const form = reactive({ password: '', confirm: '' });
  const rules = {
    password: [
      { required: true, message: '请输入新密码', trigger: 'blur' },
      { min: 8, max: 20, message: '密码长度 8-20 位', trigger: 'blur' },
    ],
    confirm: [
      { required: true, message: '请再次输入新密码', trigger: 'blur' },
      {
        validator: (_r: any, v: string) => v === form.password,
        message: '两次输入不一致',
        trigger: 'blur',
      },
    ],
  };

  async function handleSubmit() {
    try { await formRef.value?.validate(); } catch { return; }
    loading.value = true;
    try {
      await resetPassword(token.value, form.password);
      message.success('密码已重置，请使用新密码登录');
      router.push('/login');
    } catch (e: any) {
      message.error(e?.message || '重置失败');
    } finally {
      loading.value = false;
    }
  }
</script>

<style scoped lang="less">
  .reset-page {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    background: #f1f5f9;
  }
  .reset-card {
    width: 400px;
    max-width: 92vw;
  }
</style>
