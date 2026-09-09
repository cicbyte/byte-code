<template>
  <div class="system-basic">
    <n-spin :show="loading">
      <n-form ref="formRef" :model="formValue" :rules="rules" label-placement="top" class="setting-form">
        <!-- 站点信息 -->
        <div class="form-section-title">
          <n-icon size="14" color="var(--primary-color, #16a34a)"><GlobalOutlined /></n-icon>
          站点信息
        </div>
        <div class="form-grid">
          <n-form-item label="网站名称" path="siteName">
            <n-input v-model:value="formValue.siteName" placeholder="显示在登录页和浏览器标题" />
          </n-form-item>
          <n-form-item label="备案编号" path="siteIcp">
            <n-input v-model:value="formValue.siteIcp" placeholder="ICP 备案号（可选）" />
          </n-form-item>
          <n-form-item label="联系电话" path="sitePhone">
            <n-input v-model:value="formValue.sitePhone" placeholder="对外展示的联系电话" />
          </n-form-item>
          <n-form-item label="联系地址" path="siteAddress">
            <n-input v-model:value="formValue.siteAddress" placeholder="城市 / 地址" />
          </n-form-item>
        </div>

        <!-- 访问控制 -->
        <div class="form-section-title mt-6">
          <n-icon size="14" color="var(--primary-color, #16a34a)"><SafetyOutlined /></n-icon>
          访问控制
        </div>
        <div class="toggle-grid">
          <div class="toggle-card">
            <div class="toggle-info">
              <div class="toggle-label">登录验证码</div>
              <div class="toggle-desc">开启后登录页需要输入图形验证码</div>
            </div>
            <n-switch v-model:value="formValue.loginCaptcha" :checked-value="1" :unchecked-value="0" />
          </div>
          <div class="toggle-card" :class="{ danger: !formValue.siteOpen }">
            <div class="toggle-info">
              <div class="toggle-label">网站开启访问</div>
              <div class="toggle-desc">{{ formValue.siteOpen ? '站点正常运行中' : '站点已关闭，外部无法访问' }}</div>
            </div>
            <n-switch v-model:value="formValue.siteOpen" @update:value="systemOpenChange" />
          </div>
        </div>

        <n-form-item v-if="!formValue.siteOpen" label="网站关闭提示" path="siteCloseText" class="mt-4">
          <n-input
            v-model:value="formValue.siteCloseText"
            type="textarea"
            placeholder="站点关闭时向访客展示的提示文案"
            :rows="2"
          />
        </n-form-item>

        <div class="form-actions">
          <n-button type="primary" :loading="submitting" @click="formSubmit">保存设置</n-button>
        </div>
      </n-form>
    </n-spin>
  </div>
</template>

<script lang="ts" setup>
  import { ref, onMounted } from 'vue';
  import { useDialog, useMessage } from 'naive-ui';
  import { GlobalOutlined, SafetyOutlined } from '@vicons/antd';
  import { getSystemConfig, updateSystemConfig } from '@/api/setting/system';

  const rules = {
    siteName: { required: true, message: '请输入网站名称', trigger: 'blur' },
  };

  const formRef: any = ref(null);
  const message = useMessage();
  const dialog = useDialog();
  const loading = ref(false);
  const submitting = ref(false);

  const formValue = ref({
    siteName: '',
    sitePhone: '',
    siteIcp: '',
    siteAddress: '',
    loginCaptcha: 0,
    siteCloseText: '',
    siteOpen: true,
  });

  onMounted(async () => {
    loading.value = true;
    try {
      const data: any = await getSystemConfig();
      if (data) {
        formValue.value = {
          siteName: data.siteName || '',
          sitePhone: data.sitePhone || '',
          siteIcp: data.siteIcp || '',
          siteAddress: data.siteAddress || '',
          loginCaptcha: data.loginCaptcha ?? 0,
          siteCloseText: data.siteCloseText || '',
          siteOpen: data.siteOpen !== false,
        };
      }
    } catch {
      message.error('获取系统配置失败');
    } finally {
      loading.value = false;
    }
  });

  function systemOpenChange(value: boolean) {
    if (!value) {
      dialog.warning({
        title: '关闭网站访问',
        content: '关闭后外部将无法访问站点（管理员仍可登录），确定关闭吗？',
        positiveText: '确定关闭',
        negativeText: '取消',
        onNegativeClick: () => {
          formValue.value.siteOpen = true;
        },
      });
    }
  }

  async function formSubmit() {
    formRef.value.validate(async (errors) => {
      if (!errors) {
        submitting.value = true;
        try {
          await updateSystemConfig(formValue.value);
          message.success('更新成功');
        } catch {
          message.error('更新失败');
        } finally {
          submitting.value = false;
        }
      } else {
        message.error('请填写完整信息');
      }
    });
  }
</script>

<style lang="less" scoped>
  .system-basic {
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

  .toggle-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;

    @media (max-width: 640px) {
      grid-template-columns: 1fr;
    }
  }

  .toggle-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 16px;
    border: 1px solid rgba(0, 0, 0, 0.06);
    border-radius: 10px;
    transition: border-color 0.15s;

    &:hover {
      border-color: var(--primary-color, #16a34a);
    }

    &.danger {
      border-color: #e74c3c;
      background: rgba(231, 76, 60, 0.03);

      .toggle-label {
        color: #e74c3c;
      }
    }
  }

  .toggle-label {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-1, #1f2328);
  }

  .toggle-desc {
    font-size: 12px;
    color: var(--text-3, #8b949e);
    margin-top: 3px;
  }

  .setting-form {
    :deep(.n-form-item-label) {
      font-size: 13px;
      font-weight: 500;
      color: var(--text-2, #57606a);
    }
  }

  .form-actions {
    margin-top: 28px;
    padding-top: 16px;
    border-top: 1px solid rgba(0, 0, 0, 0.04);
  }
</style>
