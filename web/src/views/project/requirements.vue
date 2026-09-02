<template>
  <div>

    <n-card :bordered="false" class="proCard">
      <template #header>
        <n-space align="center">
          <n-button type="primary" @click="handleCreate">
            <template #icon>
              <n-icon><PlusOutlined /></n-icon>
            </template>
            新建需求
          </n-button>
        </n-space>
      </template>

      <n-spin :show="reqLoading">
        <n-empty v-if="!reqLoading && requirementList.length === 0" description="暂无需求" />
        <n-data-table
          v-else
          :columns="columns"
          :data="requirementList"
          :row-key="(row: any) => row.id"
          default-expand-all
        />
      </n-spin>
    </n-card>

    <!-- 新建/编辑需求弹窗 -->
    <n-modal
      v-model:show="showModal"
      preset="dialog"
      :title="isEdit ? '编辑需求' : '新建需求'"
      positive-text="确定"
      negative-text="取消"
      @positive-click="handleSubmit"
      style="width: 600px"
    >
      <n-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-placement="left"
        :label-width="100"
        class="py-4"
      >
        <n-form-item label="标题" path="title">
          <n-input v-model:value="formData.title" placeholder="请输入需求标题" />
        </n-form-item>
        <n-form-item label="类型" path="type">
          <n-select v-model:value="formData.type" :options="typeOptions" placeholder="请选择类型" />
        </n-form-item>
        <n-form-item label="优先级" path="priority">
          <n-input-number v-model:value="formData.priority" :min="1" :max="4" placeholder="1=低 2=中 3=高 4=紧急" style="width: 100%" />
        </n-form-item>
        <n-form-item label="描述" path="description">
          <n-input v-model:value="formData.description" type="textarea" placeholder="请输入描述" :rows="3" />
        </n-form-item>
        <n-form-item label="验收标准" path="acceptanceCriteria">
          <n-input v-model:value="formData.acceptanceCriteria" type="textarea" placeholder="请输入验收标准" :rows="3" />
        </n-form-item>
        <n-form-item label="父级需求" path="parentId" v-if="!isEdit">
          <n-select
            v-model:value="formData.parentId"
            :options="parentOptions"
            placeholder="无父级（顶级需求）"
            clearable
          />
        </n-form-item>
      </n-form>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import { ref, reactive, computed, onMounted, h } from 'vue';
  import { useRoute } from 'vue-router';
  import { useMessage, useDialog, NTag, NSpace, NButton } from 'naive-ui';
  import { PlusOutlined } from '@vicons/antd';
  import {
    getRequirements,
    createRequirement,
    updateRequirement,
    deleteRequirement,
  } from '@/api/project/index';
  import type { RequirementItem } from '@/api/project/index';

  const message = useMessage();
  const dialog = useDialog();
  const route = useRoute();
  const projectId = computed(() => Number(route.params.projectId));
  const reqLoading = ref(false);
  const showModal = ref(false);
  const isEdit = ref(false);
  const editId = ref<number | null>(null);
  const formRef = ref<any>(null);

  const requirementList = ref<RequirementItem[]>([]);

  const parentOptions = computed(() =>
    requirementList.value
      .filter((r) => r.type === 'epic' || r.type === 'story')
      .map((r) => ({ label: `${r.title} (${r.type})`, value: r.id }))
  );

  const typeOptions = [
    { label: 'Epic', value: 'epic' },
    { label: 'Story', value: 'story' },
    { label: 'Task', value: 'task' },
  ];

  const priorityLabels: Record<number, string> = { 1: '低', 2: '中', 3: '高', 4: '紧急' };
  const priorityColor: Record<number, string> = { 1: 'default', 2: 'info', 3: 'warning', 4: 'error' };
  const typeColor: Record<string, string> = { epic: 'success', story: 'info', task: 'default' };

  const statusTransitions: Record<string, { label: string; target: string }[]> = {
    draft: [{ label: '激活', target: 'active' }],
    active: [
      { label: '完成', target: 'completed' },
      { label: '归档', target: 'archived' },
    ],
    completed: [{ label: '归档', target: 'archived' }],
    archived: [{ label: '激活', target: 'active' }],
  };

  const formData = reactive({
    title: '',
    description: '',
    type: 'task',
    priority: 2,
    acceptanceCriteria: '',
    parentId: null as number | null,
  });

  const formRules = {
    title: { required: true, message: '请输入需求标题', trigger: 'blur' },
    type: { required: true, message: '请选择类型', trigger: 'change' },
  };

  const columns = [
    { title: '标题', key: 'title', width: 250 },
    {
      title: '类型',
      key: 'type',
      width: 80,
      render(row: RequirementItem) {
        return h(NTag, { type: typeColor[row.type] as any, size: 'small' }, () => row.type.toUpperCase());
      },
    },
    {
      title: '优先级',
      key: 'priority',
      width: 80,
      render(row: RequirementItem) {
        return h(NTag, { type: priorityColor[row.priority] as any, size: 'small' }, () => priorityLabels[row.priority] || row.priority);
      },
    },
    {
      title: '状态',
      key: 'status',
      width: 100,
      render(row: RequirementItem) {
        const statusLabels: Record<string, string> = { draft: '草稿', active: '进行中', completed: '已完成', archived: '已归档' };
        return h(NTag, { size: 'small' }, () => statusLabels[row.status] || row.status);
      },
    },
    { title: '指派人', key: 'assigneeName', width: 100 },
    {
      title: '操作',
      key: 'action',
      width: 240,
      render(row: RequirementItem) {
        const actions = [
          h(NButton, { text: true, type: 'info', onClick: () => handleEdit(row) }, () => '编辑'),
          h(NButton, { text: true, type: 'error', onClick: () => handleDelete(row) }, () => '删除'),
        ];
        const transitions = statusTransitions[row.status] || [];
        transitions.forEach((t) => {
          actions.push(
            h(NButton, { text: true, type: 'warning', onClick: () => handleTransition(row, t.target) }, () => t.label)
          );
        });
        return h(NSpace, null, () => actions);
      },
    },
  ];

  function resetForm() {
    formData.title = '';
    formData.description = '';
    formData.type = 'task';
    formData.priority = 2;
    formData.acceptanceCriteria = '';
    formData.parentId = null;
    isEdit.value = false;
    editId.value = null;
  }

  async function loadRequirements() {
    reqLoading.value = true;
    try {
      const res = await getRequirements(projectId.value);
      if (res) {
        requirementList.value = res.list || [];
      }
    } catch (e) {
      // ignore
    } finally {
      reqLoading.value = false;
    }
  }

  function handleCreate() {
    resetForm();
    showModal.value = true;
  }

  function handleEdit(item: RequirementItem) {
    isEdit.value = true;
    editId.value = item.id;
    formData.title = item.title;
    formData.description = item.description;
    formData.type = item.type;
    formData.priority = item.priority;
    formData.acceptanceCriteria = item.acceptanceCriteria;
    showModal.value = true;
  }

  function handleDelete(item: RequirementItem) {
    dialog.warning({
      title: '确认删除',
      content: `确定要删除需求「${item.title}」吗？`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await deleteRequirement(item.id);
          message.success('删除成功');
          loadRequirements();
        } catch (e) {
          message.error('删除失败');
        }
      },
    });
  }

  async function handleTransition(item: RequirementItem, targetStatus: string) {
    try {
      await updateRequirement(item.id, { status: targetStatus });
      message.success('状态更新成功');
      loadRequirements();
    } catch (e) {
      message.error('状态更新失败');
    }
  }

  async function handleSubmit() {
    try {
      await formRef.value?.validate();
    } catch {
      return false;
    }
    try {
      if (isEdit.value && editId.value) {
        await updateRequirement(editId.value, {
          title: formData.title,
          description: formData.description,
          type: formData.type,
          priority: formData.priority,
          acceptanceCriteria: formData.acceptanceCriteria,
        });
        message.success('更新成功');
      } else {
        await createRequirement(projectId.value, {
          ...formData,
          parentId: formData.parentId ?? undefined,
        });
        message.success('创建成功');
      }
      showModal.value = false;
      loadRequirements();
    } catch (e) {
      message.error('操作失败');
      return false;
    }
  }

  onMounted(loadRequirements);
</script>
