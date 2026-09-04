import { reactive, nextTick } from 'vue';
import type { ComputedRef, Ref } from 'vue';

// 树右键菜单：空白区/目录/文件三场景按需渲染；动作由页面注入

export interface CtxTarget {
  path: string;
  isDir: boolean;
  hasChildren: boolean;
  name: string;
}

/** 右键菜单触发的页面动作（新建/上传/打开等，全部由 docs.vue 提供） */
export interface CtxMenuActions {
  startInlineCreate: (dir: string, type: 'new-doc' | 'new-folder') => void;
  startInlineRename: (path: string) => void;
  openFile: (path: string) => Promise<void>;
  openMoveModal: (path: string) => void;
  handleDeleteFor: (path: string, isDir: boolean) => void;
  handleRefresh: () => Promise<void>;
  /** 触发文件选择器，并设定上传落点目录 */
  triggerUpload: (dir: string) => void;
}

export interface CtxMenuDeps {
  isKnowledge: ComputedRef<boolean>;
  selectedKeys: Ref<string[]>;
  actions: CtxMenuActions;
}

export function useCtxMenu(deps: CtxMenuDeps) {
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
    if (option.isLeaf) deps.selectedKeys.value = [String(option.key)];
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
        { label: '重扫索引', key: 'refresh', divider: true },
      ];
    } else if (target.isDir) {
      ctxMenu.options = [
        { label: `在「${target.name}」中新建文档`, key: 'new-doc' },
        { label: `在「${target.name}」中新建文件夹`, key: 'new-folder' },
        { label: `上传到「${target.name}」`, key: 'upload' },
        { label: '重命名', key: 'rename' },
        { label: '移动到…', key: 'move' },
        // 非空目录禁止删除（后端同样校验），防止误删整棵子树
        { label: '删除（仅空目录）', key: 'delete', disabled: target.hasChildren },
        { label: '重扫索引', key: 'refresh', divider: true },
      ];
    } else {
      ctxMenu.options = [
        { label: '打开', key: 'open' },
        { label: '重命名', key: 'rename', divider: true },
        { label: '移动到…', key: 'move' },
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
    // 落点：目录右键 → 该目录；文件右键 → 其父目录；空白 → 视图默认根
    // （知识库页默认根是 知识库/，否则新建路径缺前缀会被空间守卫拒绝——空知识库时
    //   空白右键是唯一创建入口，必须可用）
    const defaultDir = deps.isKnowledge.value ? '知识库' : '';
    const dir = t ? (t.isDir ? t.path : t.path.slice(0, t.path.lastIndexOf('/')) || '') : defaultDir;
    switch (key) {
      case 'new-doc':
        deps.actions.startInlineCreate(dir, 'new-doc');
        break;
      case 'new-folder':
        deps.actions.startInlineCreate(dir, 'new-folder');
        break;
      case 'upload':
        deps.actions.triggerUpload(dir);
        break;
      case 'open':
        if (t) deps.actions.openFile(t.path);
        break;
      case 'rename':
        if (t) deps.actions.startInlineRename(t.path);
        break;
      case 'move':
        if (t) deps.actions.openMoveModal(t.path);
        break;
      case 'refresh':
        deps.actions.handleRefresh();
        break;
      case 'delete':
        if (t) deps.actions.handleDeleteFor(t.path, t.isDir);
        break;
    }
  }

  return { ctxMenu, treeNodeProps, onBlankContextMenu, onCtxSelect };
}
