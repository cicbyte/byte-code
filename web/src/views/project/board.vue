<template>
  <div>
    <n-card :bordered="false" class="proCard">
      <n-spin :show="loading">
        <n-grid cols="1 s:2 m:4" :x-gap="12" :y-gap="12">
          <n-grid-item v-for="col in boardColumns" :key="col.status">
            <n-card :title="col.label" size="small" :bordered="true" :segmented="{ content: true }">
              <template #header-extra>
                <n-tag size="small" round>{{ getTasksByStatus(col.status).length }}</n-tag>
              </template>
              <n-space vertical :size="8">
                <n-empty v-if="getTasksByStatus(col.status).length === 0" description="暂无任务" size="small" />
                <n-card
                  v-for="task in getTasksByStatus(col.status)"
                  :key="task.id"
                  size="small"
                  hoverable
                  style="cursor: pointer"
                  @click="openTaskDetail(task)"
                >
                  <div class="text-sm font-medium">{{ task.title }}</div>
                  <n-space class="mt-2" size="small">
                    <n-tag :type="typeColor[task.type] as any" size="tiny">{{ task.type }}</n-tag>
                    <n-tag :type="priorityColor(task.priority) as any" size="tiny">P{{ task.priority }}</n-tag>
                  </n-space>
                  <div class="mt-2 text-xs text-gray-400" v-if="task.assigneeName">{{ task.assigneeName }}</div>
                </n-card>
              </n-space>
            </n-card>
          </n-grid-item>
        </n-grid>
      </n-spin>
    </n-card>

    <TaskDetailModal ref="taskDetailRef" />
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { getTasks } from '@/api/project/index';
  import type { TaskItem } from '@/api/project/index';
  import TaskDetailModal from '@/views/project/components/TaskDetailModal.vue';

  const route = useRoute();
  const projectId = computed(() => Number(route.params.projectId));

  const loading = ref(false);
  const taskList = ref<TaskItem[]>([]);
  const taskDetailRef = ref();

  const boardColumns = [
    { status: 'open', label: 'Open' },
    { status: 'in_progress', label: 'In Progress' },
    { status: 'review', label: 'Review' },
    { status: 'done', label: 'Done' },
  ];

  const typeColor: Record<string, string> = {
    bug: 'error',
    feature: 'success',
    improvement: 'info',
    task: 'default',
  };

  function priorityColor(p: number): 'default' | 'info' | 'warning' | 'error' {
    if (p >= 4) return 'error';
    if (p >= 3) return 'warning';
    if (p >= 2) return 'info';
    return 'default';
  }

  function getTasksByStatus(status: string) {
    return taskList.value.filter((t) => t.status === status);
  }

  function openTaskDetail(task: TaskItem) {
    taskDetailRef.value?.openModal(task.id);
  }

  async function loadTasks() {
    loading.value = true;
    try {
      const res = await getTasks(projectId.value, { page: 1, size: 100 });
      if (res) taskList.value = res.list || [];
    } catch {
      // ignore
    } finally {
      loading.value = false;
    }
  }

  onMounted(loadTasks);
</script>
