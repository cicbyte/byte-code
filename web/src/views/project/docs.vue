<template>
  <div class="docs-page">
    <n-grid class="docs-grid" cols="1 s:1 m:1 l:4 xl:4 2xl:4" responsive="screen" :x-gap="12">
      <!-- 左侧目录树 -->
      <n-gi span="1" class="docs-col">
        <n-card title="目录" size="small" :bordered="false" :segmented="{ content: true }" class="dir-card">
          <n-progress
            v-if="uploading.active"
            type="line"
            :percentage="uploading.total ? Math.round((uploading.done / uploading.total) * 100) : 0"
            :height="6"
            :border-radius="3"
            class="mb-2"
          >
            {{ uploading.done }}/{{ uploading.total }}{{ uploading.failed ? `（${uploading.failed} 失败）` : '' }}
          </n-progress>

          <div v-if="searchMode" class="search-banner">
            <span>搜索结果</span>
            <n-button size="tiny" quaternary type="primary" @click="searchKeyword = ''; exitSearchMode()">
              返回目录
            </n-button>
          </div>

          <n-input
            v-model:value="searchKeyword"
            size="small"
            placeholder="搜索标题 / 标签 / 正文"
            clearable
            class="mb-2"
            @keyup.enter="handleSearch"
            @clear="exitSearchMode"
          >
            <template #suffix>
              <n-icon style="cursor: pointer" @click="handleSearch"><SearchOutlined /></n-icon>
            </template>
          </n-input>

          <n-spin :show="treeLoading" class="dir-spin" @contextmenu.prevent="onBlankContextMenu">
            <EmptyState
              v-if="!treeLoading && treeData.length === 0"
              type="doc"
              :title="isKnowledge ? '知识库为空' : '文档目录为空'"
              description="上传或新建文档后展示在这里"
              compact
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
      <n-gi span="3" class="docs-col">
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
              <n-button size="small" @click="showHistory = true">历史</n-button>
              <n-button size="small" @click="openMetaModal">元数据</n-button>
              <n-button size="small" @click="openMoveModal()">移动/重命名</n-button>
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
            <!-- Markdown：编辑器 -->
            <MdEditor
              v-if="!currentFile.binary"
              v-model="editContent"
              :theme="isDark ? 'dark' : 'light'"
              class="docs-editor"
              placeholder="Markdown 内容（首部可带 frontmatter 元数据）"
              :toolbarsExclude="['github', 'save', 'htmlPreview', 'catalog']"
              :footers="[]"
              :sanitize="safeHtml"
            />
            <template v-else>
              <!-- docx：mammoth 转 HTML 只读预览 -->
              <div v-if="isDocx" class="docx-wrap">
                <n-spin v-if="docxLoading" class="docx-spin" />
                <div v-else-if="docxHtml" class="docx-html" v-html="docxHtml"></div>
                <n-result v-else status="warning" title="预览失败" description="Word 文档解析失败，请下载后查看">
                </n-result>
              </div>
              <!-- 图片：直出 -->
              <div v-else-if="isImage" class="img-wrap">
                <img :src="binObjectUrl" :alt="currentFile.path" />
              </div>
              <!-- 其他二进制（pdf/xlsx/zip…）：说明 + 下载 -->
              <n-result v-else status="info" title="该格式不支持在线预览" description="请下载后查看">
              </n-result>
              <n-space class="mt-3" align="center">
                <n-button size="small" :disabled="!binObjectUrl" @click="downloadBin">下载文件</n-button>
                <span class="bin-meta">{{ currentFile.path }} · {{ formatSize(currentFile.size) }}</span>
              </n-space>
            </template>
          </template>
          <EmptyState v-else type="doc" title="未选择文档" description="从左侧目录选择一个文件开始阅读" compact />
        </n-card>
      </n-gi>
    </n-grid>

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

    <!-- 移动/重命名弹窗：from 显式展示（右键目标可能与当前打开文件不同） -->
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
        <n-form-item label="从">
          <n-text depth="3" code>{{ moveFrom || '（未选择）' }}</n-text>
        </n-form-item>
        <n-form-item label="目标路径（vault 内相对路径）">
          <n-input v-model:value="moveTarget" placeholder="例：设计/新名称.md" />
        </n-form-item>
      </n-form>
    </n-modal>

    <!-- 版本历史弹窗 -->
    <DocsHistoryModal
      v-model:show="showHistory"
      :project-id="projectId"
      :path="currentFile?.path || ''"
      @restored="onHistoryRestored"
    />

    <!-- 上传：隐藏 input -->
    <input ref="uploadInputRef" type="file" style="display: none" multiple @change="handleUpload" />

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
  // 文档中枢页面：模板与编排层。实现按职责拆至 composables/：
  // useInlineEdit（树内新建/重命名）/ useCtxMenu（右键菜单）/ useBinaryPreview（docx/图片预览）
  // / useDirtyGuard（未保存守卫）
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { ref, reactive, computed, onMounted, onUnmounted, watch } from 'vue';
  import { useRoute, onBeforeRouteLeave } from 'vue-router';
  import { MdEditor } from 'md-editor-v3';
  import DOMPurify from 'dompurify';
  import 'md-editor-v3/lib/style.css';
  import { useMessage, useDialog } from 'naive-ui';
  import { SearchOutlined, FileTextOutlined } from '@vicons/antd';
  import { useDesignSetting } from '@/hooks/setting/useDesignSetting';
  import {
    getDocsTree,
    getDocsFile,
    writeDocsFile,
    uploadDocsFile,
    moveDocsPath,
    deleteDocsPath,
    updateDocsMeta,
    searchDocsApi,
    refreshDocs,
  } from '@/api/docs/index';
  import type { DocsTreeNode, DocsFile } from '@/api/docs/index';
  import { useInlineEdit } from './composables/useInlineEdit';
  import { useCtxMenu } from './composables/useCtxMenu';
  import { useBinaryPreview } from './composables/useBinaryPreview';
  import { useDirtyGuard } from './composables/useDirtyGuard';
  import DocsHistoryModal from './components/DocsHistoryModal.vue';

  const message = useMessage();
  const dialog = useDialog();
  const route = useRoute();
  const { getDarkTheme } = useDesignSetting();
  const isDark = computed(() => getDarkTheme.value === true);

  const projectId = computed(() => Number(route.params.projectId));
  // 路由 meta.vaultSpace=knowledge：知识库模式（仅 知识库/ 目录 + 发布流）；否则文档模式
  const isKnowledge = computed(() => route.meta.vaultSpace === 'knowledge');

  // ==================== 树与文件状态 ====================
  const treeLoading = ref(false);
  const treeData = ref<any[]>([]);
  const expandedKeys = ref<string[]>([]);
  const selectedKeys = ref<string[]>([]);
  const currentFile = ref<DocsFile | null>(null);
  const editContent = ref('');
  const saveLoading = ref(false);
  const refreshing = ref(false);
  const searchKeyword = ref('');

  const showHistory = ref(false);
  const showMetaModal = ref(false);
  const metaForm = reactive({ title: '', tags: [] as string[], linked: [] as string[] });

  const showMoveModal = ref(false);
  const moveTarget = ref('');
  // 移动操作的源路径：右键目标可能与当前打开文件不同，提交必须用打开时记录的源，
  // 不能取 currentFile（否则右键 B 时会把当前打开的 A 移走）
  const moveFrom = ref('');

  const uploadInputRef = ref<HTMLInputElement | null>(null);

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

  // ==================== 空间守卫 ====================
  // 文档页（work）与知识库页（knowledge）共用同一 vault 与同一套 API，
  // 视图互斥只做了读取过滤；写入操作必须约束路径前缀，防止跨空间移动/上传绕过发布流
  const KB_PREFIX = '知识库/';
  function isKnowledgePath(p: string): boolean {
    return p === '知识库' || p.startsWith(KB_PREFIX);
  }
  function assertSpaceAllowed(p: string): boolean {
    if (!p) return true;
    const inKb = isKnowledgePath(p);
    if (isKnowledge.value && !inKb) {
      message.error('知识库内不支持此操作目标（路径超出知识库空间）');
      return false;
    }
    if (!isKnowledge.value && inKb) {
      message.error('文档视图不能操作知识库路径，请到知识库页操作');
      return false;
    }
    return true;
  }

  // ==================== 目录节点集合与树加载 ====================
  // 目录节点 path 集合：点击目录是展开/收起而非读文件
  const dirPaths = new Set<string>();

  const { editing, nodeLabel, startInlineCreate, startInlineRename, cancelInline } = useInlineEdit({
    treeData, expandedKeys, projectId, isKnowledge, assertSpaceAllowed, loadTree,
  });

  function transformTree(nodes: DocsTreeNode[]): any[] {
    return nodes.map((node) => {
      if (node.isDir) dirPaths.add(node.path);
      return {
        key: node.path,
        label: nodeLabel(node),
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
      const res = await getDocsTree(projectId.value, isKnowledge.value ? 'knowledge' : 'work');
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

  // ==================== 未保存守卫 + 文件打开 ====================
  const { dirty, markOpened, confirmDiscard } = useDirtyGuard({ currentFile, editContent });

  const {
    docxHtml, docxLoading, binObjectUrl, isDocx, isImage,
    loadBinObjectUrl, loadDocxPreview, downloadBin, onUnmountCleanup: cleanupPreview,
  } = useBinaryPreview({ projectId, currentFile });
  onUnmounted(cleanupPreview);

  async function openFile(path: string) {
    try {
      const res = await getDocsFile(projectId.value, path);
      currentFile.value = res || null;
      editContent.value = res?.content || '';
      markOpened(res?.content || '');
      if (res?.binary) {
        await loadBinObjectUrl();
        if (path.toLowerCase().endsWith('.docx')) loadDocxPreview();
      }
    } catch {
      message.error('加载文件失败');
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
    if (dirty.value && key !== currentFile.value?.path) {
      if (!(await confirmDiscard('切换文件'))) {
        selectedKeys.value = currentFile.value ? [currentFile.value.path] : [];
        return;
      }
    }
    selectedKeys.value = keys;
    await openFile(key);
  }

  // 历史恢复后：重读当前文件内容（快照保底使旧内容仍在历史中可回）
  async function onHistoryRestored() {
    if (currentFile.value) {
      await openFile(currentFile.value.path);
      loadTree();
    }
  }

  function onExpandNode(keys: string[]) {
    expandedKeys.value = keys;
  }

  // ==================== 弹窗与文件操作 ====================
  function openMetaModal() {
    metaForm.title = currentFile.value?.meta?.title || '';
    metaForm.tags = [...(currentFile.value?.meta?.tags || [])];
    metaForm.linked = [...(currentFile.value?.meta?.linked || [])];
    showMetaModal.value = true;
  }

  function openMoveModal(targetPath?: string) {
    // 防御：模板 @click 无括号调用会把 Event 对象传进来（truthy 会绕过 ?? 兜底）
    const from = typeof targetPath === 'string' ? targetPath : currentFile.value?.path ?? '';
    moveFrom.value = from;
    moveTarget.value = from;
    showMoveModal.value = true;
  }

  // 右键菜单指定的落点（新建/上传目标目录）；null = 跟随当前选中文件所在目录
  let ctxTargetDir: string | null = null;

  // 上传队列状态：多文件逐个上传，列表展示进度与结果
  const uploading = reactive({ active: false, done: 0, total: 0, failed: 0 });

  function triggerUpload(dir: string) {
    ctxTargetDir = dir;
    uploadInputRef.value?.click();
  }

  async function handleUpload(e: Event) {
    const input = e.target as HTMLInputElement;
    const files = Array.from(input.files || []);
    if (files.length === 0) return;
    const baseDir = ctxTargetDir !== null ? ctxTargetDir : selectedDir.value;
    if (!assertSpaceAllowed(baseDir)) {
      input.value = '';
      ctxTargetDir = null;
      return;
    }
    uploading.active = true;
    uploading.done = 0;
    uploading.failed = 0;
    uploading.total = files.length;
    for (const file of files) {
      try {
        await uploadDocsFile(projectId.value, baseDir, file);
      } catch {
        uploading.failed++;
        message.error(`上传失败：${file.name}`);
      } finally {
        uploading.done++;
      }
    }
    if (uploading.failed < uploading.total) {
      message.success(`上传完成：${uploading.total - uploading.failed}/${uploading.total} 个文件`);
    }
    uploading.active = false;
    input.value = '';
    ctxTargetDir = null;
    loadTree();
  }

  async function handleSave() {
    if (!currentFile.value) return;
    saveLoading.value = true;
    try {
      await writeDocsFile(projectId.value, currentFile.value.path, editContent.value);
      message.success('保存成功（旧版已快照至 .history）');
      await openFile(currentFile.value.path);
      markOpened(editContent.value);
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
      await updateDocsMeta(projectId.value, currentFile.value.path, { status: 'published' });
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
          await deleteDocsPath(projectId.value, path);
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
      await updateDocsMeta(projectId.value, currentFile.value.path, {
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
    const from = moveFrom.value.trim();
    const to = moveTarget.value.trim();
    if (!from || !to) return false;
    if (from === to) return false;
    if (!assertSpaceAllowed(to)) return false;
    try {
      await moveDocsPath(projectId.value, from, to);
      message.success('移动成功');
      showMoveModal.value = false;
      // 移动的恰是当前打开文件时同步更新编辑器状态，否则保持不动
      if (currentFile.value && currentFile.value.path === from) {
        currentFile.value = null;
        editContent.value = '';
      }
      await loadTree(false);
    } catch {
      message.error('移动失败');
      return false;
    }
  }

  // ==================== 右键菜单 ====================
  const { ctxMenu, treeNodeProps, onBlankContextMenu, onCtxSelect } = useCtxMenu({
    isKnowledge,
    selectedKeys,
    actions: {
      startInlineCreate,
      startInlineRename,
      openFile,
      openMoveModal,
      handleDeleteFor,
      handleRefresh,
      triggerUpload,
    },
  });

  // ==================== 搜索结果态与目录树态 ====================
  // 两态显式建模：进入/退出统一收口（清 dirPaths、取消内联编辑、恢复展开态）
  const searchMode = ref(false);
  const savedExpanded = ref<string[]>([]);

  function exitSearchMode() {
    if (!searchMode.value) return;
    searchMode.value = false;
    expandedKeys.value = savedExpanded.value;
    cancelInline();
    loadTree();
  }

  async function handleSearch() {
    const q = searchKeyword.value.trim();
    if (!q) {
      exitSearchMode();
      return;
    }
    try {
      const res = await searchDocsApi(projectId.value, q, isKnowledge.value ? 'knowledge' : 'work');
      const items = res?.items || [];
      if (items.length === 0) {
        message.info('未搜索到匹配文档');
        return;
      }
      // 进入搜索态：记住展开态；结果平铺（点击直接打开文件）
      if (!searchMode.value) {
        searchMode.value = true;
        savedExpanded.value = [...expandedKeys.value];
        cancelInline();
        dirPaths.clear(); // 搜索结果无目录语义，旧集合跨态残留会导致误判
      }
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
      const res = await refreshDocs(projectId.value);
      message.success(`重扫完成：${res?.changed ?? 0} 变更，${res?.deleted ?? 0} 删除`);
      await loadTree();
    } catch {
      message.error('重扫失败');
    } finally {
      refreshing.value = false;
    }
  }

  // ==================== 生命周期与视图切换 ====================
  // 关联文档跳转支持：query.path 直达打开文件并展开父目录
  async function tryOpenQueryPath() {
    const qpath = route.query.path;
    if (typeof qpath !== 'string' || !qpath) return;
    await openFile(qpath);
    selectedKeys.value = [qpath];
    const parts = qpath.split('/');
    const parents: string[] = [];
    for (let i = 1; i < parts.length; i++) parents.push(parts.slice(0, i).join('/'));
    expandedKeys.value = [...new Set([...expandedKeys.value, ...parents])];
  }

  // md-editor-v3 默认 html:true 不消毒；vault 写入方含外部 agent（AI 产出不可信），
  // 编辑器实时预览与 TaskDetailModal 等四处渲染同口径必须过 DOMPurify
  function safeHtml(html: string) {
    return DOMPurify.sanitize(html);
  }

  onMounted(async () => {
    await loadTree();
    await tryOpenQueryPath();
  });

  // 同视图内的关联文档跳转只变 query 不换路由，onMounted 不重跑
  watch(() => route.query.path, async (qpath, old) => {
    if (qpath === old || typeof qpath !== 'string' || !qpath) return;
    if (dirty.value && !(await confirmDiscard('打开链接文档'))) return;
    await tryOpenQueryPath();
  });

  // 知识库/文档两路由共用本组件：页内切换时组件复用不重建，
  // onMounted 不会重跑——必须监听视图模式变化重载树并清理编辑状态
  watch(isKnowledge, async () => {
    if (dirty.value && !(await confirmDiscard('切换视图'))) return;
    selectedKeys.value = [];
    currentFile.value = null;
    editContent.value = '';
    markOpened('');
    cancelInline();
    loadTree(false);
  });

  // 路由离开守卫：编辑中点菜单跳走需确认
  onBeforeRouteLeave(async () => {
    if (dirty.value && !(await confirmDiscard('离开页面'))) return false;
    return true;
  });
</script>

<style lang="less" scoped>
  // 整页铺满内容区视口：双栏等高、树与编辑器在卡内滚动
  .docs-page {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  .docs-grid {
    flex: 1;
    min-height: 0;

    // grid item 默认高度自适应内容，需显式拉满供卡内 flex 链使用
    :deep(.n-grid-item) {
      height: 100%;
    }
  }

  .docs-col {
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

    // spin 内容层也要撑满，空状态才能在可视区居中
    .dir-card :deep(.n-spin-content) {
      min-height: 100%;
      display: flex;
      flex-direction: column;
    }

    // 编辑器铺满剩余高度
    .docs-editor {
      flex: 1;
      min-height: 0;
    }
  }

  .search-banner {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 6px;
    padding: 3px 8px;
    border-radius: 6px;
    background: var(--hover-bg);
    font-size: 12px;
    color: var(--text-2, #57606a);
  }

  // docx 预览排版（Word 转出的 HTML 无样式，补基础阅读版式）
  .docx-wrap {
    flex: 1;
    min-height: 0;
    overflow-y: auto;

    .docx-spin {
      margin: 80px auto;
    }

    .docx-html {
      max-width: 860px;
      margin: 0 auto;
      padding: 16px 24px;
      line-height: 1.75;
      color: var(--text-1, #24292f);

      :deep(h1) { font-size: 22px; margin: 20px 0 12px; }
      :deep(h2) { font-size: 18px; margin: 18px 0 10px; }
      :deep(h3) { font-size: 15px; margin: 14px 0 8px; }
      :deep(p) { margin: 8px 0; }
      :deep(table) {
        border-collapse: collapse;
        margin: 12px 0;
        width: 100%;

        td, th { border: 1px solid var(--line, #e9e9e7); padding: 6px 10px; font-size: 12.5px; }
      }
      :deep(img) { max-width: 100%; }
    }
  }

  .img-wrap {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    display: flex;
    justify-content: center;
    padding: 16px;

    img {
      max-width: 100%;
      object-fit: contain;
    }
  }

  .bin-meta {
    font-size: 12px;
    color: var(--text-3, #8b949e);
  }
</style>
