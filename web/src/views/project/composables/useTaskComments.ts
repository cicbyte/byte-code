import { ref, onMounted, onUnmounted } from 'vue';
import { getComments } from '@/api/project/index';
import type { CommentItem } from '@/api/project/index';

/**
 * 任务评论区 composable（TaskDetailModal 拆分第一步）：
 * 评论加载 + SSE 实时刷新（bc-notification 事件 → 标题精确匹配当前任务的
 * 评论类通知 → 静默重拉并贴底）。历史 bug 高发区（标题字符串匹配是已知
 * 脆弱点，集中到此处便于将来换结构化字段时一处改）。
 *
 * 用法：const { comments, refreshComments } = useTaskComments(() => task.value?.id ?? 0)
 * currentTaskId 取 0 时不响应 SSE。
 */
export function useTaskComments(currentTaskId: () => number, onRefreshed?: () => void) {
  const comments = ref<CommentItem[]>([]);

  async function refreshComments(silent = false) {
    const tid = currentTaskId();
    if (!tid) return;
    try {
      const res = await getComments(tid);
      comments.value = res?.list || [];
      if (silent) onRefreshed?.();
    } catch {
      // ignore：错误提示由 http 层统一处理
    }
  }

  // SSE：仅评论类标题 + 当前任务匹配时刷新（标题精确匹配的脆弱性已在
  // 平台侧记录为待结构化字段，换字段时只改这里）
  const COMMENT_TITLES = ['任务新评论', '评论被回复', '评论提及了你'];
  function onLiveNotification(e: Event) {
    const detail: any = (e as CustomEvent).detail;
    if (!detail) return;
    if (detail.sourceType === 'task' && Number(detail.sourceId) === currentTaskId()) {
      if (COMMENT_TITLES.includes(detail.title)) {
        refreshComments(true);
      }
    }
  }

  onMounted(() => window.addEventListener('bc-notification', onLiveNotification));
  onUnmounted(() => window.removeEventListener('bc-notification', onLiveNotification));

  return { comments, refreshComments };
}
