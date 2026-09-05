<template>
  <div>
    <n-card :bordered="false" class="proCard">
      <n-space class="mb-4" align="center">
        <n-input
          v-model:value="filter.keyword"
          placeholder="任务标题 / 描述"
          clearable
          style="width: 220px"
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        />
        <n-select
          v-model:value="filter.status"
          :options="statusOptions"
          placeholder="状态"
          clearable
          style="width: 140px"
          @update:value="handleSearch"
        />
        <n-select
          v-model:value="filter.type"
          :options="typeOptions"
          placeholder="类型"
          clearable
          style="width: 140px"
          @update:value="handleSearch"
        />
        <n-select
          v-model:value="filter.tagId"
          :options="tagOptions"
          clearable
          placeholder="标签"
          style="width: 130px"
          @update:value="handleSearch"
        />
        <n-button type="primary" @click="handleSearch">查询</n-button>
        <n-button @click="handleReset">重置</n-button>
        <n-button type="primary" @click="openCreate">新建任务</n-button>
      </n-space>

      <n-table :bordered="false" :single-line="false" size="small">
        <thead>
          <tr>
            <th>标题</th>
            <th>类型</th>
            <th>优先级</th>
            <th>标签</th>
            <th>状态</th>
            <th>指派人</th>
            <th>截止</th>
            <th>更新时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="task in taskList" :key="task.id">
            <td>
              <n-button text type="info" @click="openTaskDetail(task)">{{ task.title }}</n-button>
            </td>
            <td><n-tag :type="typeTagType(task.type)" size="small">{{ typeLabel(task.type) }}</n-tag></td>
            <td><n-tag :type="priorityTagType(task.priority)" size="small">P{{ task.priority }}</n-tag></td>
            <td>
              <n-space :size="2">
                <n-tag v-for="t in task.tags || []" :key="t" size="small" round :bordered="false">{{ t }}</n-tag>
              </n-space>
            </td>
            <td><n-tag size="small" :type="statusTagType(task.status)">{{ statusLabel(task.status) }}</n-tag></td>
            <td>{{ task.assigneeName }}</td>
            <td>
              <n-tag
                v-if="task.dueDate"
                size="small"
                :type="dueTagType(task.dueDate, task.status) || 'default'"
              >{{ dueLabel(task.dueDate) }}</n-tag>
              <span v-else>-</span>
            </td>
            <td>{{ task.updatedAt }}</td>
            <td>
              <n-space size="small">
                <n-button text type="info" @click="openTaskDetail(task)">详情</n-button>
                <n-button text type="primary" @click="openEdit(task)">编辑</n-button>
                <n-button v-if="task.status === 'done'" text type="warning" @click="handleCloseTask(task)">关闭</n-button>
                <n-button text type="error" @click="handleDelete(task)">删除</n-button>
              </n-space>
            </td>
          </tr>
        </tbody>
      </n-table>

      <div class="mt-4 flex justify-end" v-if="total > pagination.size">
        <n-pagination
          v-model:page="pagination.page"
          :page-size="pagination.size"
          :item-count="total"
          @update:page="loadTasks"
        />
      </div>
    </n-card>

    <!-- 新建/编辑任务弹窗（复用：editingId 区分） -->
    <n-modal
      v-model:show="showCreateModal"
      preset="dialog"
      :title="editingId ? '编辑任务' : '新建任务'"
      positive-text="确定"
      negative-text="取消"
      @positive-click="handleSubmit"
      style="width: 560px"
    >
      <n-form ref="formRef" :model="formData" :rules="formRules" label-placement="left" :label-width="80" class="py-4">
        <n-form-item label="标题" path="title">
          <n-input v-model:value="formData.title" placeholder="请输入任务标题" />
        </n-form-item>
        <n-form-item label="类型" path="type">
          <n-select v-model:value="formData.type" :options="typeOptions" placeholder="请选择类型" />
        </n-form-item>
        <n-form-item label="优先级" path="priority">
          <n-input-number v-model:value="formData.priority" :min="1" :max="4" placeholder="1=低 4=紧急" style="width: 100%" />
        </n-form-item>
        <n-form-item label="指派人" path="assigneeId">
          <n-select
            v-model:value="formData.assigneeId"
            :options="assigneeOptions"
            clearable
            filterable
            placeholder="缺省不指派（可后续 AI 认领）"
          />
        </n-form-item>
        <n-form-item label="截止日期" path="dueDate">
          <n-date-picker
            v-model:formatted-value="formData.dueDate"
            type="date"
            value-format="yyyy-MM-dd"
            clearable
            placeholder="缺省无截止"
            style="width: 100%"
          />
        </n-form-item>
        <n-form-item label="描述" path="description">
          <n-input v-model:value="formData.description" type="textarea" placeholder="请输入描述" :rows="3" />
        </n-form-item>
      </n-form>
    </n-modal>

    <TaskDetailModal ref="taskDetailRef" @updated="loadTasks" />
  </div>
</template>

<script lang="ts" setup>
  import { ref, reactive, computed, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { useMessage, useDialog } from 'naive-ui';
  import { getTags } from '@/api/platform/index';
  import { getTasks, createTask, updateTask, deleteTask, getMembers } from '@/api/project/index';
  import type { TaskItem } from '@/api/project/index';
  import TaskDetailModal from '@/views/project/components/TaskDetailModal.vue';
  import { dueTagType, dueLabel } from '@/utils/taskDue';
  import {
    TASK_STATUS, TASK_TYPES,
    statusLabel, statusTagType, typeLabel, typeTagType, priorityTagType,
  } from '@/enums/task';

  const route = useRoute();
  const message = useMessage();
  const dialog = useDialog();
  const projectId = computed(() => Number(route.params.projectId));

  const taskList = ref<TaskItem[]>([]);
  const total = ref(0);
  const pagination = reactive({ page: 1, size: 20 });
  const taskDetailRef = ref();

  // 状态/类型/优先级字典统一出口：enums/task.ts
  const statusOptions = TASK_STATUS.map((s) => ({ label: s.label, value: s.value }));
  const typeOptions = TASK_TYPES.map((t) => ({ label: t.label, value: t.value }));

  function openTaskDetail(task: TaskItem) {
    taskDetailRef.value?.openModal(task.id);
  }

  // 新建/编辑任务（弹窗复用：editingId 为空=新建）
  const showCreateModal = ref(false);
  const editingId = ref<number | null>(null);
  const formRef = ref<any>(null);
  const filter = reactive({ keyword: '', status: null as string | null, type: null as string | null, tagId: null as number | null });
  // 标签筛选选项（平台级标签）
  const tagOptions = ref<Array<{ label: string; value: number }>>([]);
  async function loadTagOptions() {
    try {
      const res = await getTags();
      tagOptions.value = (res?.list || []).map((t: any) => ({ label: t.name, value: t.id }));
    } catch {
      // ignore
    }
  }
  loadTagOptions();

  // 指派人选项：项目成员（转派通知由 UpdateTask 后端触发）
  const assigneeOptions = ref<Array<{ label: string; value: number }>>([]);
  async function loadAssigneeOptions() {
    try {
      const res = await getMembers(projectId.value);
      assigneeOptions.value = (res?.list || []).map((m) => ({
        label: m.realName || m.username,
        value: m.userId,
      }));
    } catch {
      // ignore
    }
  }
  loadAssigneeOptions();

  const formData = reactive({
    title: '', type: 'feature', priority: 2, description: '',
    assigneeId: null as number | null, dueDate: null as string | null,
  });
  const formRules = { title: { required: true, message: '请输入任务标题', trigger: 'blur' } };

  function resetForm() {
    editingId.value = null;
    formData.title = ''; formData.type = 'feature'; formData.priority = 2;
    formData.description = ''; formData.assigneeId = null; formData.dueDate = null;
  }

  function openCreate() {
    resetForm();
    showCreateModal.value = true;
  }

  function openEdit(task: TaskItem) {
    editingId.value = task.id;
    formData.title = task.title;
    formData.type = typeOptions.some((o) => o.value === task.type) ? task.type : 'feature';
    formData.priority = task.priority;
    formData.description = task.description || '';
    formData.assigneeId = task.assigneeId || null;
    formData.dueDate = task.dueDate || null;
    showCreateModal.value = true;
  }

  async function loadTasks() {
    try {
      const res = await getTasks(projectId.value, {
        page: pagination.page, size: pagination.size,
        status: filter.status ?? undefined,
        type: filter.type ?? undefined,
        tagId: filter.tagId ?? undefined,
        keyword: filter.keyword || undefined,
      });
      if (res) { taskList.value = res.list || []; total.value = res.total || 0; }
    } catch { /* ignore */ }
  }

  function handleSearch() {
    pagination.page = 1;
    loadTasks();
  }

  function handleReset() {
    filter.keyword = '';
    filter.status = null;
    filter.type = null;
    filter.tagId = null;
    pagination.page = 1;
    loadTasks();
  }

  async function handleSubmit() {
    try { await formRef.value?.validate(); } catch { return false; }
    try {
      if (editingId.value) {
        // 编辑：全字段更新（指派变更触发转派通知）
        await updateTask(editingId.value, {
          title: formData.title,
          type: formData.type,
          priority: formData.priority,
          description: formData.description,
          assigneeId: formData.assigneeId ?? 0,
          dueDate: formData.dueDate || '',
        });
        message.success('任务已更新');
      } else {
        await createTask(projectId.value, {
          ...formData,
          assigneeId: formData.assigneeId ?? undefined,
          dueDate: formData.dueDate || undefined,
        });
        message.success('任务创建成功');
      }
      showCreateModal.value = false;
      // 纯重置（不碰 showCreateModal）：openCreate 的二次翻转依赖 naive 补发 update:show，脆弱
      resetForm();
      loadTasks();
    } catch { message.error(editingId.value ? '更新失败' : '创建失败'); return false; }
  }

  // done→closed 清账（终态入口；列表场景快捷操作）
  function handleCloseTask(task: TaskItem) {
    dialog.warning({
      title: '确认关闭任务',
      content: `「${task.title}」将转入终态（closed），不再出现在活跃列表。`,
      positiveText: '关闭',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await updateTask(task.id, { status: 'closed' });
          message.success('已关闭');
          loadTasks();
        } catch {
          message.error('关闭失败');
        }
      },
    });
  }

  function handleDelete(task: TaskItem) {
    dialog.warning({
      title: '确认删除',
      content: `确定要删除任务「${task.title}」吗？`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        try { await deleteTask(task.id); message.success('删除成功'); loadTasks(); }
        catch { message.error('删除失败'); }
      },
    });
  }

  onMounted(loadTasks);
</script>
