<template>
  <div class="board">
    <div v-for="col in boardColumns" :key="col.status" class="board-col" :data-status="col.status">
      <div class="board-col-head">
        <span class="board-col-title">{{ col.label }}</span>
        <n-tag size="small" round :bordered="false">{{ columnTasks[col.status].length }}</n-tag>
      </div>
      <n-spin :show="loading" size="small">
        <Draggable
          :list="columnTasks[col.status]"
          item-key="id"
          group="board"
          animation="200"
          class="board-col-list"
          @change="onDrop(col.status, $event)"
        >
          <template #item="{ element }">
            <div class="task-card" @click="openTaskDetail(element)">
              <div class="task-title">{{ element.title }}</div>
              <div class="task-meta">
                <n-tag :type="typeColor[element.type] as any" size="tiny" :bordered="false">
                  {{ element.type }}
                </n-tag>
                <n-tag :type="priorityColor(element.priority)" size="tiny" :bordered="false">
                  P{{ element.priority }}
                </n-tag>
                <span v-if="element.assigneeName" class="task-assignee">{{ element.assigneeName }}</span>
              </div>
            </div>
          </template>
        </Draggable>
        <div v-if="!loading && columnTasks[col.status].length === 0" class="board-col-empty">
          暂无任务（可拖入）
        </div>
      </n-spin>
    </div>

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
  // 看板铺满内容区视口高：列横向等宽排布，列内滚动，整页不滚
  .board {
    display: flex;
    gap: 10px;
    flex: 1;
    min-height: 0;
  }

  .board-col {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    border-radius: var(--panel-radius, 12px);
    background: var(--panel-bg, #fff);
    box-shadow: var(--panel-shadow);
    overflow: hidden;

    // 列身份色条
    &[data-status='open'] {
      border-top: 3px solid #d9d9d9;
    }
    &[data-status='in_progress'] {
      border-top: 3px solid #2080f0;
    }
    &[data-status='review'] {
      border-top: 3px solid #f0a020;
    }
    &[data-status='done'] {
      border-top: 3px solid #18a058;
    }
  }

  .board-col-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 12px 6px;
    flex-shrink: 0;

    .board-col-title {
      font-size: 13px;
      font-weight: 600;
      color: #333;
    }
  }

  // 列内容：滚动区在列卡内部
  .board-col :deep(.n-spin-container) {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;

    .n-spin-content {
      flex: 1;
      min-height: 0;
      overflow-y: auto;
      padding: 0 8px 8px;
    }
  }

  .board-col-list {
    // 空列也保留拖放目标区
    min-height: 120px;
  }

  .task-card {
    background: var(--canvas, #f7f7f4);
    border: 1px solid rgb(0 21 41 / 6%);
    border-radius: 8px;
    padding: 10px;
    margin-bottom: 8px;
    cursor: pointer;
    transition: box-shadow 0.15s, border-color 0.15s;

    &:hover {
      border-color: #16a34a;
      box-shadow: 0 2px 8px rgb(0 21 41 / 8%);
    }

    .task-title {
      font-size: 13px;
      font-weight: 500;
      color: #333;
      word-break: break-all;
    }

    .task-meta {
      display: flex;
      align-items: center;
      gap: 6px;
      margin-top: 6px;

      .task-assignee {
        margin-left: auto;
        font-size: 12px;
        color: #999;
        max-width: 90px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }
  }

  .board-col-empty {
    text-align: center;
    color: #bbb;
    font-size: 12px;
    padding: 18px 0 10px;
  }
</style>
