<template>
  <n-grid cols="2 s:2 m:2 l:3 xl:3 2xl:3" responsive="screen">
    <n-grid-item>
      <n-spin :show="loading">
        <!-- 头像 -->
        <div class="avatar-block">
          <n-avatar round :size="72" :src="avatarUrl || undefined">
            <template #icon>
              <n-icon size="34"><UserOutlined /></n-icon>
            </template>
          </n-avatar>
          <div class="avatar-side">
            <n-upload
              accept="image/png,image/jpeg,image/gif,image/webp"
              :show-file-list="false"
              :custom-request="pickImage"
            >
              <n-button size="small">更换头像</n-button>
            </n-upload>
            <span class="avatar-hint">支持 png / jpg / gif / webp，选图后可框选裁剪</span>
          </div>
        </div>

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

  <!-- 框选裁剪 -->
  <AvatarCropper v-model:show="showCropper" :file="cropFile" :uploading="uploading" @confirm="doUploadAvatar" />
</template>

<script lang="ts" setup>
  import { ref, onMounted } from 'vue';
  import { useMessage } from 'naive-ui';
  import { UserOutlined } from '@vicons/antd';
  import type { UploadCustomRequestOptions } from 'naive-ui';
  import { getProfile, updateProfile, uploadAvatar } from '@/api/setting/profile';
  import { useUserStore } from '@/store/modules/user';
  import AvatarCropper from './AvatarCropper.vue';

  const rules = {
    nickname: {
      required: true,
      message: '请输入昵称',
      trigger: 'blur',
    },
  };

  const formRef: any = ref(null);
  const message = useMessage();
  const userStore = useUserStore();
  const loading = ref(false);
  const submitting = ref(false);
  const uploading = ref(false);
  const avatarUrl = ref('');

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

  // 选图：先落裁剪框，确认后才真正上传（裁剪输出 256×256 PNG，天然满足 2MB 限制）
  const showCropper = ref(false);
  const cropFile = ref<File | null>(null);

  async function pickImage({ file, onFinish, onError }: UploadCustomRequestOptions) {
    const raw = file.file as File;
    if (!raw.type.startsWith('image/')) {
      message.error('请选择图片文件');
      onError();
      return;
    }
    cropFile.value = raw;
    showCropper.value = true;
    onFinish();
  }

  async function doUploadAvatar(cropped: File) {
    uploading.value = true;
    try {
      const res = await uploadAvatar(cropped);
      avatarUrl.value = res?.avatar || avatarUrl.value;
      // 同步 user store：Header 的头像 computed 即时刷新
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

<style lang="less" scoped>
  .avatar-block {
    display: flex;
    align-items: center;
    gap: 16px;
    margin-bottom: 24px;
  }

  .avatar-side {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .avatar-hint {
    font-size: 12px;
    color: var(--text-3, #8b949e);
  }
</style>
