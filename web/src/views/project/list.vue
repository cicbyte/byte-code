<template>
  <div>
    <div class="n-layout-page-header">
      <n-card :bordered="false" title="项目列表" />
    </div>

    <n-card :bordered="false" class="mt-4 proCard">
      <template #header-extra>
        <n-button type="primary" @click="handleCreate">
          <template #icon>
            <n-icon><PlusOutlined /></n-icon>
          </template>
          新建项目
        </n-button>
      </template>

      <n-spin :show="loading">
        <n-empty v-if="!loading && projectList.length === 0" description="暂无项目" />
        <n-grid v-else cols="1 s:2 m:2 l:3 xl:4 2xl:4" responsive="screen" :x-gap="12" :y-gap="12">
          <n-grid-item v-for="item in projectList" :key="item.id">
            <n-card hoverable size="small" @click="handleDetail(item)" style="cursor: pointer">
              <template #header>
                <span class="font-medium">{{ item.name }}</span>
              </template>
              <template #header-extra>
                <n-tag :type="item.status === 1 ? 'success' : 'default'" size="small">
                  {{ item.status === 1 ? '进行中' : '已结束' }}
                </n-tag>
              </template>
              <p class="text-gray-500 text-sm line-clamp-2">{{ item.description || '暂无描述' }}</p>
              <template #footer>
                <n-space justify="space-between" align="center">
                  <span class="text-gray-400 text-xs">负责人：{{ item.creatorName }}</span>
                  <n-space size="small">
                    <n-button text type="info" size="small" @click.stop="handleEdit(item)">编辑</n-button>
                    <n-button text type="error" size="small" @click.stop="handleDelete(item)">删除</n-button>
                  </n-space>
                </n-space>
              </template>
            </n-card>
          </n-grid-item>
        </n-grid>
      </n-spin>

      <div class="mt-4 flex justify-end" v-if="total > pagination.size">
        <n-pagination
          v-model:page="pagination.page"
          v-model:page-size="pagination.size"
          :item-count="total"
          @update:page="loadData"
        />
      </div>
    </n-card>

    <!-- 新建/编辑弹窗 -->
    <n-modal
      v-model:show="showModal"
      preset="dialog"
      :title="isEdit ? '编辑项目' : '新建项目'"
      positive-text="确定"
      negative-text="取消"
      @positive-click="handleSubmit"
      style="width: 520px"
    >
      <n-form ref="formRef" :model="formData" :rules="formRules" label-placement="left" :label-width="80" class="py-4">
        <n-form-item label="名称" path="name">
          <n-input v-model:value="formData.name" placeholder="请输入项目名称" />
        </n-form-item>
        <n-form-item label="描述" path="description">
          <n-input v-model:value="formData.description" type="textarea" placeholder="请输入项目描述" :rows="3" />
        </n-form-item>
      </n-form>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import { ref, reactive, onMounted } from 'vue';
  import { useRouter } from 'vue-router';
  import { useMessage, useDialog } from 'naive-ui';
  import { PlusOutlined } from '@vicons/antd';
  import {
    getProjects,
    createProject,
    updateProject,
    deleteProject,
  } from '@/api/project/index';
  import type { ProjectItem } from '@/api/project/index';

  const router = useRouter();
  const message = useMessage();
  const dialog = useDialog();
  const loading = ref(false);
  const projectList = ref<ProjectItem[]>([]);
  const total = ref(0);
  const showModal = ref(false);
  const isEdit = ref(false);
  const editId = ref<number | null>(null);
  const formRef = ref<any>(null);

  const pagination = reactive({ page: 1, size: 12 });

  const formData = reactive({
    name: '',
    description: '',
  });

  const formRules = {
    name: { required: true, message: '请输入项目名称', trigger: 'blur' },
  };

  function resetForm() {
    formData.name = '';
    formData.description = '';
    isEdit.value = false;
    editId.value = null;
  }

  async function loadData() {
    loading.value = true;
    try {
      const res = await getProjects({ page: pagination.page, size: pagination.size });
      if (res) {
        projectList.value = res.list || [];
        total.value = res.total || 0;
      }
    } catch (e) {
      // ignore
    } finally {
      loading.value = false;
    }
  }

  function handleDetail(item: ProjectItem) {
    router.push(`/project/detail/${item.id}`);
  }

  function handleCreate() {
    resetForm();
    showModal.value = true;
  }

  function handleEdit(item: ProjectItem) {
    isEdit.value = true;
    editId.value = item.id;
    formData.name = item.name;
    formData.description = item.description;
    showModal.value = true;
  }

  function handleDelete(item: ProjectItem) {
    dialog.warning({
      title: '确认删除',
      content: `确定要删除项目「${item.name}」吗？`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await deleteProject(item.id);
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
        await updateProject(editId.value, { ...formData });
        message.success('更新成功');
      } else {
        await createProject({ ...formData });
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
