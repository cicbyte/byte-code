<template>
  <div>

    <n-card :bordered="false" class="proCard">
      <template #header>
        <n-space align="center">
          <n-select
            v-model:value="filterModule"
            :options="moduleOptions"
            placeholder="按模块筛选"
            style="width: 160px"
            clearable
            @update:value="onFilterChange"
          />
        </n-space>
      </template>

      <n-spin :show="loading">
        <EmptyState type="notify" title="暂无活动记录" v-if="!loading && activityList.length === 0" description="项目的操作动态会实时出现在这里" />
        <n-timeline v-else>
          <n-timeline-item
            v-for="item in activityList"
            :key="item.id"
            :type="timelineType(item.action)"
            :time="item.createdAt"
            :class="{ 'act-clickable': jumpable(item) }"
            @click="jumpTarget(item)"
          >
            <template #header>
              <n-space size="small" align="center">
                <span class="font-medium">{{ item.actorName }}</span>
                <n-tag size="tiny" v-if="item.targetType">{{ TARGET_TYPE_LABELS[item.targetType] || item.targetType }}</n-tag>
              </n-space>
            </template>
            <div class="text-sm text-gray-600">
              {{ actionText(item.action) }}
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
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { useRouter } from 'vue-router';
  import { ACTION_LABELS, actionText, TARGET_TYPE_LABELS } from '@/enums/activity';
  import { ref, reactive, onMounted } from 'vue';
  import { getActivities } from '@/api/platform/index';
  import type { ActivityItem } from '@/api/platform/index';

  const loading = ref(false);
  const activityList = ref<ActivityItem[]>([]);
  const total = ref(0);
  const filterModule = ref<string | null>(null);
  const pagination = reactive({ page: 1, size: 20 });

  const moduleOptions = [
    { label: '需求', value: 'requirement' },
    { label: '项目', value: 'project' },
    { label: '任务', value: 'task' },
    { label: '测试', value: 'test' },
    { label: '知识库', value: 'knowledge' },
    { label: 'Agent', value: 'ai' },
  ];

  const router = useRouter();

  // 动态条目跳转：target 是实体时可点直达（task 开抽屉、topic/feedback 跳
  // 专题/反馈页、project 跳概览、test_plan_case 跳测试计划）
  function jumpTarget(item: any) {
    const pid = item.projectId;
    if (!pid || !item.targetId) return;
    switch (item.targetType) {
      case 'task': router.push(`/project/${pid}/tasks?task=${item.targetId}`); break;
      case 'topic': router.push(`/project/${pid}/topics`); break;
      case 'feedback': router.push(`/project/${pid}/feedbacks`); break;
      case 'project': router.push(`/project/${item.targetId}/overview`); break;
      case 'test_plan_case': router.push(`/project/${pid}/test-plans`); break;
    }
  }
  function jumpable(item: any): boolean {
    return !!item.targetId && ['task', 'topic', 'feedback', 'project', 'test_plan_case'].includes(item.targetType);
  }

  function timelineType(action: string): 'default' | 'info' | 'success' | 'warning' | 'error' {
    if (action.includes('创建') || action.includes('create')) return 'success';
    if (action.includes('删除') || action.includes('delete')) return 'error';
    if (action.includes('更新') || action.includes('update')) return 'info';
    return 'default';
  }

    // 筛选变更从第 1 页重查：第 N 页改筛选会请求空页显示"暂无"
  function onFilterChange() {
    pagination.page = 1;
    loadData();
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

<style lang="less" scoped>
  :deep(.act-clickable) {
    cursor: pointer;
    transition: background 0.15s;
    border-radius: 8px;
    padding: 4px 6px;
    margin: 0 -6px;

    &:hover {
      background: var(--hover-bg, rgba(0, 0, 0, 0.03));
    }
  }
</style>
