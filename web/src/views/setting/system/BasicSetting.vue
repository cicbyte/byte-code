<template>
  <n-grid cols="2 s:2 m:2 l:3 xl:3 2xl:3" responsive="screen">
    <n-grid-item>
      <n-spin :show="loading">
        <n-form :label-width="80" :model="formValue" :rules="rules" ref="formRef">
          <n-form-item label="网站名称" path="siteName">
            <n-input v-model:value="formValue.siteName" placeholder="请输入网站名称" />
          </n-form-item>

          <n-form-item label="备案编号" path="siteIcp">
            <n-input placeholder="请输入备案编号" v-model:value="formValue.siteIcp" />
          </n-form-item>

          <n-form-item label="联系电话" path="sitePhone">
            <n-input placeholder="请输入联系电话" v-model:value="formValue.sitePhone" />
          </n-form-item>

          <n-form-item label="联系地址" path="siteAddress">
            <n-input v-model:value="formValue.siteAddress" type="textarea" placeholder="请输入联系地址" />
          </n-form-item>

          <n-form-item label="登录验证码" path="loginCaptcha">
            <n-radio-group v-model:value="formValue.loginCaptcha" name="loginCaptcha">
              <n-space>
                <n-radio :value="1">开启</n-radio>
                <n-radio :value="0">关闭</n-radio>
              </n-space>
            </n-radio-group>
          </n-form-item>

          <n-form-item label="网站开启访问" path="siteOpen">
            <n-switch
              size="large"
              v-model:value="formValue.siteOpen"
              @update:value="systemOpenChange"
            />
          </n-form-item>

          <n-form-item label="网站关闭提示" path="siteCloseText">
            <n-input
              v-model:value="formValue.siteCloseText"
              type="textarea"
              placeholder="请输入网站关闭提示"
            />
          </n-form-item>

          <div>
            <n-space>
              <n-button type="primary" @click="formSubmit" :loading="submitting">更新基本信息</n-button>
            </n-space>
          </div>
        </n-form>
      </n-spin>
    </n-grid-item>
  </n-grid>
</template>

<script lang="ts" setup>
  import { ref, onMounted } from 'vue';
  import { useDialog, useMessage } from 'naive-ui';
  import { getSystemConfig, updateSystemConfig } from '@/api/setting/system';

  const rules = {
    siteName: {
      required: true,
      message: '请输入网站名称',
      trigger: 'blur',
    },
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
    } catch (e) {
      message.error('获取系统配置失败');
    } finally {
      loading.value = false;
    }
  });

  function systemOpenChange(value) {
    if (!value) {
      dialog.warning({
        title: '提示',
        content: '您确定要关闭系统访问吗？该操作立马生效，请慎重操作！',
        positiveText: '确定',
        negativeText: '取消',
        onPositiveClick: () => {
          message.success('操作成功');
        },
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
        } catch (e) {
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
