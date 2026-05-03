<template>
  <n-grid cols="2 s:2 m:2 l:3 xl:3 2xl:3" responsive="screen">
    <n-grid-item>
      <n-spin :show="loading">
        <n-form :label-width="80" :model="formValue" :rules="rules" ref="formRef">
          <n-form-item label="昵称" path="nickname">
            <n-input v-model:value="formValue.nickname" placeholder="请输入昵称" />
          </n-form-item>

          <n-form-item label="邮箱" path="email">
            <n-input placeholder="请输入邮箱" v-model:value="formValue.email" />
          </n-form-item>

          <n-form-item label="联系电话" path="phone">
            <n-input placeholder="请输入联系电话" v-model:value="formValue.phone" />
          </n-form-item>

          <n-form-item label="联系地址" path="address">
            <n-input v-model:value="formValue.address" type="textarea" placeholder="请输入联系地址" />
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
  import { useMessage } from 'naive-ui';
  import { getProfile, updateProfile } from '@/api/setting/profile';

  const rules = {
    nickname: {
      required: true,
      message: '请输入昵称',
      trigger: 'blur',
    },
  };

  const formRef: any = ref(null);
  const message = useMessage();
  const loading = ref(false);
  const submitting = ref(false);

  const formValue = ref({
    nickname: '',
    phone: '',
    email: '',
    address: '',
  });

  onMounted(async () => {
    loading.value = true;
    try {
      const data: any = await getProfile();
      if (data) {
        formValue.value = {
          nickname: data.nickname || '',
          phone: data.phone || '',
          email: data.email || '',
          address: data.address || '',
        };
      }
    } catch (e) {
      message.error('获取个人信息失败');
    } finally {
      loading.value = false;
    }
  });

  async function formSubmit() {
    formRef.value.validate(async (errors) => {
      if (!errors) {
        submitting.value = true;
        try {
          await updateProfile(formValue.value);
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
