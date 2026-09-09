<template>
  <div class="email-setting">
    <n-spin :show="loading">
      <n-form ref="formRef" :model="formValue" :rules="rules" label-placement="top" class="setting-form">
        <!-- 服务器配置 -->
        <div class="form-section-title">
          <n-icon size="14" color="var(--primary-color, #16a34a)"><CloudServerOutlined /></n-icon>
          SMTP 服务器
        </div>
        <div class="form-grid">
          <n-form-item label="服务器地址" path="smtpHost">
            <n-input v-model:value="formValue.smtpHost" placeholder="如 smtp.example.com" />
          </n-form-item>
          <n-form-item label="端口" path="smtpPort">
            <n-input v-model:value="formValue.smtpPort" placeholder="通常 465（SSL）或 587（TLS）" />
          </n-form-item>
        </div>

        <!-- 认证与发件人 -->
        <div class="form-section-title mt-6">
          <n-icon size="14" color="var(--primary-color, #16a34a)"><MailOutlined /></n-icon>
          认证与发件人
        </div>
        <div class="form-grid">
          <n-form-item label="用户名" path="smtpUser">
            <n-input v-model:value="formValue.smtpUser" placeholder="SMTP 账号" />
          </n-form-item>
          <n-form-item label="密码 / 授权码" path="smtpPass">
            <n-input
              v-model:value="formValue.smtpPass"
              type="password"
              placeholder="不修改请留空"
              show-password-on="click"
            />
          </n-form-item>
          <n-form-item label="发件人邮箱" path="smtpFrom" :span="2">
            <n-input v-model:value="formValue.smtpFrom" placeholder="显示在邮件发件人字段" />
          </n-form-item>
        </div>

        <!-- 发送测试 -->
        <div class="form-section-title mt-6">
          <n-icon size="14" color="var(--primary-color, #16a34a)"><SendOutlined /></n-icon>
          发送测试
        </div>
        <div class="test-row">
          <n-input
            v-model:value="testTo"
            placeholder="收件邮箱（如 you@example.com）"
            size="small"
            style="flex: 1"
            @keyup.enter="handleTest"
          />
          <n-button size="small" ghost type="primary" :loading="testing" :disabled="!testTo.trim()" @click="handleTest">
            发送测试邮件
          </n-button>
        </div>
        <n-alert v-if="testResult" :type="testResult.ok ? 'success' : 'error'" :bordered="false" class="mt-3">
          {{ testResult.msg }}
        </n-alert>

        <div class="form-actions">
          <n-button type="primary" :loading="submitting" @click="formSubmit">保存邮件配置</n-button>
        </div>
      </n-form>
    </n-spin>
  </div>
</template>

<script lang="ts" setup>
  import { ref, onMounted } from 'vue';
  import { useMessage } from 'naive-ui';
  import { CloudServerOutlined, MailOutlined, SendOutlined } from '@vicons/antd';
  import { getSystemConfig, updateSystemConfig, sendTestMail } from '@/api/setting/system';

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
    } catch {
      message.error('获取邮件配置失败');
    } finally {
      loading.value = false;
    }
  });

  // 测试邮件
  const testTo = ref('');
  const testing = ref(false);
  const testResult = ref<{ ok: boolean; msg: string } | null>(null);

  async function handleTest() {
    if (!testTo.value.trim()) return;
    testing.value = true;
    testResult.value = null;
    try {
      await sendTestMail(testTo.value.trim());
      testResult.value = { ok: true, msg: `测试邮件已发送至 ${testTo.value}，请查收（注意检查垃圾箱）` };
    } catch (e: any) {
      testResult.value = { ok: false, msg: e?.message || '发送失败' };
    } finally {
      testing.value = false;
    }
  }

  async function formSubmit() {
    submitting.value = true;
    try {
      await updateSystemConfig(formValue.value);
      message.success('更新成功');
    } catch {
      message.error('更新失败');
    } finally {
      submitting.value = false;
    }
  }
</script>

<style lang="less" scoped>
  .email-setting {
    width: 100%;
  }

  .form-section-title {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 14px;
    font-weight: 600;
    color: var(--text-1, #1f2328);
    margin-bottom: 12px;

    &.mt-6 {
      margin-top: 24px;
    }
  }

  .form-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0 20px;

    @media (max-width: 640px) {
      grid-template-columns: 1fr;
    }
  }

  .setting-form {
    :deep(.n-form-item-label) {
      font-size: 13px;
      font-weight: 500;
      color: var(--text-2, #57606a);
    }
  }

  .test-row {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .form-actions {
    margin-top: 28px;
    padding-top: 16px;
    border-top: 1px solid rgba(0, 0, 0, 0.04);
  }
</style>
