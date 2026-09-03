<template>
  <div class="vault-page">
    <n-grid class="vault-grid" cols="1 s:1 m:1 l:4 xl:4 2xl:4" responsive="screen" :x-gap="12">
      <!-- 左侧目录树 -->
      <n-gi span="1" class="vault-col">
        <n-card title="目录" size="small" :bordered="false" :segmented="{ content: true }" class="dir-card">
          <template #header-extra>
            <n-space :size="4" align="center">
              <n-button size="tiny" quaternary :loading="refreshing" @click="handleRefresh">重扫</n-button>
              <n-dropdown trigger="hover" @select="handleAddNode" :options="addNodeOptions">
                <n-button type="info" ghost size="small" icon-placement="right">
                  新建
                  <template #icon>
                    <n-icon size="14"><DownOutlined /></n-icon>
                  </template>
                </n-button>
              </n-dropdown>
            </n-space>
          </template>

          <n-input
            v-model:value="searchKeyword"
            size="small"
            placeholder="搜索标题 / 标签 / 正文"
            clearable
            class="mb-2"
            @keyup.enter="handleSearch"
          >
            <template #suffix>
              <n-icon style="cursor: pointer" @click="handleSearch"><SearchOutlined /></n-icon>
            </template>
          </n-input>

          <n-spin :show="treeLoading" @contextmenu.prevent="onBlankContextMenu">
            <n-empty
              v-if="!treeLoading && treeData.length === 0"
              :description="isKnowledge ? '知识库为空' : '文档目录为空'"
              size="small"
            />
            <n-tree
              v-else
              block-line
              :data="treeData"
              :selected-keys="selectedKeys"
              :expanded-keys="expandedKeys"
              :node-props="treeNodeProps"
              @update:selected-keys="onSelectNode"
              @update:expanded-keys="onExpandNode"
            />
          </n-spin>
        </n-card>
      </n-gi>

      <!-- 右侧内容区 -->
      <n-gi span="3" class="vault-col">
        <n-card :bordered="false" :segmented="{ content: true }" class="edit-card">
          <template v-if="currentFile" #header>
            <n-space align="center">
              <n-icon size="18"><FileTextOutlined /></n-icon>
              <span>{{ currentFile.path }}</span>
              <n-tag
                v-if="currentFile.meta?.status === 'draft'"
                size="small"
                type="warning"
              >草稿</n-tag>
              <n-tag
                v-else-if="currentFile.meta?.status === 'published'"
                size="small"
                type="success"
              >已发布</n-tag>
              <n-tag v-for="t in currentFile.meta?.tags || []" :key="t" size="small">{{ t }}</n-tag>
            </n-space>
          </template>
          <template v-if="currentFile" #header-extra>
            <n-space>
              <n-button
                v-if="isKnowledge && currentFile.meta?.status === 'draft'"
                type="success"
                size="small"
                @click="handlePublish"
              >人审发布</n-button>
              <n-button size="small" @click="openMetaModal">元数据</n-button>
              <n-button size="small" @click="openMoveModal">移动/重命名</n-button>
              <n-button
                v-if="!currentFile.binary"
                type="primary"
                size="small"
                :loading="saveLoading"
                @click="handleSave"
              >保存</n-button>
              <n-button size="small" @click="handleDelete">删除</n-button>
            </n-space>
          </template>

          <template v-if="currentFile">
            <!-- 文本文件：编辑器 -->
            <MdEditor
              v-if="!currentFile.binary"
              v-model="editContent"
              :theme="isDark ? 'dark' : 'light'"
              class="vault-editor"
              placeholder="Markdown 内容（首部可带 frontmatter 元数据）"
              :toolbarsExclude="['github', 'save', 'htmlPreview', 'catalog']"
              :footers="[]"
            />
            <!-- 二进制文件：元数据卡片 -->
            <n-descriptions v-else bordered :column="2" size="small" label-placement="left">
              <n-descriptions-item label="文件">{{ currentFile.path }}</n-descriptions-item>
              <n-descriptions-item label="大小">{{ formatSize(currentFile.size) }}</n-descriptions-item>
              <n-descriptions-item label="空间">{{ currentFile.meta?.space || '-' }}</n-descriptions-item>
              <n-descriptions-item label="下载">
                <n-button size="tiny" tag="a" :href="rawHref" target="_blank">下载文件</n-button>
              </n-descriptions-item>
            </n-descriptions>
          </template>
          <n-empty v-else description="请从左侧选择一个文件" />
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
      <n-form
        ref="createFormRef"
        :model="createForm"
        :rules="createRules"
        label-placement="left"
        :label-width="80"
        class="py-4"
      >
        <n-form-item label="名称" path="name">
          <n-input v-model:value="createForm.name" :placeholder="createNamePlaceholder" />
        </n-form-item>
      </n-form>
    </n-modal>

    <!-- 元数据弹窗 -->
    <n-modal
      v-model:show="showMetaModal"
      preset="dialog"
      title="编辑元数据（frontmatter）"
      positive-text="确定"
      negative-text="取消"
      @positive-click="handleMetaSubmit"
      style="width: 480px"
    >
      <n-form label-placement="left" :label-width="80" class="py-4">
        <n-form-item label="标题">
          <n-input v-model:value="metaForm.title" placeholder="文档标题" />
        </n-form-item>
        <n-form-item label="标签">
          <n-select
            v-model:value="metaForm.tags"
            filterable
            multiple
            tag
            :show-arrow="false"
            placeholder="输入后回车添加标签"
          />
        </n-form-item>
        <n-form-item label="关联">
          <n-select
            v-model:value="metaForm.linked"
            filterable
            multiple
            tag
            :show-arrow="false"
            placeholder="task:123 / req:45 / tc:7"
          />
        </n-form-item>
      </n-form>
    </n-modal>

    <!-- 移动/重命名弹窗 -->
    <n-modal
      v-model:show="showMoveModal"
      preset="dialog"
      title="移动 / 重命名"
      positive-text="确定"
      negative-text="取消"
      @positive-click="handleMoveSubmit"
      style="width: 480px"
    >
      <n-form label-placement="top" class="py-4">
        <n-form-item label="目标路径（vault 内相对路径）">
          <n-input v-model:value="moveTarget" placeholder="例：设计/新名称.md" />
        </n-form-item>
      </n-form>
    </n-modal>

    <!-- 上传：隐藏 input -->
    <input ref="uploadInputRef" type="file" style="display: none" @change="handleUpload" />

    <!-- 树右键菜单：按空白区/目录/文件场景渲染 -->
    <n-dropdown
      trigger="manual"
      :show="ctxMenu.show"
      :x="ctxMenu.x"
      :y="ctxMenu.y"
      :options="ctxMenu.options"
      placement="bottom-start"
      @select="onCtxSelect"
      @clickoutside="ctxMenu.show = false"
    />
  </div>
</template>

<script lang="ts" setup>
  import { ref, reactive, computed, onMounted, nextTick } from 'vue';
  import { useRoute } from 'vue-router';
  import { MdEditor } from 'md-editor-v3';
  import 'md-editor-v3/lib/style.css';
  import { useMessage, useDialog } from 'naive-ui';
  import { DownOutlined, SearchOutlined, FileTextOutlined } from '@vicons/antd';
  import { useDesignSetting } from '@/hooks/setting/useDesignSetting';
  import {
    getVaultTree,
    getVaultFile,
    writeVaultFile,
    createVaultFolder,
    uploadVaultFile,
    moveVaultPath,
    deleteVaultPath,
    updateVaultMeta,
    searchVault,
    refreshVault,
    vaultRawUrl,
  } from '@/api/vault/index';
  import type { VaultTreeNode, VaultFile } from '@/api/vault/index';

  const message = useMessage();
  const dialog = useDialog();
  const route = useRoute();
  const { getDarkTheme } = useDesignSetting();
  const isDark = computed(() => getDarkTheme.value === true);

  const projectId = computed(() => Number(route.params.projectId));
  // 路由 meta.vaultSpace=knowledge：知识库模式（仅 知识库/ 目录 + 发布流）；否则文档模式（全空间）
  const isKnowledge = computed(() => route.meta.vaultSpace === 'knowledge');

  const treeLoading = ref(false);
  const treeData = ref<any[]>([]);
  const expandedKeys = ref<string[]>([]);
  const selectedKeys = ref<string[]>([]);
  const currentFile = ref<VaultFile | null>(null);
  const editContent = ref('');
  const saveLoading = ref(false);
  const refreshing = ref(false);
  const searchKeyword = ref('');

  const showCreateModal = ref(false);
  const createType = ref<'folder' | 'doc'>('doc');
  const createFormRef = ref<any>(null);
  const createForm = reactive({ name: '' });
  const createRules = { name: { required: true, message: '请输入名称', trigger: 'blur' } };
  const createNamePlaceholder = computed(() =>
    createType.value === 'folder' ? '文件夹名，例：设计' : '文件名，例：login.md'
  );

  const showMetaModal = ref(false);
  const metaForm = reactive({ title: '', tags: [] as string[], linked: [] as string[] });

  const showMoveModal = ref(false);
  const moveTarget = ref('');

  const uploadInputRef = ref<HTMLInputElement | null>(null);

  const addNodeOptions = computed(() => [
    { label: '新建文件夹', key: 'folder' },
    { label: '新建文档', key: 'doc' },
    { label: '上传文件', key: 'upload' },
  ]);

  const rawHref = computed(() =>
    currentFile.value ? vaultRawUrl(projectId.value, currentFile.value.path) : '#'
  );

  // 当前选中节点所在目录（新建/上传的落点）
  const selectedDir = computed(() => {
    if (!currentFile.value) return isKnowledge.value ? '知识库' : '';
    const p = currentFile.value.path;
    const i = p.lastIndexOf('/');
    const dir = i >= 0 ? p.slice(0, i) : '';
    return dir || (isKnowledge.value ? '知识库' : '');
  });

  function formatSize(n: number) {
    if (n < 1024) return `${n} B`;
    if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
    return `${(n / 1024 / 1024).toFixed(1)} MB`;
  }

  // 目录节点 path 集合：点击目录是展开/收起而非读文件
  const dirPaths = new Set<string>();

  function transformTree(nodes: VaultTreeNode[]): any[] {
    return nodes.map((node) => {
      if (node.isDir) dirPaths.add(node.path);
      return {
        key: node.path,
        label: node.meta?.title && node.meta.title !== node.name
          ? `${node.name}（${node.meta.title}）`
          : node.name,
        isLeaf: !node.isDir,
        prefix: () => (node.isDir ? '📁' : '📄'),
        // 空目录给空数组：undefined 会被 n-tree 当异步节点显示 loading
        children: node.isDir
          ? node.children && node.children.length
            ? transformTree(node.children)
            : []
          : undefined,
      };
    });
  }

  async function loadTree(keepSelection = true) {
    treeLoading.value = true;
    try {
      // 两视图互斥：知识库页只看知识库空间，文档页只看工作区（过程文档）
      const res = await getVaultTree(projectId.value, isKnowledge.value ? 'knowledge' : 'work');
      dirPaths.clear();
      let tree = res?.tree || [];
      // 知识库模式：页面上下文已表达"知识库"，剥掉同名顶层父节点直接展示其内容
      // （仅展示层剥皮，文件完整路径不变）
      if (isKnowledge.value && tree.length === 1 && tree[0].isDir && tree[0].path === '知识库') {
        tree = tree[0].children || [];
      }
      treeData.value = transformTree(tree);
      if (!keepSelection) {
        selectedKeys.value = [];
        currentFile.value = null;
        editContent.value = '';
      }
    } catch {
      // 错误提示由 http 层统一处理
    } finally {
      treeLoading.value = false;
    }
  }

  async function onSelectNode(keys: string[]) {
    if (keys.length === 0) return;
    const key = String(keys[0]);
    // 目录节点：点击即切换展开，不进入选中态也不读文件
    if (dirPaths.has(key)) {
      selectedKeys.value = currentFile.value ? [currentFile.value.path] : [];
      expandedKeys.value = expandedKeys.value.includes(key)
        ? expandedKeys.value.filter((k) => k !== key)
        : [...expandedKeys.value, key];
      return;
    }
    selectedKeys.value = keys;
    await openFile(key);
  }

  async function openFile(path: string) {
    try {
      const res = await getVaultFile(projectId.value, path);
      currentFile.value = res || null;
      editContent.value = res?.content || '';
    } catch {
      message.error('加载文件失败');
    }
  }

  function onExpandNode(keys: string[]) {
    expandedKeys.value = keys;
  }

  function openMetaModal() {
    metaForm.title = currentFile.value?.meta?.title || '';
    metaForm.tags = [...(currentFile.value?.meta?.tags || [])];
    metaForm.linked = [...(currentFile.value?.meta?.linked || [])];
    showMetaModal.value = true;
  }

  function openMoveModal(targetPath?: string) {
    moveTarget.value = targetPath ?? currentFile.value?.path ?? '';
    showMoveModal.value = true;
  }

  // 右键菜单指定的落点（新建/上传目标目录）；null = 跟随当前选中文件所在目录
  let ctxTargetDir: string | null = null;

  function handleAddNode(key: string, dir?: string) {
    ctxTargetDir = dir ?? null;
    if (key === 'upload') {
      uploadInputRef.value?.click();
      return;
    }
    createType.value = key as 'folder' | 'doc';
    createForm.name = '';
    showCreateModal.value = true;
  }

  async function handleCreateSubmit() {
    try {
      await createFormRef.value?.validate();
    } catch {
      return false;
    }
    const name = createForm.name.trim();
    const baseDir = ctxTargetDir !== null ? ctxTargetDir : selectedDir.value;
    const path = baseDir ? `${baseDir}/${name}` : name;
    try {
      if (createType.value === 'folder') {
        await createVaultFolder(projectId.value, path);
      } else {
        const title = name.replace(/\.md$/i, '');
        const frontmatter = isKnowledge.value
          ? `---\ntitle: ${title}\nspace: knowledge\nstatus: draft\n---\n\n`
          : '';
        await writeVaultFile(projectId.value, path.endsWith('.md') ? path : `${path}.md`, frontmatter);
      }
      message.success('创建成功');
      showCreateModal.value = false;
      loadTree();
    } catch {
      message.error('创建失败');
      return false;
    }
  }

  async function handleUpload(e: Event) {
    const input = e.target as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;
    const baseDir = ctxTargetDir !== null ? ctxTargetDir : selectedDir.value;
    try {
      const res = await uploadVaultFile(projectId.value, baseDir, file);
      message.success(`上传成功：${res?.path || file.name}`);
      loadTree();
    } catch {
      message.error('上传失败');
    } finally {
      input.value = '';
      ctxTargetDir = null;
    }
  }

  async function handleSave() {
    if (!currentFile.value) return;
    saveLoading.value = true;
    try {
      await writeVaultFile(projectId.value, currentFile.value.path, editContent.value);
      message.success('保存成功（旧版已快照至 .history）');
      await openFile(currentFile.value.path);
      loadTree();
    } catch {
      message.error('保存失败');
    } finally {
      saveLoading.value = false;
    }
  }

  async function handlePublish() {
    if (!currentFile.value) return;
    try {
      await updateVaultMeta(projectId.value, currentFile.value.path, { status: 'published' });
      message.success('已发布');
      await openFile(currentFile.value.path);
    } catch {
      message.error('发布失败');
    }
  }

  function handleDelete() {
    if (!currentFile.value) return;
    handleDeleteFor(currentFile.value.path, false);
  }

  // ==================== 右键菜单（空白区 / 目录 / 文件 三场景） ====================

  interface CtxTarget {
    path: string;
    isDir: boolean;
    hasChildren: boolean;
    name: string;
  }

  const ctxMenu = reactive({
    show: false,
    x: 0,
    y: 0,
    options: [] as Array<{ label: string; key: string; disabled?: boolean; divider?: boolean }>,
  });
  let ctxTarget: CtxTarget | null = null;

  function treeNodeProps({ option }: { option: any }) {
    return {
      onContextmenu: (e: MouseEvent) => onNodeContextMenu(e, option),
    };
  }

  function onNodeContextMenu(e: MouseEvent, option: any) {
    e.preventDefault();
    e.stopPropagation();
    openCtxMenu(e, {
      path: String(option.key),
      isDir: !option.isLeaf,
      hasChildren: Array.isArray(option.children) && option.children.length > 0,
      name: String(option.key).split('/').pop() || '',
    });
    // 右键文件时同步视觉选中（不打开）
    if (option.isLeaf) selectedKeys.value = [String(option.key)];
  }

  function onBlankContextMenu(e: MouseEvent) {
    openCtxMenu(e, null);
  }

  function openCtxMenu(e: MouseEvent, target: CtxTarget | null) {
    ctxTarget = target;
    if (!target) {
      // 空白区：新建/上传落在当前视图根（知识库页为 知识库/，文档页为根）
      ctxMenu.options = [
        { label: '新建文档', key: 'new-doc' },
        { label: '新建文件夹', key: 'new-folder' },
        { label: '上传文件', key: 'upload' },
      ];
    } else if (target.isDir) {
      ctxMenu.options = [
        { label: `在「${target.name}」中新建文档`, key: 'new-doc' },
        { label: `在「${target.name}」中新建文件夹`, key: 'new-folder' },
        { label: `上传到「${target.name}」`, key: 'upload' },
        { label: '重命名 / 移动', key: 'move', divider: true },
        // 非空目录禁止删除（后端同样校验），防止误删整棵子树
        { label: '删除（仅空目录）', key: 'delete', disabled: target.hasChildren },
      ];
    } else {
      ctxMenu.options = [
        { label: '打开', key: 'open' },
        { label: '重命名 / 移动', key: 'move', divider: true },
        { label: '删除', key: 'delete' },
      ];
    }
    ctxMenu.show = false;
    nextTick(() => {
      ctxMenu.x = e.clientX;
      ctxMenu.y = e.clientY;
      ctxMenu.show = true;
    });
  }

  function onCtxSelect(key: string) {
    ctxMenu.show = false;
    const t = ctxTarget;
    // 落点：目录右键 → 该目录；文件右键 → 其父目录；空白 → 不指定（落当前视图默认根）
    const dir = t ? (t.isDir ? t.path : t.path.slice(0, t.path.lastIndexOf('/')) || '') : undefined;
    switch (key) {
      case 'new-doc':
        handleAddNode('doc', dir);
        break;
      case 'new-folder':
        handleAddNode('folder', dir);
        break;
      case 'upload':
        handleAddNode('upload', dir);
        break;
      case 'open':
        if (t) openFile(t.path);
        break;
      case 'move':
        if (t) openMoveModal(t.path);
        break;
      case 'delete':
        if (t) handleDeleteFor(t.path, t.isDir);
        break;
    }
  }

  function handleDeleteFor(path: string, isDir: boolean) {
    dialog.warning({
      title: '确认删除',
      content: isDir
        ? `确定要删除空目录「${path}」吗？`
        : `确定要删除「${path}」吗？删除前会快照到 .history。`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await deleteVaultPath(projectId.value, path);
          message.success('删除成功');
          if (currentFile.value && currentFile.value.path === path) {
            currentFile.value = null;
            editContent.value = '';
          }
          await loadTree(false);
        } catch {
          message.error('删除失败');
        }
      },
    });
  }

  async function handleMetaSubmit() {
    if (!currentFile.value) return false;
    try {
      await updateVaultMeta(projectId.value, currentFile.value.path, {
        title: metaForm.title,
        tags: metaForm.tags,
        linked: metaForm.linked,
      });
      message.success('元数据已更新');
      showMetaModal.value = false;
      await openFile(currentFile.value.path);
      loadTree();
    } catch {
      message.error('更新失败');
      return false;
    }
  }

  async function handleMoveSubmit() {
    if (!currentFile.value || !moveTarget.value.trim()) return false;
    try {
      await moveVaultPath(projectId.value, currentFile.value.path, moveTarget.value.trim());
      message.success('移动成功');
      showMoveModal.value = false;
      await loadTree(false);
    } catch {
      message.error('移动失败');
      return false;
    }
  }

  async function handleSearch() {
    const q = searchKeyword.value.trim();
    if (!q) {
      loadTree();
      return;
    }
    try {
      const res = await searchVault(projectId.value, q, isKnowledge.value ? 'knowledge' : 'work');
      const items = res?.items || [];
      if (items.length === 0) {
        message.info('未搜索到匹配文档');
        return;
      }
      // 搜索结果平铺为树（点击直接打开文件）
      treeData.value = items.map((it) => ({
        key: it.path,
        label: `${it.title} — ${it.path}`,
        isLeaf: true,
        prefix: () => '📄',
      }));
      expandedKeys.value = [];
    } catch {
      message.error('搜索失败');
    }
  }

  async function handleRefresh() {
    refreshing.value = true;
    try {
      const res = await refreshVault(projectId.value);
      message.success(`重扫完成：${res?.changed ?? 0} 变更，${res?.deleted ?? 0} 删除`);
      await loadTree();
    } catch {
      message.error('重扫失败');
    } finally {
      refreshing.value = false;
    }
  }

  onMounted(() => {
    loadTree();
  });
</script>

<style lang="less" scoped>
  // 整页铺满内容区视口：双栏等高、树与编辑器在卡内滚动
  .vault-page {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  .vault-grid {
    flex: 1;
    min-height: 0;

    // grid item 默认高度自适应内容，需显式拉满供卡内 flex 链使用
    :deep(.n-grid-item) {
      height: 100%;
    }
  }

  .vault-col {
    min-height: 0;

    .dir-card,
    .edit-card {
      height: 100%;
      display: flex;
      flex-direction: column;

      :deep(.n-card__content) {
        flex: 1;
        min-height: 0;
        display: flex;
        flex-direction: column;
      }
    }

    // 目录树区滚动
    .dir-card :deep(.n-spin-container) {
      flex: 1;
      min-height: 0;
      overflow-y: auto;
    }

    // 编辑器铺满剩余高度
    .vault-editor {
      flex: 1;
      min-height: 0;
    }
  }
</style>
