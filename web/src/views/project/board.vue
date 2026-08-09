<template>
  <div>
    <n-card :bordered="false" class="proCard">
      <n-spin :show="loading">
        <n-grid cols="1 s:2 m:4" :x-gap="12" :y-gap="12">
          <n-grid-item v-for="col in boardColumns" :key="col.status">
            <n-card :title="col.label" size="small" :bordered="true" :segmented="{ content: true }">
              <template #header-extra>
                <n-tag size="small" round>{{ columnTasks[col.status].length }}</n-tag>
              </template>
              <Draggable
                :list="columnTasks[col.status]"
                item-key="id"
                group="board"
                animation="200"
                class="board-col-list"
                @change="onDrop(col.status, $event)"
              >
                <template #item="{ element }">
                  <n-card
                    class="task-card"
                    size="small"
                    hoverable
                    style="cursor: pointer"
                    @click="openTaskDetail(element)"
                  >
                    <div class="text-sm font-medium">{{ element.title }}</div>
                    <n-space class="mt-2" size="small">
                      <n-tag :type="typeColor[element.type] as any" size="tiny">{{ element.type }}</n-tag>
                      <n-tag :type="priorityColor(element.priority) as any" size="tiny">
                        P{{ element.priority }}
                      </n-tag>
                    </n-space>
                    <div class="mt-2 text-xs text-gray-400" v-if="element.assigneeName">
                      {{ element.assigneeName }}
                    </div>
                  </n-card>
                </template>
              </Draggable>
              <n-empty
                v-if="columnTasks[col.status].length === 0"
                description="暂无任务（可拖入）"
                size="small"
              />
            </n-card>
          </n-grid-item>
        </n-grid>
      </n-spin>
    </n-card>

    <TaskDetailModal ref="taskDetailRef" />
  </div>
</template>

<script lang="ts" setup>
  import { ref, reactive, computed, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { useMessage } from 'naive-ui';
  import Draggable from 'vuedraggable';
  import { getTasks, updateTask } from '@/api/project/index';
  import type { TaskItem } from '@/api/project/index';
  import TaskDetailModal from '@/views/project/components/TaskDetailModal.vue';

  const route = useRoute();
  const message = useMessage();
  const projectId = computed(() => Number(route.params.projectId));

  const loading = ref(false);
  const taskDetailRef = ref();

  const boardColumns = [
    { status: 'open', label: 'Open' },
    { status: 'in_progress', label: 'In Progress' },
    { status: 'review', label: 'Review' },
    { status: 'done', label: 'Done' },
  ];

  // 每列独立的任务数组供 Draggable 原地变更；closed 状态不在看板四列内，不展示
  const columnTasks = reactive<Record<string, TaskItem[]>>({
    open: [],
    in_progress: [],
    review: [],
    done: [],
  });

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

  function syncColumns(list: TaskItem[]) {
    const map: Record<string, TaskItem[]> = { open: [], in_progress: [], review: [], done: [] };
    for (const t of list) {
      if (map[t.status]) map[t.status].push(t);
    }
    Object.assign(columnTasks, map);
  }

  // vuedraggable 的 change 事件：目标列收到 {added: {element, newIndex}}
  function onDrop(status: string, evt: { added?: { element: TaskItem; newIndex: number } }) {
    const task = evt.added?.element;
    if (!task || task.status === status) return;
    // 乐观更新本地状态，失败时整板重拉还原
    task.status = status;
    updateTask(task.id, { status }).catch((e: any) => {
      message.error(e.message || '更新任务状态失败');
      loadTasks();
    });
  }

  function openTaskDetail(task: TaskItem) {
    taskDetailRef.value?.openModal(task.id);
  }

  async function loadTasks() {
    loading.value = true;
    try {
      // 看板需要整板展示，一次取足量（超大规模项目需改虚拟滚动）
      const res = await getTasks(projectId.value, { page: 1, size: 500 });
      syncColumns(res?.list || []);
    } catch {
      // ignore
    } finally {
      loading.value = false;
    }
  }

  onMounted(loadTasks);
</script>

<style lang="less" scoped>
  .board-col-list {
    // 空列也保留足够高的拖放目标区
    min-height: 160px;

    .task-card {
      margin-bottom: 8px;
    }
  }
</style>
