<template>
  <div>
    <!-- 关联项目（owner/maintainer 治理反馈投递面，PRD §4.1） -->
    <n-card title="关联项目" :bordered="false" class="proCard">
      <template #header-extra>
        <n-button v-if="canGovern" size="small" @click="showRelation = true">添加关联</n-button>
      </template>
      <n-spin :show="relationsLoading">
        <EmptyState v-if="!relationsLoading && relations.length === 0" type="member" title="未关联其他项目" description="建立关联后可互相投递跨项目反馈（线索由对方 Agent 分析是否建任务）" compact />
        <n-table v-else :bordered="false" :single-line="false" size="small">
          <thead><tr><th>项目</th><th>关联时间</th><th style="width: 80px">操作</th></tr></thead>
          <tbody>
            <tr v-for="r in relations" :key="r.id">
              <td>{{ r.name }}（#{{ r.projectId }}）</td>
              <td>{{ (r.createdAt || '').slice(0, 16) }}</td>
              <td><n-button v-if="canGovern" text type="error" size="small" @click="handleRemoveRelation(r)">移除</n-button></td>
            </tr>
          </tbody>
        </n-table>
      </n-spin>
    </n-card>

    <!-- 项目分组（owner/maintainer 治理） -->
    <n-card title="项目分组" :bordered="false" class="mt-4">
      <template #header-extra>
        <n-button v-if="canGovern" size="small" @click="showGroupAdd = true">加入分组</n-button>
      </template>
      <n-spin :show="groupsLoading">
        <EmptyState v-if="!groupsLoading && myGroups.length === 0" type="generic" title="未加入任何分组" description="同分组的项目自动互为关联（反馈/引用免准入），比手动建关联更省事" compact />
        <n-space v-else :size="8">
          <n-tag v-for="g in myGroups" :key="g.id" :closable="canGovern" size="small" type="success" @close="handleRemoveGroup(g)">
            {{ g.name }}
            <span class="text-xs opacity-60 ml-1">({{ (g.projects || []).length }}个项目)</span>
          </n-tag>
        </n-space>
      </n-spin>
    </n-card>

    <!-- Danger Zone：归档/解档 + 删除（owner/超管可见；转交在成员页） -->
    <n-card title="危险操作" :bordered="false" class="mt-4">
      <template #header-extra><n-tag size="small" :bordered="false" type="warning">owner 专属</n-tag></template>
      <n-space vertical :size="12">
        <n-alert type="info" :show-icon="false">
          归档后项目对成员与 Agent 只读（任务/文档/记忆均不可写），你仍可编辑与解档；删除为硬删除且连带全部数据，不可恢复。
        </n-alert>
        <n-space>
          <n-button v-if="projectStatus !== 3" type="warning" secondary @click="handleArchive(true)">归档项目</n-button>
          <n-button v-else type="success" secondary @click="handleArchive(false)">解除归档</n-button>
          <n-button type="error" secondary @click="handleDeleteProject">删除项目</n-button>
        </n-space>
      </n-space>
    </n-card>

    <!-- 添加关联弹窗 -->
    <n-modal v-model:show="showRelation" preset="dialog" title="添加关联项目" :show-icon="false">
      <n-select v-model:value="relationTarget" :options="relationOptions" placeholder="选择项目（须为你可访问的项目）" size="small" />
      <template #action>
        <n-space>
          <n-button size="small" @click="showRelation = false">取消</n-button>
          <n-button size="small" type="primary" :disabled="!relationTarget" @click="handleAddRelation">关联</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 加入分组弹窗 -->
    <n-modal v-model:show="showGroupAdd" preset="dialog" title="加入分组" :show-icon="false">
      <n-radio-group v-model:value="groupAddMode" size="small" class="mb-3">
        <n-radio-button value="existing">加入已有分组</n-radio-button>
        <!-- 新建分组定义属平台级权限（platform_groups），未授予者只加入已有分组 -->
        <n-radio-button v-if="canCreateGroup" value="new">新建分组</n-radio-button>
      </n-radio-group>
      <n-select
        v-if="groupAddMode === 'existing'"
        v-model:value="groupTarget"
        :options="groupOptions"
        placeholder="选择分组"
        size="small"
      />
      <n-input v-else v-model:value="newGroupName" placeholder="分组名（如 cicbyte 生态）" size="small" />
      <template #action>
        <n-space>
          <n-button size="small" @click="showGroupAdd = false">取消</n-button>
          <n-button size="small" type="primary" :disabled="groupAddMode === 'existing' ? !groupTarget : !newGroupName.trim()" :loading="groupAdding" @click="handleAddGroup">确认</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { ref, reactive, computed, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { useMessage, useDialog } from 'naive-ui';
  import { useRouter } from 'vue-router';
  import { getRelations, addRelation, removeRelation, getProjects, getMembers, updateProject, deleteProject } from '@/api/project/index';
  import { getGroups, createGroup, addProjectToGroup, removeProjectFromGroup } from '@/api/project/group';
  import type { GroupItem } from '@/api/project/group';
  import { useUserStore } from '@/store/modules/user';
  import { usePerm } from '@/composables/usePerm';

  const route = useRoute();
  const router = useRouter();
  const message = useMessage();
  const dialog = useDialog();
  const userStore = useUserStore();
  const { has, isAdmin } = usePerm();

  const projectId = computed(() => Number(route.params.projectId));

  // ==================== 治理权限（PRD §4.1：关联/分组治理 owner+maintainer） ====================
  const myRole = ref('');
  const canGovern = computed(() => ['owner', 'maintainer'].includes(myRole.value) || isAdmin.value);
  // 新建分组定义走平台字典 platform_groups（P1-2 后端同口径）
  const canCreateGroup = computed(() => isAdmin.value || has('platform_groups'));

  // ==================== Danger Zone（归档/删除） ====================
  const projectStatus = ref<number>(1);
  async function loadProjectStatus() {
    try {
      const res = await getProjects({ keyword: '', status: null as any });
      const cur = (res?.list || []).find((p: any) => p.id === projectId.value);
      projectStatus.value = cur?.status ?? 1;
    } catch { /* 列表接口口径兜底 */ }
  }

  function handleArchive(archive: boolean) {
    dialog.warning({
      title: archive ? '归档项目' : '解除归档',
      content: archive
        ? '归档后成员与 Agent 只读（含任务认领/文档写入），确定归档吗？'
        : '解档后项目恢复全员可写，确定吗？',
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await updateProject(projectId.value, { status: archive ? 3 : 1 });
          message.success(archive ? '项目已归档' : '已解除归档');
          loadProjectStatus();
        } catch (e: any) {
          message.error(e?.message || '操作失败');
        }
      },
    });
  }

  function handleDeleteProject() {
    dialog.warning({
      title: '删除项目',
      content: '删除将连带清除任务/文档/记忆/成员关系等全部数据，且不可恢复。确定删除吗？',
      positiveText: '确认删除',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await deleteProject(projectId.value);
          message.success('项目已删除');
          router.push('/project/list');
        } catch (e: any) {
          message.error(e?.message || '删除失败');
        }
      },
    });
  }

  async function loadMyRole() {
    try {
      const res = await getMembers(projectId.value);
      const myId = Number((userStore?.info as any)?.userId || 0);
      const meRow = (res?.list || []).find((m: any) => m.userType !== 'ai' && m.userId === myId);
      myRole.value = meRow?.role || '';
    } catch { myRole.value = ''; }
  }

  // ==================== 关联项目 ====================
  const relationsLoading = ref(false);
  const relations = ref<any[]>([]);
  const showRelation = ref(false);
  const relationTarget = ref<number | null>(null);
  const relationOptions = ref<Array<{ label: string; value: number }>>([]);

  async function loadRelations() {
    relationsLoading.value = true;
    try {
      const [relRes, projRes] = await Promise.all([getRelations(projectId.value), getProjects({ size: 100 })]);
      relations.value = relRes?.list || [];
      const relatedIds = new Set((relRes?.list || []).map((r: any) => r.projectId));
      relationOptions.value = (projRes?.list || [])
        .filter((p: any) => p.id !== projectId.value && !relatedIds.has(p.id))
        .map((p: any) => ({ label: p.name, value: p.id }));
    } catch { /* ignore */ }
    finally { relationsLoading.value = false; }
  }

  async function handleAddRelation() {
    if (!relationTarget.value) return;
    try {
      await addRelation(projectId.value, { relatedProjectId: relationTarget.value });
      message.success('关联成功');
      showRelation.value = false;
      relationTarget.value = null;
      loadRelations();
    } catch (e: any) {
      message.error(e?.message || '关联失败');
    }
  }

  async function handleRemoveRelation(r: any) {
    try {
      await removeRelation(projectId.value, r.projectId);
      message.success('已移除关联');
      loadRelations();
    } catch (e: any) {
      message.error(e?.message || '移除失败');
    }
  }

  // ==================== 项目分组 ====================
  const groupsLoading = ref(false);
  const allGroups = ref<GroupItem[]>([]);
  const myGroups = computed(() =>
    allGroups.value.filter((g) => (g.projects || []).some((p) => p.id === projectId.value))
  );
  const groupOptions = computed(() =>
    allGroups.value
      .filter((g) => !(g.projects || []).some((p) => p.id === projectId.value))
      .map((g) => ({ label: `${g.name}（${(g.projects || []).length}个项目）`, value: g.id }))
  );
  const showGroupAdd = ref(false);
  const groupAddMode = ref<'existing' | 'new'>('existing');
  const groupTarget = ref<number | null>(null);
  const newGroupName = ref('');
  const groupAdding = ref(false);

  async function loadGroups() {
    groupsLoading.value = true;
    try {
      const res = await getGroups();
      allGroups.value = res?.list || [];
    } catch (e: any) {
      message.error(e?.message || '加载分组失败');
    } finally { groupsLoading.value = false; }
  }

  async function handleAddGroup() {
    groupAdding.value = true;
    try {
      let gid = groupTarget.value;
      if (groupAddMode.value === 'new') {
        const res = await createGroup({ name: newGroupName.value.trim() });
        gid = res?.id || null;
      }
      if (gid) {
        await addProjectToGroup(gid, projectId.value);
        message.success('已加入分组');
        showGroupAdd.value = false;
        groupTarget.value = null;
        newGroupName.value = '';
        loadGroups();
      }
    } catch (e: any) {
      message.error(e?.message || '操作失败');
    } finally { groupAdding.value = false; }
  }

  async function handleRemoveGroup(g: GroupItem) {
    try {
      await removeProjectFromGroup(g.id, projectId.value);
      message.success(`已从「${g.name}」移除`);
      loadGroups();
    } catch (e: any) {
      message.error(e?.message || '移除失败');
    }
  }

  onMounted(() => {
    loadMyRole();
    loadProjectStatus();
    loadRelations();
    loadGroups();
  });
</script>
