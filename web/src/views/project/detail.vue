<template>
  <div>
    <div class="n-layout-page-header">
      <n-card :bordered="false">
        <n-space justify="space-between" align="center">
          <n-space align="center">
            <n-button text @click="router.push('/project/list')">
              <template #icon>
                <n-icon><ArrowLeftOutlined /></n-icon>
              </template>
              返回列表
            </n-button>
            <span class="text-lg font-medium">{{ projectInfo.name || '项目详情' }}</span>
            <n-tag v-if="projectInfo.status !== undefined" :type="projectInfo.status === 1 ? 'success' : 'default'" size="small">
              {{ projectInfo.status === 1 ? '进行中' : '已结束' }}
            </n-tag>
          </n-space>
        </n-space>
      </n-card>
    </div>

    <n-card :bordered="false" class="mt-4 proCard">
      <n-tabs v-model:value="activeTab" type="line" animated>
        <!-- 看板视图 -->
        <n-tab-pane name="board" tab="任务看板">
          <n-spin :show="tasksLoading">
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
        </n-tab-pane>

        <!-- 任务列表 -->
        <n-tab-pane name="list" tab="任务列表">
          <n-space class="mb-4" align="center">
            <n-select
              v-model:value="taskFilter.status"
              :options="[{ label: '全部', value: '' }, ...boardColumns.map((c) => ({ label: c.label, value: c.status }))]"
              placeholder="状态筛选"
              style="width: 140px"
              clearable
              @update:value="loadTasks"
            />
            <n-select
              v-model:value="taskFilter.type"
              :options="[{ label: '全部', value: '' }, ...typeOptions]"
              placeholder="类型筛选"
              style="width: 140px"
              clearable
              @update:value="loadTasks"
            />
            <n-button type="primary" @click="showCreateTaskModal = true">新建任务</n-button>
          </n-space>
          <n-table :bordered="false" :single-line="false" size="small">
            <thead>
              <tr>
                <th>标题</th>
                <th>类型</th>
                <th>优先级</th>
                <th>状态</th>
                <th>指派人</th>
                <th>更新时间</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="task in filteredTaskList" :key="task.id">
                <td>
                  <n-button text type="info" @click="openTaskDetail(task)">{{ task.title }}</n-button>
                </td>
                <td><n-tag :type="typeColor[task.type] as any" size="small">{{ task.type }}</n-tag></td>
                <td><n-tag :type="priorityColor(task.priority) as any" size="small">P{{ task.priority }}</n-tag></td>
                <td><n-tag size="small">{{ statusLabel(task.status) }}</n-tag></td>
                <td>{{ task.assigneeName }}</td>
                <td>{{ task.updatedAt }}</td>
                <td>
                  <n-space size="small">
                    <n-button text type="info" @click="openTaskDetail(task)">详情</n-button>
                    <n-button text type="error" @click="handleDeleteTask(task)">删除</n-button>
                  </n-space>
                </td>
              </tr>
            </tbody>
          </n-table>
          <div class="mt-4 flex justify-end" v-if="taskTotal > taskPagination.size">
            <n-pagination
              v-model:page="taskPagination.page"
              :page-size="taskPagination.size"
              :item-count="taskTotal"
              @update:page="loadTasks"
            />
          </div>
        </n-tab-pane>

        <!-- Sprint -->
        <n-tab-pane name="sprint" tab="Sprint">
          <n-button type="primary" class="mb-4" @click="showCreateSprintModal = true">新建 Sprint</n-button>
          <n-spin :show="sprintsLoading">
            <n-empty v-if="!sprintsLoading && sprintList.length === 0" description="暂无 Sprint" />
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
                <tr v-for="sprint in sprintList" :key="sprint.id">
                  <td>{{ sprint.name }}</td>
                  <td>{{ sprint.goal }}</td>
                  <td>{{ sprint.startDate }}</td>
                  <td>{{ sprint.endDate }}</td>
                  <td>
                    <n-tag :type="sprintStatusColor[sprint.status] as any" size="small">
                      {{ sprintStatusLabel[sprint.status] || sprint.status }}
                    </n-tag>
                  </td>
                </tr>
              </tbody>
            </n-table>
          </n-spin>
        </n-tab-pane>

        <!-- 成员 -->
        <n-tab-pane name="members" tab="成员">
          <n-button type="primary" class="mb-4" @click="showAddMemberModal = true">添加成员</n-button>
          <n-spin :show="membersLoading">
            <n-empty v-if="!membersLoading && memberList.length === 0" description="暂无成员" />
            <n-table v-else :bordered="false" :single-line="false" size="small">
              <thead>
                <tr>
                  <th>用户名</th>
                  <th>姓名</th>
                  <th>角色</th>
                  <th>加入时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="member in memberList" :key="member.id">
                  <td>{{ member.username }}</td>
                  <td>{{ member.realName }}</td>
                  <td>{{ member.role }}</td>
                  <td>{{ member.joinedAt }}</td>
                  <td>
                    <n-button text type="error" @click="handleRemoveMember(member)">移除</n-button>
                  </td>
                </tr>
              </tbody>
            </n-table>
          </n-spin>
        </n-tab-pane>
      </n-tabs>
    </n-card>

    <!-- 新建任务弹窗 -->
    <n-modal
      v-model:show="showCreateTaskModal"
      preset="dialog"
      title="新建任务"
      positive-text="确定"
      negative-text="取消"
      @positive-click="handleCreateTask"
      style="width: 560px"
    >
      <n-form ref="taskFormRef" :model="taskFormData" :rules="taskFormRules" label-placement="left" :label-width="80" class="py-4">
        <n-form-item label="标题" path="title">
          <n-input v-model:value="taskFormData.title" placeholder="请输入任务标题" />
        </n-form-item>
        <n-form-item label="类型" path="type">
          <n-select v-model:value="taskFormData.type" :options="typeOptions" placeholder="请选择类型" />
        </n-form-item>
        <n-form-item label="优先级" path="priority">
          <n-input-number v-model:value="taskFormData.priority" :min="1" :max="4" placeholder="1=低 4=紧急" style="width: 100%" />
        </n-form-item>
        <n-form-item label="描述" path="description">
          <n-input v-model:value="taskFormData.description" type="textarea" placeholder="请输入描述" :rows="3" />
        </n-form-item>
      </n-form>
    </n-modal>

    <!-- 新建 Sprint 弹窗 -->
    <n-modal
      v-model:show="showCreateSprintModal"
      preset="dialog"
      title="新建 Sprint"
      positive-text="确定"
      negative-text="取消"
      @positive-click="handleCreateSprint"
      style="width: 520px"
    >
      <n-form ref="sprintFormRef" :model="sprintFormData" :rules="sprintFormRules" label-placement="left" :label-width="80" class="py-4">
        <n-form-item label="名称" path="name">
          <n-input v-model:value="sprintFormData.name" placeholder="请输入 Sprint 名称" />
        </n-form-item>
        <n-form-item label="目标" path="goal">
          <n-input v-model:value="sprintFormData.goal" type="textarea" placeholder="请输入目标" :rows="2" />
        </n-form-item>
        <n-form-item label="开始日期" path="startDate">
          <n-date-picker v-model:formatted-value="sprintFormData.startDate" type="date" placeholder="选择开始日期" style="width: 100%" />
        </n-form-item>
        <n-form-item label="结束日期" path="endDate">
          <n-date-picker v-model:formatted-value="sprintFormData.endDate" type="date" placeholder="选择结束日期" style="width: 100%" />
        </n-form-item>
      </n-form>
    </n-modal>

    <!-- 添加成员弹窗 -->
    <n-modal
      v-model:show="showAddMemberModal"
      preset="dialog"
      title="添加成员"
      positive-text="确定"
      negative-text="取消"
      @positive-click="handleAddMember"
      style="width: 420px"
    >
      <n-form ref="memberFormRef" :model="memberFormData" :rules="memberFormRules" label-placement="left" :label-width="80" class="py-4">
        <n-form-item label="用户ID" path="userId">
          <n-input-number v-model:value="memberFormData.userId" placeholder="请输入用户ID" style="width: 100%" />
        </n-form-item>
        <n-form-item label="角色" path="role">
          <n-input v-model:value="memberFormData.role" placeholder="请输入角色" />
        </n-form-item>
      </n-form>
    </n-modal>

    <!-- 任务详情弹窗 -->
    <TaskDetailModal ref="taskDetailRef" />
  </div>
</template>

<script lang="ts" setup>
  import { ref, reactive, computed, onMounted } from 'vue';
  import { useRouter, useRoute } from 'vue-router';
  import { useMessage, useDialog } from 'naive-ui';
  import { ArrowLeftOutlined } from '@vicons/antd';
  import {
    getProject,
    getTasks,
    createTask,
    deleteTask,
    getSprints,
    createSprint,
    getMembers,
    addMember,
    removeMember,
  } from '@/api/project/index';
  import type { ProjectItem, TaskItem, SprintItem, MemberItem } from '@/api/project/index';
  import TaskDetailModal from '@/views/project/components/TaskDetailModal.vue';

  const router = useRouter();
  const route = useRoute();
  const message = useMessage();
  const dialog = useDialog();

  const projectId = computed(() => Number(route.params.id));
  const activeTab = ref('board');

  // 项目信息
  const projectInfo = reactive<Partial<ProjectItem>>({});

  // 看板
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

  const typeOptions = [
    { label: 'Bug', value: 'bug' },
    { label: 'Feature', value: 'feature' },
    { label: 'Improvement', value: 'improvement' },
    { label: 'Task', value: 'task' },
  ];

  const statusLabels: Record<string, string> = {
    open: '待处理',
    in_progress: '进行中',
    review: '审核中',
    done: '已完成',
  };

  function statusLabel(status: string) {
    return statusLabels[status] || status;
  }

  // 任务
  const tasksLoading = ref(false);
  const taskList = ref<TaskItem[]>([]);
  const taskTotal = ref(0);
  const taskPagination = reactive({ page: 1, size: 20 });
  const taskFilter = reactive({ status: '', type: '' });

  const filteredTaskList = computed(() => {
    let list = taskList.value;
    if (taskFilter.status) {
      list = list.filter((t) => t.status === taskFilter.status);
    }
    if (taskFilter.type) {
      list = list.filter((t) => t.type === taskFilter.type);
    }
    return list;
  });

  function getTasksByStatus(status: string) {
    return taskList.value.filter((t) => t.status === status);
  }

  // Sprint
  const sprintsLoading = ref(false);
  const sprintList = ref<SprintItem[]>([]);
  const sprintStatusColor: Record<string, string> = { planning: 'default', active: 'info', completed: 'success' };
  const sprintStatusLabel: Record<string, string> = { planning: '规划中', active: '进行中', completed: '已完成' };

  // 成员
  const membersLoading = ref(false);
  const memberList = ref<MemberItem[]>([]);

  // 弹窗控制
  const showCreateTaskModal = ref(false);
  const showCreateSprintModal = ref(false);
  const showAddMemberModal = ref(false);
  const taskDetailRef = ref();

  // 任务表单
  const taskFormRef = ref<any>(null);
  const taskFormData = reactive({
    title: '',
    type: 'task',
    priority: 2,
    description: '',
  });
  const taskFormRules = {
    title: { required: true, message: '请输入任务标题', trigger: 'blur' },
  };

  // Sprint 表单
  const sprintFormRef = ref<any>(null);
  const sprintFormData = reactive({
    name: '',
    goal: '',
    startDate: '',
    endDate: '',
  });
  const sprintFormRules = {
    name: { required: true, message: '请输入 Sprint 名称', trigger: 'blur' },
  };

  // 成员表单
  const memberFormRef = ref<any>(null);
  const memberFormData = reactive({ userId: null as number | null, role: '' });
  const memberFormRules = {
    userId: { required: true, type: 'number', message: '请输入用户ID', trigger: 'blur' },
    role: { required: true, message: '请输入角色', trigger: 'blur' },
  };

  async function loadProject() {
    try {
      const res = await getProject(projectId.value);
      if (res) Object.assign(projectInfo, res);
    } catch (e) {
      // ignore
    }
  }

  async function loadTasks() {
    tasksLoading.value = true;
    try {
      const res = await getTasks(projectId.value, {
        page: taskPagination.page,
        size: taskPagination.size,
        status: taskFilter.status || undefined,
        type: taskFilter.type || undefined,
      });
      if (res) {
        taskList.value = res.list || [];
        taskTotal.value = res.total || 0;
      }
    } catch (e) {
      // ignore
    } finally {
      tasksLoading.value = false;
    }
  }

  async function loadSprints() {
    sprintsLoading.value = true;
    try {
      const res = await getSprints(projectId.value);
      sprintList.value = res?.list || [];
    } catch (e) {
      // ignore
    } finally {
      sprintsLoading.value = false;
    }
  }

  async function loadMembers() {
    membersLoading.value = true;
    try {
      const res = await getMembers(projectId.value);
      memberList.value = res?.list || [];
    } catch (e) {
      // ignore
    } finally {
      membersLoading.value = false;
    }
  }

  function openTaskDetail(task: TaskItem) {
    taskDetailRef.value?.openModal(task.id);
  }

  async function handleCreateTask() {
    try {
      await taskFormRef.value?.validate();
    } catch {
      return false;
    }
    try {
      await createTask(projectId.value, { ...taskFormData });
      message.success('任务创建成功');
      showCreateTaskModal.value = false;
      taskFormData.title = '';
      taskFormData.type = 'task';
      taskFormData.priority = 2;
      taskFormData.description = '';
      loadTasks();
    } catch (e) {
      message.error('创建失败');
      return false;
    }
  }

  function handleDeleteTask(task: TaskItem) {
    dialog.warning({
      title: '确认删除',
      content: `确定要删除任务「${task.title}」吗？`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await deleteTask(task.id);
          message.success('删除成功');
          loadTasks();
        } catch (e) {
          message.error('删除失败');
        }
      },
    });
  }

  async function handleCreateSprint() {
    try {
      await sprintFormRef.value?.validate();
    } catch {
      return false;
    }
    try {
      await createSprint(projectId.value, { ...sprintFormData });
      message.success('Sprint 创建成功');
      showCreateSprintModal.value = false;
      sprintFormData.name = '';
      sprintFormData.goal = '';
      sprintFormData.startDate = '';
      sprintFormData.endDate = '';
      loadSprints();
    } catch (e) {
      message.error('创建失败');
      return false;
    }
  }

  async function handleAddMember() {
    try {
      await memberFormRef.value?.validate();
    } catch {
      return false;
    }
    try {
      await addMember(projectId.value, { userId: memberFormData.userId!, role: memberFormData.role });
      message.success('成员添加成功');
      showAddMemberModal.value = false;
      memberFormData.userId = null;
      memberFormData.role = '';
      loadMembers();
    } catch (e) {
      message.error('添加失败');
      return false;
    }
  }

  function handleRemoveMember(member: MemberItem) {
    dialog.warning({
      title: '确认移除',
      content: `确定要移除成员「${member.username}」吗？`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await removeMember(projectId.value, member.userId);
          message.success('移除成功');
          loadMembers();
        } catch (e) {
          message.error('移除失败');
        }
      },
    });
  }

  onMounted(() => {
    loadProject();
    loadTasks();
    loadSprints();
    loadMembers();
  });
</script>
