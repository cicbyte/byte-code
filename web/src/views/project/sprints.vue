<template>
  <div class="p-4">
    <n-card title="Sprint 管理">
      <template #header-extra>
        <n-button type="primary" @click="showCreate = true">新建 Sprint</n-button>
      </template>

      <n-data-table :columns="columns" :data="sprints" :loading="loading" />
    </n-card>

    <n-modal v-model:show="showCreate" title="新建 Sprint" preset="card" style="width: 500px">
      <n-form ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="80">
        <n-form-item label="名称" path="name">
          <n-input v-model:value="form.name" placeholder="Sprint 1" />
        </n-form-item>
        <n-form-item label="目标" path="goal">
          <n-input v-model:value="form.goal" type="textarea" placeholder="Sprint 目标" />
        </n-form-item>
        <n-form-item label="开始日期" path="startDate">
          <n-date-picker v-model:value="form.startDate" type="date" style="width: 100%" />
        </n-form-item>
        <n-form-item label="结束日期" path="endDate">
          <n-date-picker v-model:value="form.endDate" type="date" style="width: 100%" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showCreate = false">取消</n-button>
          <n-button type="primary" @click="handleCreate" :loading="submitting">创建</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, h } from 'vue';
import { NCard, NButton, NDataTable, NModal, NForm, NFormItem, NInput, NDatePicker, NSpace, NTag, useMessage } from 'naive-ui';
import { getSprints, createSprint } from '@/api/project';

const message = useMessage();
const loading = ref(false);
const submitting = ref(false);
const sprints = ref<any[]>([]);
const showCreate = ref(false);
const projectId = ref(0);
const formRef = ref();

const form = ref({
  name: '',
  goal: '',
  startDate: null as number | null,
  endDate: null as number | null,
});

const rules = {
  name: { required: true, message: '请输入名称' },
  startDate: { required: true, message: '请选择开始日期' },
  endDate: { required: true, message: '请选择结束日期' },
};

const statusMap: Record<string, { type: 'default' | 'info' | 'success' | 'warning'; label: string }> = {
  planning: { type: 'default', label: '规划中' },
  active: { type: 'info', label: '进行中' },
  completed: { type: 'success', label: '已完成' },
};

const columns = [
  { title: '名称', key: 'name' },
  { title: '目标', key: 'goal' },
  { title: '开始日期', key: 'startDate' },
  { title: '结束日期', key: 'endDate' },
  {
    title: '状态', key: 'status',
    render: (row: any) => {
      const s = statusMap[row.status] || { type: 'default' as const, label: row.status };
      return h(NTag, { type: s.type, size: 'small' }, () => s.label);
    },
  },
];

async function loadSprints() {
  loading.value = true;
  try {
    // Sprint 需要在项目上下文中，这里显示所有项目的 Sprint
    // TODO: 实现全局 Sprint 列表
    sprints.value = [];
  } finally {
    loading.value = false;
  }
}

async function handleCreate() {
  submitting.value = true;
  try {
    const fmt = (ts: number) => new Date(ts).toISOString().split('T')[0];
    await createSprint(projectId.value, {
      name: form.value.name,
      goal: form.value.goal,
      startDate: form.value.startDate ? fmt(form.value.startDate) : '',
      endDate: form.value.endDate ? fmt(form.value.endDate) : '',
    });
    message.success('创建成功');
    showCreate.value = false;
    form.value = { name: '', goal: '', startDate: null, endDate: null };
    loadSprints();
  } catch (e: any) {
    message.error(e.message || '创建失败');
  } finally {
    submitting.value = false;
  }
}

onMounted(loadSprints);
</script>
