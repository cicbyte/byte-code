<template>
  <div>
    <n-card :bordered="false" class="proCard">
      <template #header-extra>
        <n-button type="primary" @click="showCreate = true">新建 Sprint</n-button>
      </template>

      <n-spin :show="loading">
        <n-empty v-if="!loading && sprints.length === 0" description="暂无 Sprint" />
        <n-table v-else :bordered="false" :single-line="false" size="small">
          <thead>
            <tr>
              <th>名称</th>
              <th>目标</th>
              <th>开始日期</th>
              <th>结束日期</th>
              <th>状态</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="sprint in sprints" :key="sprint.id">
              <td>{{ sprint.name }}</td>
              <td>{{ sprint.goal }}</td>
              <td>{{ sprint.startDate }}</td>
              <td>{{ sprint.endDate }}</td>
              <td>
                <n-tag :type="statusMap[sprint.status]?.type || 'default'" size="small">
                  {{ statusMap[sprint.status]?.label || sprint.status }}
                </n-tag>
              </td>
            </tr>
          </tbody>
        </n-table>
      </n-spin>
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
          <n-date-picker v-model:formatted-value="form.startDate" type="date" placeholder="选择开始日期" style="width: 100%" />
        </n-form-item>
        <n-form-item label="结束日期" path="endDate">
          <n-date-picker v-model:formatted-value="form.endDate" type="date" placeholder="选择结束日期" style="width: 100%" />
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
  import { ref, computed, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { useMessage } from 'naive-ui';
  import { getSprints, createSprint } from '@/api/project';
  import type { SprintItem } from '@/api/project';

  const route = useRoute();
  const message = useMessage();
  const projectId = computed(() => Number(route.params.projectId));

  const loading = ref(false);
  const submitting = ref(false);
  const sprints = ref<SprintItem[]>([]);
  const showCreate = ref(false);
  const formRef = ref();

  const form = ref({ name: '', goal: '', startDate: '', endDate: '' });
  const rules = {
    name: { required: true, message: '请输入名称' },
  };

  const statusMap: Record<string, { type: 'default' | 'info' | 'success' | 'warning'; label: string }> = {
    planning: { type: 'default', label: '规划中' },
    active: { type: 'info', label: '进行中' },
    completed: { type: 'success', label: '已完成' },
  };

  async function loadSprints() {
    loading.value = true;
    try {
      const res = await getSprints(projectId.value);
      sprints.value = res?.list || [];
    } catch {
      // ignore
    } finally {
      loading.value = false;
    }
  }

  async function handleCreate() {
    submitting.value = true;
    try {
      await createSprint(projectId.value, { ...form.value });
      message.success('创建成功');
      showCreate.value = false;
      form.value = { name: '', goal: '', startDate: '', endDate: '' };
      loadSprints();
    } catch (e: any) {
      message.error(e.message || '创建失败');
    } finally {
      submitting.value = false;
    }
  }

  onMounted(loadSprints);
</script>
