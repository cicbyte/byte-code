<template>
  <div class="basic-setting">
    <!-- 头像区：大卡片突出展示 -->
    <div class="profile-hero">
      <div class="profile-avatar-wrap">
        <n-avatar round :size="88" :src="avatarUrl || undefined" class="profile-avatar">
          <template #icon>
            <n-icon size="42"><UserOutlined /></n-icon>
          </template>
        </n-avatar>
        <div class="profile-avatar-info">
          <div class="profile-name">{{ formValue.nickname || userStore.getUserInfo?.username || '—' }}</div>
          <div class="profile-sub">{{ userStore.getUserInfo?.username || '' }}</div>
          <n-upload
            accept="image/png,image/jpeg,image/gif,image/webp"
            :show-file-list="false"
            :custom-request="pickImage"
            class="mt-2"
          >
            <n-button size="small" ghost type="primary">更换头像</n-button>
          </n-upload>
        </div>
      </div>
      <div class="profile-hint">
        <n-icon size="14" color="var(--text-3)"><InfoCircleOutlined /></n-icon>
        <span>支持 png / jpg / gif / webp，选图后可框选裁剪，输出 256×256</span>
      </div>
    </div>

    <n-divider class="setting-divider" />

    <!-- 表单区：两列布局充分利用宽度 -->
    <n-spin :show="loading">
      <n-form
        ref="formRef"
        :model="formValue"
        :rules="rules"
        label-placement="top"
        label-width="auto"
        class="setting-form"
      >
        <div class="form-section-title">
          <n-icon size="14" color="var(--primary-color, #16a34a)"><IdcardOutlined /></n-icon>
          基本信息
        </div>
        <div class="form-grid">
          <n-form-item label="昵称" path="nickname">
            <n-input v-model:value="formValue.nickname" placeholder="显示名称" />
          </n-form-item>
          <n-form-item label="邮箱" path="email">
            <n-input v-model:value="formValue.email" placeholder="用于接收通知" />
          </n-form-item>
        </div>

        <div class="form-section-title mt-6">
          <n-icon size="14" color="var(--primary-color, #16a34a)"><PhoneOutlined /></n-icon>
          联系方式
        </div>
        <div class="form-grid">
          <n-form-item label="联系电话" path="phone">
            <n-input v-model:value="formValue.phone" placeholder="手机号" />
          </n-form-item>
          <n-form-item label="联系地址" path="address">
            <n-input v-model:value="formValue.address" placeholder="城市 / 地址" />
          </n-form-item>
        </div>

        <div class="form-actions">
          <n-button type="primary" :loading="submitting" @click="formSubmit">保存修改</n-button>
        </div>
      </n-form>
    </n-spin>

    <!-- 裁剪弹窗 -->
    <AvatarCropper v-model:show="showCropper" :file="cropFile" :uploading="uploading" @confirm="doUploadAvatar" />
  </div>
</template>

<script lang="ts" setup>
  import { ref, reactive, onMounted } from 'vue';
  import { useMessage } from 'naive-ui';
  import {
    UserOutlined,
    InfoCircleOutlined,
    IdcardOutlined,
    PhoneOutlined,
  } from '@vicons/antd';
  import type { UploadCustomRequestOptions } from 'naive-ui';
  import { getProfile, updateProfile, uploadAvatar } from '@/api/setting/profile';
  import { useUserStore } from '@/store/modules/user';
  import AvatarCropper from './AvatarCropper.vue';

  const message = useMessage();
  const userStore = useUserStore();
  const loading = ref(false);
  const submitting = ref(false);
  const uploading = ref(false);
  const avatarUrl = ref('');

  const formRef = ref<any>(null);
  const formValue = ref({ nickname: '', phone: '', email: '', address: '' });
  const rules = {
    nickname: { required: true, message: '请输入昵称', trigger: 'blur' },
  };

  // 裁剪
  const showCropper = ref(false);
  const cropFile = ref<File | null>(null);

  async function pickImage({ file }: UploadCustomRequestOptions) {
    const raw = file.file as File;
    if (!raw.type.startsWith('image/')) {
      message.error('请选择图片文件');
      return;
    }
    cropFile.value = raw;
    showCropper.value = true;
  }

  async function doUploadAvatar(cropped: File) {
    uploading.value = true;
    try {
      const res = await uploadAvatar(cropped);
      avatarUrl.value = res?.avatar || avatarUrl.value;
      if (userStore.info) {
        userStore.info = { ...userStore.info, avatar: avatarUrl.value };
      }
      userStore.setAvatar(avatarUrl.value);
      message.success('头像已更新');
      showCropper.value = false;
    } catch (e: any) {
      message.error(e.message || '头像上传失败');
    } finally {
      uploading.value = false;
    }
  }

  onMounted(async () => {
    loading.value = true;
    try {
      const data: any = await getProfile();
      if (data) {
        avatarUrl.value = data.avatar || '';
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
    formRef.value.validate(async (errors: any) => {
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

<style lang="less" scoped>
  .basic-setting {
    width: 100%;
  }

  .profile-hero {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 4px 16px;
  }

  .profile-avatar-wrap {
    display: flex;
    align-items: center;
    gap: 20px;
  }

  .profile-avatar {
    border: 3px solid rgba(22, 163, 74, 0.15);
    box-shadow: 0 2px 12px rgba(22, 163, 74, 0.1);
  }

  .profile-avatar-info {
    .profile-name {
      font-size: 20px;
      font-weight: 700;
      color: var(--text-1, #1f2328);
      line-height: 1.3;
    }

    .profile-sub {
      font-size: 13px;
      color: var(--text-3, #8b949e);
      margin-top: 2px;
    }
  }

  .profile-hint {
    display: flex;
    align-items: center;
    gap: 5px;
    font-size: 12px;
    color: var(--text-3, #8b949e);
  }

  .setting-divider {
    margin: 4px 0 20px;
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

  .form-actions {
    margin-top: 28px;
    padding-top: 16px;
    border-top: 1px solid rgba(0, 0, 0, 0.04);
  }
</style>
