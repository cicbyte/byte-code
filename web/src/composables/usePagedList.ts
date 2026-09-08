import { ref, reactive } from 'vue';

/**
 * 分页列表统一 composable（审计 M1/M2/M4 一次性收敛）：
 * - 筛选变更自动重置页码（历史修复模式 9 页各写一遍的坑）
 * - 删除后末页删空自动回退一页（此前全军覆没的越界场景）
 * - loading/list/total/pagination 四件套 + 错误静默的统一出口
 * - 分页参数命名在 fetcher 内消化（后端 page/size 与 pageNum/pageSize
 *   两套并存的差异，收敛到各页 api 函数内部映射，视图层统一 page/size）
 *
 * 用法：
 *   const { loading, list, total, pagination, load, onFilterChange, afterRemove } = usePagedList(fetcher)
 *   fetcher: (page: number, size: number) => Promise<{ list: T[]; total: number }>
 */
export function usePagedList<T = any>(
  fetcher: (page: number, size: number) => Promise<{ list: T[]; total: number }>,
  defaultSize = 20
) {
  const loading = ref(false);
  const list = ref<T[]>([]) as any;
  const total = ref(0);
  const pagination = reactive({ page: 1, size: defaultSize });
  let loadSeq = 0;

  async function load() {
    const seq = ++loadSeq;
    loading.value = true;
    try {
      const res = await fetcher(pagination.page, pagination.size);
      if (seq !== loadSeq) return; // 过期响应（筛选/翻页快速切换）
      list.value = res?.list || [];
      total.value = res?.total || 0;
      // 末页删空回退：停在第 N 页删掉最后一条后 page 越界 → 回退重查一次
      if (list.value.length === 0 && pagination.page > 1) {
        pagination.page -= 1;
        return load();
      }
    } catch (e) {
      if (seq === loadSeq) list.value = [];
      throw e;
    } finally {
      if (seq === loadSeq) loading.value = false;
    }
  }

  /** 筛选条件变更入口：重置页码后重查（视图层 @update:value 统一绑这里） */
  function onFilterChange() {
    pagination.page = 1;
    return load();
  }

  /** 删除成功后的统一回调：满页删除由 load 的末页回退兜底 */
  function afterRemove() {
    return load();
  }

  function onPageChange(p: number) {
    pagination.page = p;
    return load();
  }

  return { loading, list, total, pagination, load, onFilterChange, afterRemove, onPageChange };
}
