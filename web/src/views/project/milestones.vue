<template>
  <div>
    <n-card :bordered="false" title="里程碑" class="proCard">
      <template #header-extra>
        <n-button type="primary" @click="openCreate">新建里程碑</n-button>
      </template>

      <n-spin :show="loading">
        <EmptyState type="generic" title="暂无里程碑" v-if="!loading && milestones.length === 0" description="创建里程碑划定阶段目标" />
        <n-table v-else :bordered="false" :single-line="false" size="small">
          <thead>
            <tr>
              <th>名称</th>
              <th>描述</th>
              <th>目标日期</th>
              <th>状态</th>
              <th style="width: 100px">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in milestones" :key="item.id">
              <td>{{ item.name }}</td>
              <td>{{ item.description || '-' }}</td>
              <td>{{ item.targetDate || '-' }}</td>
              <td>
                <n-tag size="small" :type="statusType(item.status)">{{ statusLabel(item.status) }}</n-tag>
              </td>
              <td>
                <n-space size="small">
                  <n-button text type="primary" size="small" @click="openEdit(item)">编辑</n-button>
                  <n-button text type="error" size="small" @click="handleDelete(item)">删除</n-button>
                </n-space>
              </td>
            </tr>
          </tbody>
        </n-table>
      </n-spin>
    </n-card>

    <!-- 新建/编辑弹窗 -->
    <n-modal v-model:show="showModal" :title="isEdit ? '编辑里程碑' : '新建里程碑'" preset="card" style="width: 500px">
      <n-form ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="80">
        <n-form-item label="名称" path="name">
          <n-input v-model:value="form.name" placeholder="请输入里程碑名称" />
        </n-form-item>
        <n-form-item label="描述" path="description">
          <n-input v-model:value="form.description" type="textarea" placeholder="请输入描述" />
        </n-form-item>
        <n-form-item label="目标日期" path="targetDate">
          <n-date-picker v-model:formatted-value="form.targetDate" type="date" placeholder="选择目标日期" style="width: 100%" clearable />
        </n-form-item>
        <n-form-item label="状态" path="status">
          <n-select v-model:value="form.status" :options="statusOptions" placeholder="选择状态" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showModal = false">取消</n-button>
          <n-button type="primary" @click="handleSubmit" :loading="submitting">
            {{ isEdit ? '保存' : '创建' }}
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { ref, computed, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { MILESTONE_STATUS } from '@/enums/entities';
  import { useMessage, useDialog } from 'naive-ui';
  import { getMilestones, createMilestone, updateMilestone, deleteMilestone } from '@/api/project/index';
  import type { MilestoneItem } from '@/api/project/index';

  const route = useRoute();
  const message = useMessage();
  const dialog = useDialog();
  const projectId = computed(() => Number(route.params.projectId));

  const loading = ref(false);
  const submitting = ref(false);
  const milestones = ref<MilestoneItem[]>([]);
  const showModal = ref(false);
  const isEdit = ref(false);
  const editId = ref<number | null>(null);

  const form = ref({ name: '', description: '', targetDate: '', status: 'planning' });
  const rules = { name: { required: true, message: '请输入名称' } };

  // 字典统一出口：enums/entities.ts（原 statusOptions 与 statusLabel 内映射重复两遍）
  const statusOptions = MILESTONE_STATUS.options;
  const statusLabel = MILESTONE_STATUS.label;
  const statusType = MILESTONE_STATUS.tagType;

  function openCreate() {
    isEdit.value = false;
    editId.value = null;
    form.value = { name: '', description: '', targetDate: '', status: 'planning' };
    showModal.value = true;
  }

  function openEdit(item: MilestoneItem) {
    isEdit.value = true;
    editId.value = item.id;
    form.value = {
      name: item.name,
      description: item.description || '',
      targetDate: item.targetDate || '',
      status: item.status || 'planning',
    };
    showModal.value = true;
  }

  async function handleSubmit() {
    submitting.value = true;
    try {
      if (isEdit.value && editId.value) {
        await updateMilestone(editId.value, { ...form.value });
        message.success('更新成功');
      } else {
        await createMilestone(projectId.value, { ...form.value });
        message.success('创建成功');
      }
      showModal.value = false;
      loadMilestones();
    } catch (e: any) {
      message.error(e.message || (isEdit.value ? '更新失败' : '创建失败'));
    } finally {
      submitting.value = false;
    }
  }

  function handleDelete(item: MilestoneItem) {
    dialog.warning({
      title: '删除里程碑',
      content: `确定要删除「${item.name}」吗？关联的需求将解除关联但不会被删除。`,
      positiveText: '删除',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await deleteMilestone(item.id);
          message.success('删除成功');
          loadMilestones();
        } catch (e: any) {
          message.error(e.message || '删除失败');
        }
      },
    });
  }

  async function loadMilestones() {
    loading.value = true;
    try {
      const res = await getMilestones(projectId.value);
      milestones.value = res?.list || [];
    } catch {
      // ignore
    } finally {
      loading.value = false;
    }
  }

  onMounted(loadMilestones);
</script>
