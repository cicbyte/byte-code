<template>
  <div>
    <n-card :bordered="false" title="项目列表" class="proCard">
      <template #header-extra>
        <!-- 创建项目走权限字典 project_create（后端 CreateProject 同口径；
             role2 预绑，无权限者藏入口） -->
        <n-button v-if="canCreate" type="primary" @click="handleCreate">
          <template #icon>
            <n-icon><PlusOutlined /></n-icon>
          </template>
          新建项目
        </n-button>
      </template>

      <n-space class="mb-4" align="center">
        <n-input
          v-model:value="filter.keyword"
          placeholder="项目名称 / 描述"
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
        <n-button type="primary" @click="handleSearch">查询</n-button>
        <n-button @click="handleReset">重置</n-button>
      </n-space>

      <n-spin :show="loading">
        <EmptyState type="generic" title="还没有项目" v-if="!loading && projectList.length === 0" description="点击右上角「新建项目」开始" />
        <n-grid v-else cols="1 s:2 m:2 l:3 xl:4 2xl:4" responsive="screen" :x-gap="12" :y-gap="12">
          <n-grid-item v-for="item in projectList" :key="item.id">
            <n-card
              hoverable
              size="small"
              style="cursor: pointer"
              @click="handleDetail(item)"
              @contextmenu="onCtxMenu($event, item)"
            >
              <template #header>
                <span class="font-medium">{{ item.name }}</span>
                <n-tag size="tiny" :bordered="false" class="ml-2" style="font-family: monospace">{{ item.code }}</n-tag>
              </template>
              <template #header-extra>
                <n-tag :type="item.status === 1 ? 'success' : 'default'" size="small">
                  {{ PROJECT_STATUS.label(item.status) }}
                </n-tag>
              </template>
              <p class="text-gray-500 text-sm line-clamp-2">{{ item.description || '暂无描述' }}</p>
              <template #footer>
                <n-space justify="space-between" align="center">
                  <span class="text-gray-400 text-xs">负责人：{{ item.ownerName || item.creatorName }}</span>
                  <n-space size="small">
                    <n-button text type="info" size="small" @click.stop="handleEdit(item)">编辑</n-button>
                    <n-button text type="error" size="small" @click.stop="handleDelete(item)">删除</n-button>
                  </n-space>
                </n-space>
              </template>
            </n-card>
          </n-grid-item>
        </n-grid>
      </n-spin>

      <div class="mt-4 flex justify-end" v-if="total > pagination.size">
        <n-pagination
          v-model:page="pagination.page"
          v-model:page-size="pagination.size"
          @update:page-size="onPageSizeChange"
          :item-count="total"
          @update:page="loadData"
        />
      </div>
    </n-card>

    <!-- 右键快捷菜单（单例受控：x/y 定位，naive 手动模式标准写法） -->
    <n-dropdown
      trigger="manual"
      placement="bottom-start"
      :show="ctxMenuId !== null"
      :x="ctxMenuX"
      :y="ctxMenuY"
      :options="ctxOptions()"
      @select="(key: string | number) => onCtxSelect(key as string)"
      @clickoutside="ctxMenuId = null"
    />

    <!-- Agent 接入码（右键菜单直达）：一次性 24h，关闭即不再展示 -->
    <n-modal v-model:show="showAgentCode" preset="dialog" title="Agent 项目接入" :show-icon="false" style="width: 560px">
      <n-space vertical :size="10" class="py-2">
        <n-alert type="info" :show-icon="false">
          把以下信息交给 Agent 侧（或其驱动者），复制即用。接入码一次性、24 小时有效，关闭后不再展示。
        </n-alert>
        <div>
          <div class="text-xs mb-1">服务器地址（写入 ~/.bc/config.toml 的 server_url）</div>
          <n-space align="center" :size="8">
            <n-code :code="serverUrl" language="text" style="font-size: 12px" />
            <n-button text size="tiny" type="primary" @click="copyToClipboard(serverUrl).then(() => message.success('已复制'))">复制</n-button>
          </n-space>
        </div>
        <div>
          <div class="text-xs mb-1">项目接入码</div>
          <n-space align="center" :size="8">
            <n-code :code="agentCode" language="text" style="font-size: 12px" />
            <n-button text size="tiny" type="primary" @click="copyToClipboard(agentCode).then(() => message.success('已复制'))">复制</n-button>
          </n-space>
          <n-text depth="3" style="font-size: 12px">有效期至：{{ agentCodeExpires }}</n-text>
        </div>
        <div>
          <div class="text-xs mb-1">四步接入命令</div>
          <n-space align="flex-start" :size="8">
            <n-code :code="onboardCommands" language="bash" style="font-size: 12px; white-space: pre" />
            <n-button text size="tiny" type="primary" @click="copyToClipboard(onboardCommands).then(() => message.success('已复制'))">复制</n-button>
          </n-space>
        </div>
      </n-space>
    </n-modal>

    <!-- 转交负责人（邀请制）：右键菜单/成员页共用组件 -->
    <TransferDialog v-model:show="showTransferDialog" :project-id="transferProjectId" />

    <n-modal
      v-model:show="showModal"
      preset="dialog"
      :title="isEdit ? '编辑项目' : '新建项目'"
      positive-text="确定"
      negative-text="取消"
      @positive-click="handleSubmit"
      style="width: 520px"
    >
      <n-form ref="formRef" :model="formData" :rules="formRules" label-placement="left" :label-width="80" class="py-4">
        <n-form-item label="名称" path="name">
          <n-input v-model:value="formData.name" placeholder="请输入项目名称" />
        </n-form-item>
        <n-form-item label="描述" path="description">
          <n-input v-model:value="formData.description" type="textarea" placeholder="请输入项目描述" :rows="3" />
        </n-form-item>
      </n-form>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { PROJECT_STATUS } from '@/enums/entities';
  import { ref, reactive, computed, onMounted } from 'vue';
  import { useRouter } from 'vue-router';
  import { useMessage, useDialog } from 'naive-ui';
  import { PlusOutlined } from '@vicons/antd';
  import { useUserStore } from '@/store/modules/user';
  import { copyToClipboard } from '@/utils/clipboard';
  import { createAgentJoinCode } from '@/api/agent/index';
  import { usePerm } from '@/composables/usePerm';
  import TransferDialog from '@/views/project/components/TransferDialog.vue';
  import type { DropdownOption } from 'naive-ui';
  import {
    getProjects,
    createProject,
    updateProject,
    deleteProject,
  } from '@/api/project/index';
  import type { ProjectItem } from '@/api/project/index';

  const router = useRouter();
  const message = useMessage();
  const dialog = useDialog();
  const userStore = useUserStore();
  const { has, isAdmin } = usePerm();
  const canCreate = computed(() => has('project_create'));
  const loading = ref(false);
  const projectList = ref<ProjectItem[]>([]);
  const total = ref(0);
  const showModal = ref(false);
  const isEdit = ref(false);
  const editId = ref<number | null>(null);
  const formRef = ref<any>(null);

  const pagination = reactive({ page: 1, size: 12 });
  const filter = reactive({ keyword: '', status: null as number | null });
  const statusOptions = PROJECT_STATUS.options;

  const formData = reactive({
    name: '',
    description: '',
  });

  const formRules = {
    name: { required: true, message: '请输入项目名称', trigger: 'blur' },
  };

  function resetForm() {
    formData.name = '';
    formData.description = '';
    isEdit.value = false;
    editId.value = null;
  }

  // 每页条数变化回到第 1 页（改大 size 后当前页可能越界；
  // 此前绑定的 loadProjects 未定义属潜伏运行时错误）
  function onPageSizeChange() {
    pagination.page = 1;
    loadData();
  }

  async function loadData() {
    loading.value = true;
    try {
      const res = await getProjects({
        page: pagination.page,
        size: pagination.size,
        keyword: filter.keyword || undefined,
        status: filter.status ?? undefined,
      });
      if (res) {
        projectList.value = res.list || [];
        total.value = res.total || 0;
      }
    } finally {
      loading.value = false;
    }
  }

  function handleSearch() {
    pagination.page = 1;
    loadData();
  }

  function handleReset() {
    filter.keyword = '';
    filter.status = null;
    pagination.page = 1;
    loadData();
  }

  function handleDetail(item: ProjectItem) {
    router.push(`/project/${item.id}/board`);
  }

  // ---- 右键快捷菜单：导航 + 移交（邀请制） ----
  const ctxMenuId = ref<number | null>(null);
  const ctxMenuX = ref(0);
  const ctxMenuY = ref(0);
  const ctxItem = ref<ProjectItem | null>(null);
  const showTransferDialog = ref(false);
  const transferProjectId = ref(0);
  const canTransfer = computed(() => isAdmin.value);

  function onCtxMenu(e: MouseEvent, item: ProjectItem) {
    e.preventDefault();
    ctxItem.value = item;
    ctxMenuId.value = item.id;
    ctxMenuX.value = e.clientX;
    ctxMenuY.value = e.clientY;
  }

  function ctxOptions(): DropdownOption[] {
    const item = ctxItem.value;
    if (!item) return [];
    const ops: DropdownOption[] = [
      { label: '打开看板', key: 'board' },
      { label: '任务', key: 'tasks' },
      { label: '知识库', key: 'docs' },
      { label: '成员管理', key: 'members' },
      { label: 'Agent 接入码…', key: 'agent-code' },
      { label: '项目设置', key: 'settings' },
      { label: `复制短码 ${item.code}`, key: 'copy-code' },
    ];
    // 移交：现任 owner（移交后即换人）或平台管理员；后端同口径兜底。
    // ownerId 缺失（无 owner 行）回退 creator
    const myId = Number((userStore?.info as any)?.userId || 0);
    const owner = item.ownerId || item.creatorId;
    if (canTransfer.value || owner === myId) {
      ops.push({ label: '移交负责人…', key: 'transfer' });
    }
    return ops;
  }

  // ---- Agent 接入码（右键菜单直达；与成员页同款展示与复制） ----
  const showAgentCode = ref(false);
  const agentCode = ref('');
  const agentCodeExpires = ref('');
  const serverUrl = computed(() => window.location.origin + '/api');
  const onboardCommands = computed(() =>
    [
      '# 1. 配置服务器（一次性，写入 ~/.bc/config.toml）',
      'bcode config set server ' + serverUrl.value,
      '',
      '# 2. 注册身份（bc_ key 自动落本地）',
      'bcode register <agent-name>',
      '',
      '# 3. 凭码加入本项目',
      'bcode join ' + (agentCode.value || '<接入码>'),
      '',
      '# 4. 在项目目录建立会话（开工包）',
      'bcode start',
    ].join('\n')
  );

  async function handleGenAgentCode(item: ProjectItem) {
    try {
      const res = await createAgentJoinCode(item.id);
      agentCode.value = res.code;
      agentCodeExpires.value = res.expiresAt;
      showAgentCode.value = true;
    } catch {
      // http 层统一提示（无管理权限等）
    }
  }

  function onCtxSelect(key: string) {
    const item = ctxItem.value;
    ctxMenuId.value = null;
    if (!item) return;
    switch (key) {
      case 'board':
        router.push(`/project/${item.id}/board`);
        break;
      case 'tasks':
        router.push(`/project/${item.id}/tasks`);
        break;
      case 'docs':
        router.push(`/project/${item.id}/docs`);
        break;
      case 'members':
        router.push(`/project/${item.id}/members`);
        break;
      case 'settings':
        router.push(`/project/${item.id}/settings`);
        break;
      case 'agent-code':
        handleGenAgentCode(item);
        break;
      case 'copy-code':
        copyToClipboard(item.code).then(
          () => message.success(`已复制短码：${item.code}`),
          () => message.error('复制失败')
        );
        break;
      case 'transfer':
        transferProjectId.value = item.id;
        showTransferDialog.value = true;
        break;
    }
  }

  function handleCreate() {
    resetForm();
    showModal.value = true;
  }

  function handleEdit(item: ProjectItem) {
    isEdit.value = true;
    editId.value = item.id;
    formData.name = item.name;
    formData.description = item.description;
    showModal.value = true;
  }

  function handleDelete(item: ProjectItem) {
    dialog.warning({
      title: '确认删除',
      content: `确定要删除项目「${item.name}」吗？`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await deleteProject(item.id);
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
        await updateProject(editId.value, { ...formData });
        message.success('更新成功');
      } else {
        await createProject({ ...formData });
        message.success('创建成功');
      }
      showModal.value = false;
      loadData();
    } catch (e) {
      message.error('操作失败');
      return false;
    }
  }

  onMounted(() => {
    loadData();
  });
</script>
