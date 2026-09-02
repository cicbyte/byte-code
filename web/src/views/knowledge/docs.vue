<template>
  <div>
    <div class="n-layout-page-header">
      <n-card :bordered="false" title="知识库" />
    </div>

    <n-grid class="mt-4" cols="1 s:1 m:1 l:4 xl:4 2xl:4" responsive="screen" :x-gap="12">
      <!-- 左侧文档树 -->
      <n-gi span="1">
        <n-card title="文档目录" size="small" :bordered="false" :segmented="{ content: true }">
          <template #header-extra>
            <n-dropdown trigger="hover" @select="handleAddNode" :options="addNodeOptions">
              <n-button type="info" ghost size="small" icon-placement="right">
                新建
                <template #icon>
                  <n-icon size="14"><DownOutlined /></n-icon>
                </template>
              </n-button>
            </n-dropdown>
          </template>

          <n-spin :show="treeLoading">
            <n-empty v-if="!treeLoading && treeData.length === 0" description="暂无文档" size="small" />
            <n-tree
              v-else
              block-line
              :data="treeData"
              :selected-keys="selectedKeys"
              :expanded-keys="expandedKeys"
              @update:selected-keys="onSelectNode"
              @update:expanded-keys="onExpandNode"
            />
          </n-spin>
        </n-card>
      </n-gi>

      <!-- 右侧编辑区 -->
      <n-gi span="3">
        <n-card :bordered="false" :segmented="{ content: true }">
          <template v-if="currentDoc" #header>
            <n-space align="center">
              <n-icon size="18"><FormOutlined /></n-icon>
              <span>{{ currentDoc.title }}</span>
            </n-space>
          </template>
          <template v-if="currentDoc" #header-extra>
            <n-space>
              <n-button type="primary" size="small" :loading="saveLoading" @click="handleSave">保存</n-button>
              <n-button size="small" @click="handleDeleteDoc">删除</n-button>
            </n-space>
          </template>

          <template v-if="currentDoc">
            <MdEditor
              v-model="editContent"
              :style="{ height: 'calc(100vh - 280px)' }"
              placeholder="请输入文档内容（支持 Markdown）"
              :toolbarsExclude="['github', 'save', 'htmlPreview', 'catalog']"
              :footers="[]"
            />
          </template>
          <n-empty v-else description="请从左侧选择一个文档" />
        </n-card>
      </n-gi>
    </n-grid>

    <!-- 新建弹窗 -->
    <n-modal
      v-model:show="showCreateModal"
      preset="dialog"
      :title="createType === 'folder' ? '新建文件夹' : '新建文档'"
      positive-text="确定"
      negative-text="取消"
      @positive-click="handleCreateSubmit"
      style="width: 420px"
    >
      <n-form ref="createFormRef" :model="createForm" :rules="createRules" label-placement="left" :label-width="80" class="py-4">
        <n-form-item label="名称" path="title">
          <n-input v-model:value="createForm.title" placeholder="请输入名称" />
        </n-form-item>
      </n-form>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import { ref, reactive, onMounted, computed } from 'vue';
  import { useRoute } from 'vue-router';
  import { MdEditor } from 'md-editor-v3';
  import 'md-editor-v3/lib/style.css';
  import { useMessage, useDialog } from 'naive-ui';
  import { DownOutlined, FormOutlined } from '@vicons/antd';
  import { getDocTree, getDoc, createDoc, updateDoc, deleteDoc } from '@/api/knowledge/index';
  import type { DocItem, DocTreeNode } from '@/api/knowledge/index';

  const message = useMessage();
  const dialog = useDialog();

  // 硬编码项目ID（后续可从路由获取）
  const route = useRoute();
  const currentProjectId = computed(() => Number(route.params.projectId));

  const treeLoading = ref(false);
  const treeData = ref<any[]>([]);
  const expandedKeys = ref<string[]>([]);
  const selectedKeys = ref<string[]>([]);
  const currentDoc = ref<DocItem | null>(null);
  const editContent = ref('');
  const saveLoading = ref(false);
  const showCreateModal = ref(false);
  const createType = ref<string>('doc');
  const createFormRef = ref<any>(null);
  const createForm = reactive({ title: '' });
  const createRules = {
    title: { required: true, message: '请输入名称', trigger: 'blur' },
  };

  const addNodeOptions = [
    { label: '新建文件夹', key: 'folder' },
    { label: '新建文档', key: 'doc' },
  ];

  function transformTree(nodes: DocTreeNode[]): any[] {
    return nodes.map((node) => ({
      key: String(node.id),
      label: node.title,
      prefix: () => (node.type === 'folder' ? '📁' : '📄'),
      children: node.children ? transformTree(node.children) : undefined,
    }));
  }

  async function loadTree() {
    treeLoading.value = true;
    try {
      const res = await getDocTree(currentProjectId.value);
      treeData.value = transformTree(res?.tree || []);
    } catch (e) {
      // ignore
    } finally {
      treeLoading.value = false;
    }
  }

  async function onSelectNode(keys: string[]) {
    selectedKeys.value = keys;
    if (keys.length === 0) {
      currentDoc.value = null;
      editContent.value = '';
      return;
    }
    const docId = Number(keys[0]);
    try {
      const res = await getDoc(docId);
      currentDoc.value = res || null;
      editContent.value = res?.content || '';
    } catch (e) {
      message.error('加载文档失败');
    }
  }

  function onExpandNode(keys: string[]) {
    expandedKeys.value = keys;
  }

  function handleAddNode(key: string) {
    createType.value = key;
    createForm.title = '';
    showCreateModal.value = true;
  }

  async function handleCreateSubmit() {
    try {
      await createFormRef.value?.validate();
    } catch {
      return false;
    }
    try {
      const parentId = selectedKeys.value.length ? Number(selectedKeys.value[0]) : undefined;
      await createDoc({
        projectId: currentProjectId.value,
        title: createForm.title,
        type: createType.value,
        parentId,
      });
      message.success('创建成功');
      showCreateModal.value = false;
      loadTree();
    } catch (e) {
      message.error('创建失败');
      return false;
    }
  }

  async function handleSave() {
    if (!currentDoc.value) return;
    saveLoading.value = true;
    try {
      await updateDoc(currentDoc.value.id, { content: editContent.value });
      message.success('保存成功');
    } catch (e) {
      message.error('保存失败');
    } finally {
      saveLoading.value = false;
    }
  }

  function handleDeleteDoc() {
    if (!currentDoc.value) return;
    dialog.warning({
      title: '确认删除',
      content: `确定要删除「${currentDoc.value.title}」吗？`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await deleteDoc(currentDoc.value!.id);
          message.success('删除成功');
          currentDoc.value = null;
          editContent.value = '';
          selectedKeys.value = [];
          loadTree();
        } catch (e) {
          message.error('删除失败');
        }
      },
    });
  }

  onMounted(() => {
    loadTree();
  });
</script>
