<template>
  <div>
    <div class="n-layout-page-header">
      <n-card :bordered="false" title="产品列表" />
    </div>

    <n-card :bordered="false" class="mt-4 proCard">
      <template #header-extra>
        <n-button type="primary" @click="handleCreate">
          <template #icon>
            <n-icon><PlusOutlined /></n-icon>
          </template>
          新建产品
        </n-button>
      </template>

      <n-spin :show="loading">
        <n-empty v-if="!loading && productList.length === 0" description="暂无产品" />
        <n-table v-else :bordered="false" :single-line="false">
          <thead>
            <tr>
              <th>名称</th>
              <th>描述</th>
              <th>负责人</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in productList" :key="item.id">
              <td>{{ item.name }}</td>
              <td>{{ item.description }}</td>
              <td>{{ item.ownerName }}</td>
              <td>
                <n-tag :type="item.status === 'active' ? 'success' : 'default'" size="small">
                  {{ item.status === 'active' ? '启用' : '禁用' }}
                </n-tag>
              </td>
              <td>
                <n-space>
                  <n-button text type="primary" @click="router.push(`/product/${item.id}/overview`)">进入</n-button>
                  <n-button text type="info" @click="handleEdit(item)">编辑</n-button>
                  <n-button text type="error" @click="handleDelete(item)">删除</n-button>
                </n-space>
              </td>
            </tr>
          </tbody>
        </n-table>
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
      :title="isEdit ? '编辑产品' : '新建产品'"
      positive-text="确定"
      negative-text="取消"
      @positive-click="handleSubmit"
      @update:show="onModalClose"
      style="width: 520px"
    >
      <n-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-placement="left"
        :label-width="80"
        class="py-4"
      >
        <n-form-item label="名称" path="name">
          <n-input v-model:value="formData.name" placeholder="请输入产品名称" />
        </n-form-item>
        <n-form-item label="描述" path="description">
          <n-input
            v-model:value="formData.description"
            type="textarea"
            placeholder="请输入产品描述"
            :rows="3"
          />
        </n-form-item>
      </n-form>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import { ref, reactive, onMounted, nextTick } from 'vue';
  import { useRouter } from 'vue-router';
  import { useMessage, useDialog } from 'naive-ui';
  import { PlusOutlined } from '@vicons/antd';
  import {
    getProducts,
    createProduct,
    updateProduct,
    deleteProduct,
  } from '@/api/product/index';
  import type { ProductItem } from '@/api/product/index';

  const message = useMessage();
  const dialog = useDialog();
  const router = useRouter();
  const loading = ref(false);
  const productList = ref<ProductItem[]>([]);
  const total = ref(0);
  const showModal = ref(false);
  const isEdit = ref(false);
  const editId = ref<number | null>(null);
  const formRef = ref<any>(null);

  const pagination = reactive({ page: 1, size: 10 });

  const formData = reactive({
    name: '',
    description: '',
  });

  const formRules = {
    name: { required: true, message: '请输入产品名称', trigger: 'blur' },
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
      const res = await getProducts({ page: pagination.page, size: pagination.size });
      if (res) {
        productList.value = res.list || [];
        total.value = res.total || 0;
      }
    } catch (e) {
      // ignore
    } finally {
      loading.value = false;
    }
  }

  function handleCreate() {
    resetForm();
    showModal.value = true;
  }

  function handleEdit(item: ProductItem) {
    isEdit.value = true;
    editId.value = item.id;
    formData.name = item.name;
    formData.description = item.description;
    showModal.value = true;
  }

  function handleDelete(item: ProductItem) {
    dialog.warning({
      title: '确认删除',
      content: `确定要删除产品「${item.name}」吗？`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await deleteProduct(item.id);
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
        await updateProduct(editId.value, { ...formData });
        message.success('更新成功');
      } else {
        await createProduct({ ...formData });
        message.success('创建成功');
      }
    } catch (e) {
      message.error('操作失败');
      return false;
    }
  }

  function onModalClose(show: boolean) {
    if (!show) {
      nextTick(() => loadData());
    }
  }

  onMounted(() => {
    loadData();
  });
</script>
