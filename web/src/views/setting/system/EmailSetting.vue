<template>
  <n-grid cols="2 s:2 m:2 l:3 xl:3 2xl:3" responsive="screen">
    <n-grid-item>
      <n-spin :show="loading">
        <n-form :label-width="120" :model="formValue" :rules="rules" ref="formRef">
          <n-form-item label="发件人邮箱" path="smtpFrom">
            <n-input v-model:value="formValue.smtpFrom" placeholder="请输入发件人邮箱" />
          </n-form-item>

          <n-form-item label="SMTP服务器地址" path="smtpHost">
            <n-input v-model:value="formValue.smtpHost" placeholder="请输入SMTP服务器地址" />
          </n-form-item>

          <n-form-item label="SMTP服务器端口" path="smtpPort">
            <n-input v-model:value="formValue.smtpPort" placeholder="请输入SMTP服务器端口" />
          </n-form-item>

          <n-form-item label="SMTP用户名" path="smtpUser">
            <n-input v-model:value="formValue.smtpUser" placeholder="请输入SMTP用户名" />
          </n-form-item>

          <n-form-item label="SMTP密码" path="smtpPass">
            <n-input type="password" v-model:value="formValue.smtpPass" placeholder="不修改请留空" show-password-on="click" />
          </n-form-item>

          <div>
            <n-space>
              <n-button type="primary" @click="formSubmit" :loading="submitting">更新邮件信息</n-button>
            </n-space>
          </div>
        </n-form>
      </n-spin>
    </n-grid-item>
  </n-grid>
</template>

<script lang="ts" setup>
  import { ref, onMounted } from 'vue';
  import { useMessage } from 'naive-ui';
  import { getSystemConfig, updateSystemConfig } from '@/api/setting/system';

  const rules = {};

  const formRef: any = ref(null);
  const message = useMessage();
  const loading = ref(false);
  const submitting = ref(false);

  const formValue = ref({
    smtpFrom: '',
    smtpHost: '',
    smtpPort: '',
    smtpUser: '',
    smtpPass: '',
  });

  onMounted(async () => {
    loading.value = true;
    try {
      const data: any = await getSystemConfig();
      if (data) {
        formValue.value = {
          smtpFrom: data.smtpFrom || '',
          smtpHost: data.smtpHost || '',
          smtpPort: data.smtpPort || '',
          smtpUser: data.smtpUser || '',
          smtpPass: '',
        };
      }
    } catch (e) {
      message.error('获取邮件配置失败');
    } finally {
      loading.value = false;
    }
  });

  async function formSubmit() {
    submitting.value = true;
    try {
      await updateSystemConfig(formValue.value);
      message.success('更新成功');
    } catch (e) {
      message.error('更新失败');
    } finally {
      submitting.value = false;
    }
  }
</script>
