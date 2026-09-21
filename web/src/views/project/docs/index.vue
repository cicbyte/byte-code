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
           data-test-id="project-docs.search-input">
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
            <template v-else>
              <!-- 勾选后出现的批量归类操作条 -->
              <div v-if="checkedKeys.length" class="batch-bar" data-test-id="project-docs.batch-bar">
                <span class="text-xs">已选 {{ checkedKeys.length }} 项</span>
                <n-button size="tiny" type="primary" @click="openBatchMove" data-test-id="project-docs.batch-move-btn">
                  移动到…
                </n-button>
                <n-button size="tiny" quaternary @click="checkedKeys = []">清空</n-button>
              </div>
              <n-tree
                block-line
                checkable
                :cascade="false"
                :data="treeData"
                v-model:checked-keys="checkedKeys"
                :selected-keys="selectedKeys"
                :expanded-keys="expandedKeys"
                :node-props="treeNodeProps"
                @update:selected-keys="onSelectNode"
                @update:expanded-keys="onExpandNode"
              />
            </template>
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
              <n-tag v-for="t in currentFile.meta?.tags || []" :key="t" size="small" :data-test-id="`project-docs.item-${t}`">{{ t }}</n-tag>
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

            <!-- 附件：挂文档路径键（entity_key = projectId:path），与任务附件同款交互 -->
            <n-divider style="margin: 14px 0 8px" />
            <n-space :size="8" align="center" class="mb-2">
              <span class="text-sm font-medium">附件（{{ docAttachments.length }}）</span>
              <n-button size="tiny" :loading="uploadingDocAtt" @click="docAttInputRef?.click()">
                上传附件
              </n-button>
              <span class="text-xs text-gray-400">截图/设计稿/参考资料挂在本文档上</span>
            </n-space>
            <div v-if="docAttachments.length === 0" class="text-xs text-gray-400 mb-2">暂无附件</div>
            <n-space v-for="a in docAttachments" :key="a.id" justify="space-between" align="center" class="w-full doc-att-row" :data-test-id="`project-docs.item-${a.id}`">
              <n-space :size="8" align="center">
                <span class="text-sm">{{ a.originalName }}</span>
                <span class="text-xs text-gray-400">{{ formatSize(a.fileSize) }}</span>
                <span class="text-xs text-gray-400">{{ a.uploaderName || '' }}</span>
              </n-space>
              <n-space :size="2">
                <n-button text type="info" size="tiny" @click="downloadDocAtt(a)">下载</n-button>
                <n-button text type="error" size="tiny" @click="removeDocAtt(a)">删除</n-button>
              </n-space>
            </n-space>
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

    <!-- 批量归类移动：勾选多项后移入目标目录（保持原名） -->
    <n-modal
      v-model:show="showBatchMove"
      preset="dialog"
      title="批量移动到目录"
      positive-text="移动"
      negative-text="取消"
      @positive-click="handleBatchMoveSubmit"
      style="width: 520px"
    >
      <n-form label-placement="top" class="py-3">
        <n-form-item :label="`将 ${batchTargets.length} 项移动到目录`">
          <n-tree-select
            v-model:value="batchDest"
            :options="dirOptions"
            key-field="key"
            label-field="label"
            default-expand-all
            placeholder="选择目标目录"
            data-test-id="project-docs.batch-target-select"
          />
        </n-form-item>
        <n-form-item label="新建子目录（可选，不存在会自动创建）">
          <n-input
            v-model:value="batchSubdir"
            placeholder="例：ops（留空则直接放入所选目录）"
            data-test-id="project-docs.batch-subdir-input"
          />
        </n-form-item>
        <div class="text-xs text-gray-400">
          保持原文件名移入；目标已存在同名项的将跳过并在结果中提示。勾选了目录时其内容随之移动。
        </div>
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
    <input ref="docAttInputRef" type="file" style="display: none" multiple @change="onDocAttFiles" />

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
  import { getAttachments, uploadAttachment, downloadAttachment, deleteAttachment } from '@/api/attachment/index';
  import { useInlineEdit } from '../composables/useInlineEdit';
  import { useCtxMenu } from '../composables/useCtxMenu';
  import { useBinaryPreview } from '../composables/useBinaryPreview';
  import { useDirtyGuard } from '../composables/useDirtyGuard';
  import DocsHistoryModal from '../components/DocsHistoryModal.vue';

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

  // ==================== 批量归类移动（勾选多项 → 移入目录） ====================
  const checkedKeys = ref<string[]>([]);
  const showBatchMove = ref(false);
  const batchDest = ref<string | null>(null);
  const batchSubdir = ref('');

  // 目录树（仅目录节点）作为移动目标选项；根目录 = 移到顶层
  const dirOptions = computed(() => {
    const walk = (nodes: any[]): any[] => {
      const out: any[] = [];
      for (const n of nodes) {
        if (n.isDir || dirPaths.has(n.key as string)) {
          out.push({ key: n.key, label: n.label, children: walk(n.children || []) });
        }
      }
      return out;
    };
    return [{ key: '', label: '（顶层）', children: walk(treeData.value) }];
  });

  // 勾选目标的祖先去重：选了目录就跳过其子孙（目录整体移动已带内容）
  function dedupeAncestors(keys: string[]): string[] {
    return keys.filter((k) => !keys.some((o) => o !== k && k.startsWith(o + '/')));
  }
  const batchTargets = computed(() => dedupeAncestors(checkedKeys.value));

  // 选中项的公共父目录：归类场景默认留在原地所在目录（勾选 → 只填新子目录名）
  function commonParentDir(keys: string[]): string {
    if (!keys.length) return '';
    const dirs = keys.map((k) => k.split('/').slice(0, -1).join('/'));
    const parts = dirs[0].split('/').filter(Boolean);
    const prefix: string[] = [];
    for (const p of parts) {
      const cand = [...prefix, p].join('/');
      if (dirs.every((d) => d === cand || d.startsWith(cand + '/'))) prefix.push(p);
      else break;
    }
    return prefix.join('/');
  }

  function openBatchMove() {
    if (!checkedKeys.value.length) return;
    batchDest.value = commonParentDir(batchTargets.value) || null;
    batchSubdir.value = '';
    showBatchMove.value = true;
  }

  async function handleBatchMoveSubmit() {
    const subdir = batchSubdir.value.trim().replace(/^\/+|\/+$/g, '');
    if (subdir && (subdir.includes('..') || subdir.startsWith('/'))) {
      message.error('子目录名不合法');
      return false;
    }
    const dest = [batchDest.value || '', subdir].filter(Boolean).join('/');
    if (!assertSpaceAllowed(dest)) return false;
    let ok = 0;
    const failed: string[] = [];
    for (const from of batchTargets.value) {
      const name = from.split('/').pop() as string;
      const to = dest ? `${dest}/${name}` : name;
      try {
        await moveDocsPath(projectId.value, from, to);
        ok++;
      } catch {
        failed.push(name);
      }
    }
    showBatchMove.value = false;
    checkedKeys.value = [];
    if (failed.length) {
      message.warning(`移动完成：成功 ${ok} 项，跳过 ${failed.length} 项（目标已存在）：${failed.join('、')}`);
    } else {
      message.success(`已移动 ${ok} 项`);
    }
    loadTree();
  }

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
      loadDocAttachments();
    } catch {
      message.error('加载文件失败');
    }
  }

  // ==================== 文档附件（挂路径键 projectId:path） ====================
  const docAttInputRef = ref<HTMLInputElement | null>(null);
  const docAttachments = ref<any[]>([]);
  const uploadingDocAtt = ref(false);

  function docEntityKey(): string {
    return currentFile.value ? `${projectId.value}:${currentFile.value.path}` : '';
  }

  async function loadDocAttachments() {
    const key = docEntityKey();
    if (!key) {
      docAttachments.value = [];
      return;
    }
    try {
      const res = await getAttachments('doc', 0, key);
      docAttachments.value = res?.list || [];
    } catch {
      docAttachments.value = [];
    }
  }

  async function onDocAttFiles(e: Event) {
    const input = e.target as HTMLInputElement;
    const files = Array.from(input.files || []);
    const key = docEntityKey();
    if (!files.length || !key) return;
    uploadingDocAtt.value = true;
    for (const f of files) {
      try {
        await uploadAttachment({ entityType: 'doc', entityId: 0, entityKey: key, file: f });
      } catch {
        message.error(`上传失败：${f.name}`);
      }
    }
    uploadingDocAtt.value = false;
    input.value = '';
    loadDocAttachments();
  }

  async function downloadDocAtt(a: any) {
    try {
      const res = await downloadAttachment(a.id);
      if (res?.url) window.open(res.url, '_blank');
    } catch {
      message.error('获取下载链接失败');
    }
  }

  async function removeDocAtt(a: any) {
    try {
      await deleteAttachment(a.id);
      docAttachments.value = docAttachments.value.filter((x: any) => x.id !== a.id);
    } catch {
      message.error('删除附件失败');
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

  // 勾选后的批量归类操作条
  .batch-bar {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-bottom: 6px;
    padding: 3px 8px;
    border-radius: 6px;
    background: var(--primary-color-hover, rgba(22, 163, 74, 0.12));
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
