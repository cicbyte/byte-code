<template>
  <div>
    <n-card :bordered="false" title="角色权限" class="proCard">
      <template #header-extra>
        <n-button type="primary" @click="openCreate">
          <template #icon>
            <n-icon><PlusOutlined /></n-icon>
          </template>
          新增角色
        </n-button>
      </template>

      <n-spin :show="loading">
        <EmptyState
          v-if="!loading && roles.length === 0"
          type="generic"
          title="暂无角色"
          description="新增角色后可为其分配可见菜单，再到用户管理里绑定用户"
          compact
        />
        <n-table v-else :bordered="false" :single-line="false" size="small">
          <thead>
            <tr>
              <th class="col-idx">#</th>
              <th>角色名称</th>
              <th>说明</th>
              <th>权限范围</th>
              <th>创建时间</th>
              <th style="width: 180px">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(item, __ix) in roles" :key="item.id">
              <td class="col-idx">{{ __ix + 1 }}</td>
              <td class="font-medium">
                {{ item.name }}
                <n-tag v-if="Number(item.id) <= 2" size="tiny" :bordered="false" class="ml-1">内置</n-tag>
              </td>
              <td>{{ item.explain || '-' }}</td>
              <td>
                <n-space :size="4" v-if="scopeTitles(item).length">
                  <n-tag v-for="t in scopeTitles(item)" :key="t" size="small" :bordered="false">{{ t }}</n-tag>
                </n-space>
                <span v-else class="text-gray-400">无（仅个人可见页面）</span>
              </td>
              <td class="text-xs text-gray-400">{{ (item.create_date || item.createDate || '').slice(0, 10) }}</td>
              <td>
                <n-space size="small">
                  <n-button v-if="Number(item.id) !== 1" text type="info" size="small" @click="openMenuAuth(item)">
                    菜单权限
                  </n-button>
                  <n-button text type="primary" size="small" @click="openEdit(item)">编辑</n-button>
                  <n-button v-if="Number(item.id) > 2" text type="error" size="small" @click="handleDelete(item)">
                    删除
                  </n-button>
                </n-space>
              </td>
            </tr>
          </tbody>
        </n-table>
      </n-spin>
    </n-card>

    <!-- 新建/编辑角色 -->
    <n-modal
      v-model:show="showForm"
      preset="dialog"
      :title="isEdit ? '编辑角色' : '新增角色'"
      positive-text="保存"
      negative-text="取消"
      @positive-click="submitForm"
      style="width: 480px"
    >
      <n-form ref="formRef" :model="form" :rules="formRules" label-placement="left" label-width="80" class="py-4">
        <n-form-item label="角色名称" path="name">
          <n-input v-model:value="form.name" placeholder="请输入角色名称" :disabled="isEdit && editId !== null && editId <= 2" />
        </n-form-item>
        <n-form-item label="角色说明" path="explain">
          <n-input v-model:value="form.explain" type="textarea" placeholder="请输入角色说明" :rows="2" />
        </n-form-item>
      </n-form>
    </n-modal>

    <!-- 菜单权限分配 -->
    <n-modal
      v-model:show="showMenuAuth"
      preset="dialog"
      :title="`分配 ${currentRole?.name} 的菜单权限`"
      positive-text="保存"
      negative-text="取消"
      @positive-click="submitMenuAuth"
      style="width: 520px"
    >
      <n-alert type="info" :bordered="false" class="mt-2 mb-3">
        勾选的菜单对该角色可见；个人可见页面（仪表盘/项目/我的任务/通知等）不受此控制。变更对该角色下的用户在下次刷新页面时生效。
      </n-alert>
      <div class="menu-tree-wrap">
        <n-tree
          block-line
          cascade
          checkable
          :data="menuTree"
          :expanded-keys="expandedKeys"
          :checked-keys="checkedKeys"
          @update:checked-keys="(keys: string[]) => (checkedKeys = keys)"
          @update:expanded-keys="(keys: string[]) => (expandedKeys = keys)"
        />
      </div>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { ref, reactive, computed, onMounted } from 'vue';
  import { useMessage, useDialog } from 'naive-ui';
  import { PlusOutlined } from '@vicons/antd';
  import {
    getRoleList,
    createRole,
    updateRole,
    deleteRole,
    updateRoleMenus,
    getMenuList,
  } from '@/api/system/role';
  import type { RoleItem, MenuDictItem } from '@/api/system/role';

  const message = useMessage();
  const dialog = useDialog();

  const loading = ref(false);
  const submitting = ref(false);
  const roles = ref<RoleItem[]>([]);

  // 菜单字典树（后端 sys_menus，key 用 id 字符串与存储对齐）
  const menuTree = ref<MenuDictItem[]>([]);
  const idTitle = ref<Map<number, string>>(new Map());
  const parentIds = ref<Set<number>>(new Set());
  const leafIds = ref<Set<number>>(new Set());

  // 角色表单
  const showForm = ref(false);
  const isEdit = ref(false);
  const editId = ref<number | null>(null);
  const formRef = ref();
  const form = reactive({ name: '', explain: '' });
  const formRules = {
    name: [
      { required: true, message: '请输入角色名称', trigger: 'blur' },
      { min: 2, max: 30, message: '角色名长度2-30位', trigger: 'blur' },
    ],
  };

  // 菜单权限弹窗
  const showMenuAuth = ref(false);
  const currentRole = ref<RoleItem | null>(null);
  const checkedKeys = ref<string[]>([]);
  const expandedKeys = ref<string[]>([]);

  function collectDict(nodes: MenuDictItem[]) {
    const parents = new Set<number>();
    const leaves = new Set<number>();
    const titles = new Map<number, string>();
    const walk = (list: MenuDictItem[]) => {
      for (const n of list) {
        titles.set(n.id, n.label);
        if (n.children?.length) {
          parents.add(n.id);
          walk(n.children);
        } else {
          leaves.add(n.id);
        }
      }
    };
    walk(nodes);
    parentIds.value = parents;
    leafIds.value = leaves;
    idTitle.value = titles;
  }

  async function loadMenus() {
    try {
      const res = await getMenuList();
      const raw = res?.list || [];
      // 字典节点必须带数值 id（树的 key 与存储主键都靠它）；缺失说明后端是旧版本，
      // 直接放行会让所有节点 key 相同（表现为点一个子节点全部勾选）
      const hasValidIds = (nodes: MenuDictItem[]): boolean =>
        nodes.every((n) => typeof n.id === 'number' && (!n.children?.length || hasValidIds(n.children)));
      if (!hasValidIds(raw)) {
        menuTree.value = [];
        message.warning('菜单字典异常（后端版本过旧），请重启后端后刷新页面');
        return;
      }
      const toTree = (list: MenuDictItem[]): MenuDictItem[] =>
        (list || []).map((n) => ({
          id: n.id,
          label: n.label,
          key: String(n.id),
          children: n.children?.length ? toTree(n.children) : undefined,
        }));
      menuTree.value = toTree(raw);
      collectDict(menuTree.value);
    } catch {
      // 字典加载失败不阻塞角色列表；分配入口将不可保存
    }
  }

  async function loadRoles() {
    loading.value = true;
    try {
      const res = await getRoleList({ pageNum: 1, pageSize: 100 });
      roles.value = res?.list || [];
    } catch {
      // ignore
    } finally {
      loading.value = false;
    }
  }

  // 权限范围列：只展示叶子菜单标题（父节点随子级级联，单独列出是噪音）
  function scopeTitles(item: RoleItem): string[] {
    return (item.menu_ids || [])
      .filter((id) => leafIds.value.has(id))
      .map((id) => idTitle.value.get(id) || `#${id}`);
  }

  function openCreate() {
    isEdit.value = false;
    editId.value = null;
    form.name = '';
    form.explain = '';
    showForm.value = true;
  }

  function openEdit(item: RoleItem) {
    isEdit.value = true;
    editId.value = Number(item.id);
    form.name = item.name;
    form.explain = item.explain || '';
    showForm.value = true;
  }

  async function submitForm() {
    try {
      await formRef.value?.validate();
    } catch {
      return false;
    }
    if (submitting.value) return false;
    submitting.value = true;
    try {
      if (isEdit.value && editId.value) {
        const data: { explain?: string; name?: string } = { explain: form.explain };
        if (editId.value > 2) data.name = form.name;
        await updateRole(editId.value, data);
        message.success('更新成功');
      } else {
        await createRole({ name: form.name, explain: form.explain || undefined });
        message.success('创建成功，可在「菜单权限」里为其分配可见页面');
      }
      showForm.value = false;
      loadRoles();
    } catch (e: any) {
      message.error(e.message || '保存失败');
      return false;
    } finally {
      submitting.value = false;
    }
  }

  function openMenuAuth(item: RoleItem) {
    if (!menuTree.value.length) {
      message.error('菜单字典未就绪（若后端未重启请先重启并刷新页面）');
      return;
    }
    currentRole.value = item;
    // 回显只放叶子 id：父节点由级联推导，直接放父 id 会把未选的兄弟子级一起勾上
    checkedKeys.value = (item.menu_ids || [])
      .filter((id) => leafIds.value.has(id))
      .map((id) => String(id));
    expandedKeys.value = menuTree.value.map((n) => n.key);
    showMenuAuth.value = true;
  }

  async function submitMenuAuth() {
    if (!currentRole.value) return false;
    const ids = checkedKeys.value
      .map((k) => Number(k))
      .filter((id) => leafIds.value.has(id));
    if (submitting.value) return false;
    submitting.value = true;
    try {
      await updateRoleMenus(Number(currentRole.value.id), ids);
      message.success('菜单权限已更新');
      showMenuAuth.value = false;
      loadRoles();
    } catch (e: any) {
      message.error(e.message || '保存失败');
      return false;
    } finally {
      submitting.value = false;
    }
  }

  function handleDelete(item: RoleItem) {
    dialog.warning({
      title: '删除角色',
      content: `确定要删除「${item.name}」吗？已绑定该角色的用户会失去对应菜单权限。`,
      positiveText: '删除',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await deleteRole(Number(item.id));
          message.success('删除成功');
          loadRoles();
        } catch (e: any) {
          message.error(e.message || '删除失败');
        }
      },
    });
  }

  onMounted(() => {
    loadRoles();
    loadMenus();
  });
</script>

<style lang="less" scoped>
  .menu-tree-wrap {
    max-height: 420px;
    overflow: auto;
    padding: 4px 0;
  }
</style>
