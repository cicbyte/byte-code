import type { ToolbarNames } from 'md-editor-v3';

// 平台统一 MdEditor 精选工具栏：文法高频项 + 预览/页内全屏/全屏。
// 勿改回 toolbarsExclude 全量默认——窄容器下尾部按钮会被裁掉（全屏等不可见）。
export const mdToolbars: ToolbarNames[] = [
  'bold', 'italic', 'title', 'quote',
  'unorderedList', 'orderedList', 'task',
  'codeRow', 'code', 'link', 'image', 'table',
  '-',
  'preview', 'pageFullscreen', 'fullscreen',
];

// 编辑类抽屉统一宽度：与任务详情抽屉同款（880 上限、92vw 小屏自适应）。
// 编辑器工具栏按钮多，480~560px 会把尾部按钮裁掉。
export function editDrawerWidth() {
  return typeof window !== 'undefined' ? Math.min(880, window.innerWidth * 0.92) : 720;
}
