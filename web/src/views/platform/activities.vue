<template>
  <div>
    <div class="n-layout-page-header">
      <n-card :bordered="false" title="活动流" />
    </div>

    <n-card :bordered="false" class="mt-4 proCard">
      <template #header>
        <n-space align="center">
          <n-select
            v-model:value="filterModule"
            :options="moduleOptions"
            placeholder="按模块筛选"
            style="width: 160px"
            clearable
            @update:value="loadData"
          />
        </n-space>
      </template>

      <n-spin :show="loading">
        <n-empty v-if="!loading && activityList.length === 0" description="暂无活动记录" />
        <n-timeline v-else>
          <n-timeline-item
            v-for="item in activityList"
            :key="item.id"
            :type="timelineType(item.action)"
            :time="item.createdAt"
          >
            <template #header>
              <n-space size="small" align="center">
                <span class="font-medium">{{ item.actorName }}</span>
                <n-tag size="tiny" v-if="item.targetType">{{ item.targetType }}</n-tag>
              </n-space>
            </template>
            <div class="text-sm text-gray-600">
              {{ item.action }}
              <span v-if="item.targetName" class="text-blue-500">「{{ item.targetName }}」</span>
            </div>
            <div v-if="item.detail" class="text-xs text-gray-400 mt-1">{{ item.detail }}</div>
          </n-timeline-item>
        </n-timeline>
      </n-spin>

      <div class="mt-4 flex justify-end" v-if="total > pagination.size">
        <n-pagination
          v-model:page="pagination.page"
          :page-size="pagination.size"
          :item-count="total"
          @update:page="loadData"
        />
      </div>
    </n-card>
  </div>
</template>

<script lang="ts" setup>
  import { ref, reactive, onMounted } from 'vue';
  import { getActivities } from '@/api/platform/index';
  import type { ActivityItem } from '@/api/platform/index';

  const loading = ref(false);
  const activityList = ref<ActivityItem[]>([]);
  const total = ref(0);
  const filterModule = ref<string | null>(null);
  const pagination = reactive({ page: 1, size: 20 });

  const moduleOptions = [
    { label: '产品', value: 'product' },
    { label: '需求', value: 'requirement' },
    { label: '项目', value: 'project' },
    { label: '任务', value: 'task' },
    { label: '测试', value: 'test' },
    { label: '知识库', value: 'knowledge' },
    { label: 'AI', value: 'ai' },
  ];

  function timelineType(action: string): 'default' | 'info' | 'success' | 'warning' | 'error' {
    if (action.includes('创建') || action.includes('create')) return 'success';
    if (action.includes('删除') || action.includes('delete')) return 'error';
    if (action.includes('更新') || action.includes('update')) return 'info';
    return 'default';
  }

  async function loadData() {
    loading.value = true;
    try {
      const res = await getActivities({
        module: filterModule.value ?? undefined,
        page: pagination.page,
        size: pagination.size,
      });
      if (res) {
        activityList.value = res.list || [];
        total.value = res.total || 0;
      }
    } catch (e) {
      // ignore
    } finally {
      loading.value = false;
    }
  }

  onMounted(() => {
    loadData();
  });
</script>
