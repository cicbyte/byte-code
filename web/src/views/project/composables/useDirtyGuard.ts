import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useDialog } from 'naive-ui';
import type { Ref, ComputedRef } from 'vue';
import type { DocsFile } from '@/api/docs/index';

// 未保存守卫：dirty 检测（编辑内容 vs 打开时快照）+ 路由/浏览器关闭拦截。
// 拦截点（切换文件/切视图/路由离开）由页面自行调用 confirmDiscard

export interface DirtyGuardDeps {
  currentFile: Ref<DocsFile | null>;
  editContent: Ref<string>;
}

export function useDirtyGuard(deps: DirtyGuardDeps) {
  const dialog = useDialog();

  // 编辑内容与打开时快照不一致即脏；二进制/未打开视为干净
  const openedSnapshot = ref('');
  const dirty = computed(
    () =>
      !!deps.currentFile.value &&
      !deps.currentFile.value.binary &&
      deps.editContent.value !== openedSnapshot.value
  );

  function markOpened(content: string) {
    openedSnapshot.value = content;
  }

  // 三选项确认：返回 true=放弃更改继续
  function confirmDiscard(action: string): Promise<boolean> {
    return new Promise((resolve) => {
      dialog.warning({
        title: '有未保存的更改',
        content: `当前文档已修改未保存，${action}将丢弃更改。`,
        positiveText: '放弃更改',
        negativeText: '留在本页',
        onPositiveClick: () => resolve(true),
        onNegativeClick: () => resolve(false),
        onClose: () => resolve(false),
      });
    });
  }

  // 浏览器关闭/刷新兜底（原生确认框，SPA 内无法用自定义弹窗）
  const beforeUnloadGuard = (e: BeforeUnloadEvent) => {
    if (!dirty.value) return;
    e.preventDefault();
    e.returnValue = '';
  };
  onMounted(() => window.addEventListener('beforeunload', beforeUnloadGuard));
  onUnmounted(() => window.removeEventListener('beforeunload', beforeUnloadGuard));

  return { dirty, markOpened, confirmDiscard };
}
