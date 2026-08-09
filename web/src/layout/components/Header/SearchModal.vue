<template>
  <n-modal v-model:show="show" preset="card" title="全局搜索" style="width: 560px">
    <n-input
      v-model:value="keyword"
      placeholder="搜索任务 / 需求 / 文档 / 测试用例，回车搜索"
      clearable
      @keyup.enter="doSearch"
    >
      <template #prefix>
        <n-icon><SearchOutlined /></n-icon>
      </template>
    </n-input>
    <n-spin :show="loading" size="small">
      <div v-if="searched && results.length === 0" class="search-empty">没有找到相关内容</div>
      <div v-for="group in grouped" :key="group.module" class="search-group">
        <div class="search-group-title">
          <n-tag size="small" :type="moduleTag(group.module)">{{ moduleLabel(group.module) }}</n-tag>
          <span class="search-group-count">{{ group.items.length }} 条</span>
        </div>
        <div v-for="item in group.items" :key="item.id" class="search-item" @click="go(item)">
          <span class="search-item-title">{{ item.title }}</span>
          <span class="search-item-id">#{{ item.id }}</span>
        </div>
      </div>
    </n-spin>
  </n-modal>
</template>

<script lang="ts" setup>
  import { computed, nextTick, ref } from 'vue';
  import { useRouter } from 'vue-router';
  import { SearchOutlined } from '@vicons/antd';
  import { search } from '@/api/platform';
  import type { SearchResult } from '@/api/platform';

  const router = useRouter();

  const show = ref(false);
  const keyword = ref('');
  const loading = ref(false);
  const searched = ref(false);
  const results = ref<SearchResult[]>([]);

  const grouped = computed(() => {
    const order = ['task', 'requirement', 'doc', 'test_case'];
    const map = new Map<string, SearchResult[]>();
    for (const r of results.value) {
      if (!map.has(r.module)) map.set(r.module, []);
      map.get(r.module)!.push(r);
    }
    return order.filter((m) => map.has(m)).map((m) => ({ module: m, items: map.get(m)! }));
  });

  function open() {
    show.value = true;
    nextTick(() => {
      // 弹窗渲染后聚焦输入框
      const el = document.querySelector('.n-modal .n-input input') as HTMLInputElement | null;
      el?.focus();
    });
  }

  async function doSearch() {
    const q = keyword.value.trim();
    if (!q) return;
    loading.value = true;
    searched.value = true;
    try {
      const res = await search({ q, page: 1, size: 20 });
      results.value = res?.list || [];
    } catch {
      results.value = [];
    } finally {
      loading.value = false;
    }
  }

  function go(item: SearchResult) {
    show.value = false;
    switch (item.module) {
      case 'task':
        router.push(`/project/${item.projectId}/tasks`);
        break;
      case 'requirement':
        router.push(`/project/${item.projectId}/requirements`);
        break;
      case 'doc':
        router.push('/knowledge');
        break;
      case 'test_case':
        router.push('/test');
        break;
    }
  }

  function moduleLabel(m: string): string {
    const labels: Record<string, string> = {
      task: '任务',
      requirement: '需求',
      doc: '文档',
      test_case: '测试用例',
    };
    return labels[m] || m;
  }

  function moduleTag(m: string): 'default' | 'info' | 'success' | 'warning' {
    const tags: Record<string, 'default' | 'info' | 'success' | 'warning'> = {
      task: 'info',
      requirement: 'success',
      doc: 'warning',
      test_case: 'default',
    };
    return tags[m] || 'default';
  }

  defineExpose({ open });
</script>

<style lang="less" scoped>
  .search-empty {
    padding: 24px 0;
    color: #999;
    text-align: center;
  }

  .search-group {
    margin-top: 12px;
  }

  .search-group-title {
    display: flex;
    gap: 8px;
    align-items: center;
    margin-bottom: 4px;
  }

  .search-group-count {
    color: #999;
    font-size: 12px;
  }

  .search-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 7px 10px;
    cursor: pointer;
    border-radius: 4px;

    &:hover {
      background: rgba(0, 0, 0, 0.03);
    }
  }

  .search-item-title {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .search-item-id {
    flex: none;
    margin-left: 12px;
    color: #bbb;
    font-size: 12px;
  }
</style>
