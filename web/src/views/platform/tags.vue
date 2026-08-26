<template>
  <div>
    <n-card :bordered="false" title="标签管理" class="proCard">
      <template #header-extra>
        <n-button type="primary" @click="handleCreate">
          <template #icon>
            <n-icon><PlusOutlined /></n-icon>
          </template>
          新建标签
        </n-button>
      </template>

      <n-spin :show="loading">
        <n-empty v-if="!loading && tagList.length === 0" description="暂无标签" />
        <n-space v-else>
          <n-tag
            v-for="item in tagList"
            :key="item.id"
            :color="{ color: item.color, textColor: getTextColor(item.color) }"
            closable
            @close="handleDelete(item)"
            @click="handleEdit(item)"
            style="cursor: pointer"
          >
            {{ item.name }}
          </n-tag>
        </n-space>
      </n-spin>
    </n-card>

    <!-- 新建/编辑弹窗 -->
    <n-modal
      v-model:show="showModal"
      preset="dialog"
      :title="isEdit ? '编辑标签' : '新建标签'"
      positive-text="确定"
      negative-text="取消"
      @positive-click="handleSubmit"
      style="width: 420px"
    >
      <n-form ref="formRef" :model="formData" :rules="formRules" label-placement="left" :label-width="80" class="py-4">
        <n-form-item label="名称" path="name">
          <n-input v-model:value="formData.name" placeholder="请输入标签名称" />
        </n-form-item>
        <n-form-item label="颜色" path="color">
          <n-color-picker v-model:value="formData.color" :modes="['hex']" />
        </n-form-item>
      </n-form>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import { ref, reactive, onMounted } from 'vue';
  import { useMessage, useDialog } from 'naive-ui';
  import { PlusOutlined } from '@vicons/antd';
  import { getTags, createTag, updateTag, deleteTag } from '@/api/platform/index';
  import type { TagItem } from '@/api/platform/index';

  const message = useMessage();
  const dialog = useDialog();
  const loading = ref(false);
  const tagList = ref<TagItem[]>([]);
  const showModal = ref(false);
  const isEdit = ref(false);
  const editId = ref<number | null>(null);
  const formRef = ref<any>(null);

  const formData = reactive({ name: '', color: '#1890ff' });
  const formRules = {
    name: { required: true, message: '请输入标签名称', trigger: 'blur' },
  };

  function getTextColor(bgColor: string): string {
    if (!bgColor || bgColor.length < 7) return '#fff';
    const hex = bgColor.replace('#', '');
    const r = parseInt(hex.substring(0, 2), 16);
    const g = parseInt(hex.substring(2, 4), 16);
    const b = parseInt(hex.substring(4, 6), 16);
    const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255;
    return luminance > 0.5 ? '#000' : '#fff';
  }

  async function loadData() {
    loading.value = true;
    try {
      const res = await getTags();
      tagList.value = res?.list || [];
    } catch (e) {
      // ignore
    } finally {
      loading.value = false;
    }
  }

  function handleCreate() {
    formData.name = '';
    formData.color = '#1890ff';
    isEdit.value = false;
    editId.value = null;
    showModal.value = true;
  }

  function handleEdit(item: TagItem) {
    isEdit.value = true;
    editId.value = item.id;
    formData.name = item.name;
    formData.color = item.color;
    showModal.value = true;
  }

  function handleDelete(item: TagItem) {
    dialog.warning({
      title: '确认删除',
      content: `确定要删除标签「${item.name}」吗？`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await deleteTag(item.id);
          message.success('删除成功');
          loadData();
        } catch (e) {
          message.error('删除失败');
        }
      },
    });
  }

  async function handleSubmit() {
    try {
      await formRef.value?.validate();
    } catch {
      return false;
    }
    try {
      if (isEdit.value && editId.value) {
        await updateTag(editId.value, { ...formData });
        message.success('更新成功');
      } else {
        await createTag({ ...formData });
        message.success('创建成功');
      }
      showModal.value = false;
      loadData();
    } catch (e) {
      message.error('操作失败');
      return false;
    }
  }

  onMounted(() => {
    loadData();
  });
</script>
