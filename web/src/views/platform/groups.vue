<template>
  <div>
    <n-card :bordered="false" title="项目分组" class="proCard">
      <template #header-extra>
        <n-button type="primary" @click="openCreate">
          <template #icon><n-icon><PlusOutlined /></n-icon></template>
          新建分组
        </n-button>
      </template>

      <n-spin :show="loading">
        <EmptyState
          v-if="!loading && groups.length === 0"
          type="generic"
          title="还没有分组"
          description="创建分组并把相关项目加进来，同分组的项目自动互为关联（反馈/引用免手动建关联）"
        />
        <div v-else class="group-list">
          <div v-for="g in groups" :key="g.id" class="group-card">
            <div class="group-head">
              <span class="group-name">{{ g.name }}</span>
              <n-tag size="tiny" :bordered="false">{{ (g.projects || []).length }} 个项目</n-tag>
              <n-space size="small" class="ml-auto">
                <n-button text type="info" size="small" @click="openAddMember(g)">添加项目</n-button>
                <n-button text type="warning" size="small" @click="openEdit(g)">编辑</n-button>
                <n-popconfirm @positive-click="handleDelete(g)">
                  <template #trigger>
                    <n-button text type="error" size="small">删除</n-button>
                  </template>
                  删除分组「{{ g.name }}」后成员关系一并清除，同分组的隐式关联立即失效。
                </n-popconfirm>
              </n-space>
            </div>
            <div v-if="g.description" class="group-desc">{{ g.description }}</div>
            <div v-if="(g.projects || []).length" class="group-projects">
              <n-tag
                v-for="p in g.projects"
                :key="p.id"
                size="small"
                closable
                @close="handleRemoveMember(g, p)"
              >
                {{ p.name }}
                <span class="text-xs opacity-50 ml-1">{{ p.code }}</span>
              </n-tag>
            </div>
            <div v-else class="group-empty">暂无成员项目</div>
          </div>
        </div>
      </n-spin>
    </n-card>

    <!-- 新建/编辑分组 -->
    <n-modal v-model:show="showForm" preset="dialog" :title="isEdit ? '编辑分组' : '新建分组'" :show-icon="false" style="width: 440px">
      <n-form label-placement="left" label-width="70" class="py-3">
        <n-form-item label="分组名" required>
          <n-input v-model:value="form.name" placeholder="如 cicbyte 生态" :disabled="isEdit" />
        </n-form-item>
        <n-form-item label="描述">
          <n-input v-model:value="form.description" type="textarea" placeholder="这个分组是什么用途" :rows="2" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button size="small" @click="showForm = false">取消</n-button>
          <n-button size="small" type="primary" :disabled="!form.name.trim()" :loading="submitting" @click="handleSubmit">保存</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 添加项目到分组 -->
    <n-modal v-model:show="showMember" preset="dialog" :title="`添加项目到「${currentGroup?.name}」`" :show-icon="false" style="width: 440px">
      <n-select
        v-model:value="memberTarget"
        :options="memberOptions"
        placeholder="选择项目"
        size="small"
        filterable
      />
      <template #action>
        <n-space>
          <n-button size="small" @click="showMember = false">取消</n-button>
          <n-button size="small" type="primary" :disabled="!memberTarget" :loading="submitting" @click="handleAddMember">添加</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { ref, reactive, computed, onMounted } from 'vue';
  import { useMessage } from 'naive-ui';
  import { PlusOutlined } from '@vicons/antd';
  import {
    getGroups,
    createGroup,
    updateGroup,
    deleteGroup,
    addProjectToGroup,
    removeProjectFromGroup,
  } from '@/api/project/group';
  import type { GroupItem } from '@/api/project/group';
  import { getProjects } from '@/api/project/index';

  const message = useMessage();
  const loading = ref(false);
  const submitting = ref(false);
  const groups = ref<GroupItem[]>([]);
  const allProjects = ref<any[]>([]);

  const showForm = ref(false);
  const isEdit = ref(false);
  const editId = ref<number | null>(null);
  const form = reactive({ name: '', description: '' });

  const showMember = ref(false);
  const currentGroup = ref<GroupItem | null>(null);
  const memberTarget = ref<number | null>(null);
  const memberOptions = computed(() => {
    if (!currentGroup.value) return [];
    const inGroup = new Set((currentGroup.value.projects || []).map((p) => p.id));
    return allProjects.value
      .filter((p) => !inGroup.has(p.id))
      .map((p) => ({ label: `${p.name}（${p.code}）`, value: p.id }));
  });

  async function load() {
    loading.value = true;
    try {
      const [gRes, pRes] = await Promise.all([getGroups(), getProjects({ size: 100 })]);
      groups.value = gRes?.list || [];
      allProjects.value = pRes?.list || [];
    } catch {
      // ignore
    } finally {
      loading.value = false;
    }
  }

  function openCreate() {
    isEdit.value = false;
    editId.value = null;
    form.name = '';
    form.description = '';
    showForm.value = true;
  }

  function openEdit(g: GroupItem) {
    isEdit.value = true;
    editId.value = g.id;
    form.name = g.name;
    form.description = g.description || '';
    showForm.value = true;
  }

  async function handleSubmit() {
    if (!form.name.trim()) return;
    submitting.value = true;
    try {
      if (isEdit.value && editId.value) {
        await updateGroup(editId.value, { description: form.description });
        message.success('更新成功');
      } else {
        await createGroup({ name: form.name.trim(), description: form.description || undefined });
        message.success('创建成功');
      }
      showForm.value = false;
      load();
    } catch (e: any) {
      message.error(e?.message || '操作失败');
    } finally {
      submitting.value = false;
    }
  }

  function openAddMember(g: GroupItem) {
    currentGroup.value = g;
    memberTarget.value = null;
    showMember.value = true;
  }

  async function handleAddMember() {
    if (!currentGroup.value || !memberTarget.value) return;
    submitting.value = true;
    try {
      await addProjectToGroup(currentGroup.value.id, memberTarget.value);
      message.success('已添加');
      showMember.value = false;
      load();
    } catch (e: any) {
      message.error(e?.message || '添加失败');
    } finally {
      submitting.value = false;
    }
  }

  async function handleRemoveMember(g: GroupItem, p: any) {
    try {
      await removeProjectFromGroup(g.id, p.id);
      message.success(`已从「${g.name}」移除 ${p.name}`);
      load();
    } catch (e: any) {
      message.error(e?.message || '移除失败');
    }
  }

  async function handleDelete(g: GroupItem) {
    try {
      await deleteGroup(g.id);
      message.success(`分组「${g.name}」已删除`);
      load();
    } catch (e: any) {
      message.error(e?.message || '删除失败');
    }
  }

  onMounted(load);
</script>

<style lang="less" scoped>
  .group-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .group-card {
    padding: 14px 16px;
    border: 1px solid var(--line, #e9e9e7);
    border-radius: 8px;

    &:hover {
      border-color: var(--primary-color, #16a34a);
    }
  }

  .group-head {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .group-name {
    font-weight: 600;
    font-size: 15px;
  }

  .group-desc {
    margin-top: 6px;
    color: var(--text-3, #8b949e);
    font-size: 13px;
  }

  .group-projects {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin-top: 10px;
  }

  .group-empty {
    margin-top: 10px;
    color: var(--text-3, #8b949e);
    font-size: 13px;
  }
</style>
