import { reactive, ref, h } from 'vue';
import { NInput, useMessage } from 'naive-ui';
import type { Ref, ComputedRef } from 'vue';
import { writeDocsFile, createDocsFolder, moveDocsPath } from '@/api/docs/index';
import type { DocsTreeNode } from '@/api/docs/index';

// 树内原位输入框（新建/重命名）：Enter/失焦提交，Esc 取消，替代弹窗输全路径

export const EDIT_KEY = '__inline-editing__';

/** 空间守卫断言（由页面注入：知识库/文档视图互斥的路径前缀校验） */
export interface InlineEditDeps {
  treeData: Ref<any[]>;
  expandedKeys: Ref<string[]>;
  projectId: ComputedRef<number>;
  isKnowledge: ComputedRef<boolean>;
  assertSpaceAllowed: (p: string) => boolean;
  loadTree: () => Promise<void>;
}

export function useInlineEdit(deps: InlineEditDeps) {
  const message = useMessage();

  const editing = reactive({
    active: false,
    mode: '' as '' | 'new-doc' | 'new-folder' | 'rename',
    dir: '', // 新建：目标父目录；重命名：节点所在目录
    path: '', // 重命名：原节点 path
    value: '',
  });

  function renderNodeInput() {
    return h(NInput, {
      size: 'tiny',
      value: editing.value,
      autofocus: true,
      placeholder: editing.mode === 'new-folder' ? '文件夹名' : '名称（.md 可省略）',
      onFocus: () => {
        // 重命名时全选原名便于整体替换
        const el = document.querySelector('.n-tree .n-input input') as HTMLInputElement | null;
        if (el && editing.mode === 'rename') el.select();
      },
      'onUpdate:value': (v: string) => (editing.value = v),
      onKeydown: (e: KeyboardEvent) => {
        if (e.key === 'Enter') commitInline();
        else if (e.key === 'Escape') cancelInline();
      },
      onBlur: () => {
        // Esc 取消后仍会触发一次 blur，此时不再提交
        if (editing.active) commitInline();
      },
    });
  }

  function nodeLabel(node: DocsTreeNode) {
    // 恒返回函数：编辑态判断必须在渲染时求值（而非 transformTree 构建时），
    // 否则 startInlineRename 只改 editing 状态不会触发 label 重渲染。
    // 节点只显示文件名：frontmatter title（常与正文 H1 相同）在树里冗余
    return () => {
      if (editing.active && editing.mode === 'rename' && node.path === editing.path) {
        return renderNodeInput();
      }
      return node.name;
    };
  }

  // 在树上插入待命名节点并进入编辑
  function startInlineCreate(dir: string, type: 'new-doc' | 'new-folder') {
    if (editing.active) cancelInline();
    editing.active = true;
    editing.mode = type;
    editing.dir = dir;
    editing.path = '';
    editing.value = '';
    const pending = {
      key: EDIT_KEY,
      label: () => renderNodeInput(),
      isLeaf: true,
      prefix: () => (type === 'new-folder' ? '📁' : '📄'),
    };
    if (!dir) {
      deps.treeData.value.unshift(pending);
    } else {
      const insertInto = (list: any[]): boolean =>
        list.some((item) => {
          if (item.key === dir) {
            item.children = item.children || [];
            item.children.unshift({ ...pending, key: EDIT_KEY });
            deps.expandedKeys.value = Array.from(new Set([...deps.expandedKeys.value, dir]));
            return true;
          }
          return item.children ? insertInto(item.children) : false;
        });
      if (!insertInto(deps.treeData.value)) deps.treeData.value.unshift(pending);
    }
  }

  function startInlineRename(nodePath: string) {
    if (editing.active) cancelInline();
    editing.active = true;
    editing.mode = 'rename';
    editing.path = nodePath;
    editing.dir = nodePath.slice(0, nodePath.lastIndexOf('/')) || '';
    editing.value = nodePath.split('/').pop() || '';
  }

  function cancelInline() {
    editing.active = false;
    editing.mode = '';
    // 移除 pending 节点（新建）
    const removePending = (list: any[]): void => {
      const i = list.findIndex((item) => item.key === EDIT_KEY);
      if (i >= 0) list.splice(i, 1);
      list.forEach((item) => item.children && removePending(item.children));
    };
    removePending(deps.treeData.value);
    // 重命名：label 恢复由响应式重渲染（editing.path 清空）
    editing.path = '';
  }

  async function commitInline() {
    if (!editing.active) return;
    const name = editing.value.trim();
    const { mode, dir } = editing;
    if (!name || name.includes('/') || name.includes('..')) {
      cancelInline();
      return;
    }
    // 空间守卫：目标路径必须落在当前视图空间内
    const targetPath = dir ? `${dir}/${name}` : name;
    if (!deps.assertSpaceAllowed(targetPath)) {
      cancelInline();
      return;
    }
    editing.active = false; // 先关编辑态，防 blur 二次提交
    try {
      if (mode === 'new-doc') {
        const file = name.toLowerCase().endsWith('.md') ? name : `${name}.md`;
        const title = name.replace(/\.md$/i, '');
        const frontmatter = deps.isKnowledge.value
          ? `---\ntitle: ${title}\nspace: knowledge\nstatus: draft\n---\n\n`
          : '';
        await writeDocsFile(deps.projectId.value, dir ? `${dir}/${file}` : file, frontmatter);
      } else if (mode === 'new-folder') {
        await createDocsFolder(deps.projectId.value, dir ? `${dir}/${name}` : name);
      } else if (mode === 'rename') {
        const old = editing.path;
        editing.path = '';
        await moveDocsPath(deps.projectId.value, old, dir ? `${dir}/${name}` : name);
      }
      await deps.loadTree();
    } catch {
      message.error('操作失败');
      await deps.loadTree();
    }
  }

  return { editing, nodeLabel, startInlineCreate, startInlineRename, cancelInline, commitInline };
}
