<template>
  <div>
    <!-- 关联项目（owner 治理反馈投递面） -->
    <n-card title="关联项目" :bordered="false" class="proCard">
      <template #header-extra>
        <n-button size="small" @click="showRelation = true">添加关联</n-button>
      </template>
      <n-spin :show="relationsLoading">
        <EmptyState v-if="!relationsLoading && relations.length === 0" type="member" title="未关联其他项目" description="建立关联后可互相投递跨项目反馈（线索由对方 Agent 分析是否建任务）" compact />
        <n-table v-else :bordered="false" :single-line="false" size="small">
          <thead><tr><th>项目</th><th>关联时间</th><th style="width: 80px">操作</th></tr></thead>
          <tbody>
            <tr v-for="r in relations" :key="r.id">
              <td>{{ r.name }}（#{{ r.projectId }}）</td>
              <td>{{ (r.createdAt || '').slice(0, 16) }}</td>
              <td><n-button text type="error" size="small" @click="handleRemoveRelation(r)">移除</n-button></td>
            </tr>
          </tbody>
        </n-table>
      </n-spin>
    </n-card>

    <!-- 项目分组 -->
    <n-card title="项目分组" :bordered="false" class="mt-4">
      <template #header-extra>
        <n-button size="small" @click="showGroupAdd = true">加入分组</n-button>
      </template>
      <n-spin :show="groupsLoading">
        <EmptyState v-if="!groupsLoading && myGroups.length === 0" type="generic" title="未加入任何分组" description="同分组的项目自动互为关联（反馈/引用免准入），比手动建关联更省事" compact />
        <n-space v-else :size="8">
          <n-tag v-for="g in myGroups" :key="g.id" closable size="small" type="success" @close="handleRemoveGroup(g)">
            {{ g.name }}
            <span class="text-xs opacity-60 ml-1">({{ (g.projects || []).length }}个项目)</span>
          </n-tag>
        </n-space>
      </n-spin>
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
        <n-radio-button value="new">新建分组</n-radio-button>
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
  import { useMessage } from 'naive-ui';
  import { getRelations, addRelation, removeRelation, getProjects } from '@/api/project/index';
  import { getGroups, createGroup, addProjectToGroup, removeProjectFromGroup } from '@/api/project/group';
  import type { GroupItem } from '@/api/project/group';

  const route = useRoute();
  const message = useMessage();

  const projectId = computed(() => Number(route.params.projectId));

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
    loadRelations();
    loadGroups();
  });
</script>
