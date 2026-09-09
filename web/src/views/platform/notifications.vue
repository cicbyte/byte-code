<template>
  <div>
    <n-card :bordered="false" class="proCard">
      <template #header>
        <n-space :size="8" align="center">
          <span>通知中心</span>
          <n-badge :value="unread" :max="99" show-zero>
            <n-tag size="small" :bordered="false">未读</n-tag>
          </n-badge>
        </n-space>
      </template>
      <template #header-extra>
        <n-button size="small" :disabled="unread === 0" @click="handleReadAll">全部已读</n-button>
      </template>

      <n-space class="mb-4" align="center">
        <n-radio-group v-model:value="readFilter" size="small" @update:value="onFilterChange">
          <n-radio-button value="0">全部</n-radio-button>
          <n-radio-button value="1">未读</n-radio-button>
          <n-radio-button value="2">已读</n-radio-button>
        </n-radio-group>
        <n-select
          v-model:value="typeFilter"
          :options="typeOptions"
          size="small"
          clearable
          placeholder="全部类型"
          style="width: 140px"
          @update:value="onFilterChange"
        />
      </n-space>

      <n-spin :show="loading">
        <EmptyState type="notify" title="暂无通知" v-if="!loading && list.length === 0" description="任务指派、评论提及与到期提醒会送达这里" compact />
        <div v-else class="notice-list">
          <div
            v-for="item in list"
            :key="item.id"
            class="notice-row clickable"
            :class="{ unread: item.isRead === 0 }"
            @click="openSource(item)"
          >
            <div class="notice-main">
              <n-space :size="8" align="center">
                <n-tag size="small" :bordered="false" :type="typeTagType[item.type] || 'default'">
                  {{ typeLabelMap[item.type] || item.type }}
                </n-tag>
                <span class="notice-title" :class="{ bold: item.isRead === 0 }">{{ item.title }}</span>
                <n-tag v-if="item.sourceType" size="tiny" :bordered="false">
                  {{ sourceLabels[item.sourceType] || item.sourceType }}
                </n-tag>
              </n-space>
              <div class="notice-content">{{ item.content }}</div>
            </div>
            <n-space :size="8" align="center" :wrap="false">
              <span class="notice-time">{{ fmtTime(item.createdAt) }}</span>
              <n-button v-if="item.isRead === 0" text type="primary" size="tiny" @click="handleRead(item)">
                标为已读
              </n-button>
            </n-space>
          </div>
        </div>
      </n-spin>

      <div class="mt-4 flex justify-end" v-if="total > pagination.size">
        <n-pagination
          show-size-picker
          :page-sizes="[10, 20, 50, 100]"
          v-model:page="pagination.page"
          :page-size="pagination.size"
          :item-count="total"
          @update:page="onPageChange"
          @update:page-size="onPageSizeChange"
        />
      </div>
    </n-card>
  </div>
</template>

<script lang="ts" setup>
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { ref, reactive, onMounted } from 'vue';
  import { useMessage } from 'naive-ui';
  import { getNotifications, readNotification, readAllNotifications, getUnreadCount } from '@/api/platform/index';
  import { usePagedList } from '@/composables/usePagedList';
  import type { NotificationItem } from '@/api/platform/index';
  import { NOTICE_TYPE_LABELS, NOTICE_TYPE_TAG, NOTICE_SOURCE_LABELS, noticeTypeOptions } from '@/enums/notification';
  import { useRouter } from 'vue-router';

  const message = useMessage();
  const router = useRouter();
  const unread = ref(0);
  const readFilter = ref('0');
  const typeFilter = ref<string | null>(null);

  // 字典统一出口：enums/notification.ts（与 Header 铃铛共用）
  const typeOptions = noticeTypeOptions;
  const typeTagType = NOTICE_TYPE_TAG;
  const typeLabelMap = NOTICE_TYPE_LABELS;
  const sourceLabels = NOTICE_SOURCE_LABELS;

  function fmtTime(ts: string): string {
    if (!ts) return '';
    const d = ts.slice(0, 10);
    const hm = ts.slice(11, 16);
    const today = new Date();
    const t = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`;
    return d === t ? hm : `${d.slice(5)} ${hm}`;
  }

  // usePagedList 统一四件套（含筛选重置页码/末页删空回退；已读批量后刷新走 load）
  const {
    loading, list, total, pagination,
    load, onFilterChange, onPageChange, onPageSizeChange,
  } = usePagedList<any>((page: number, size: number) =>
    getNotifications({
      unread: Number(readFilter.value) || undefined,
      type: typeFilter.value || undefined,
      page,
      size,
    })
  );

  async function loadUnread() {
    try {
      const res: any = await getUnreadCount();
      unread.value = (res && res.count) || 0;
    } catch {
      // ignore
    }
  }



  // 行点击=已读+跳转源实体（与铃铛同款；"标为已读"按钮只读不跳）
  async function openSource(item: NotificationItem) {
    if (item.isRead === 0) await handleRead(item);
    try {
      if (item.sourceType === 'task' && item.sourceId) {
        // 动态 import：避免静态引 api 层进入 alova↔store 循环依赖
        const { getTask } = await import('@/api/project/index');
        const t = await getTask(item.sourceId);
        if (t?.projectId) router.push(`/project/${t.projectId}/tasks?task=${item.sourceId}`);
      } else if (item.sourceType === 'feedback' && item.sourceId) {
        // sourceId 是 projectId：反馈通知直达收件箱（原来跳概览，用户反馈导航断链）
        popoverShow.value = false;
        router.push(`/project/${item.sourceId}/feedbacks`);
      } else if (item.sourceType === 'project' && item.sourceId) {
        router.push(`/project/${item.sourceId}/overview`);
      }
    } catch {
      message.warning('无法打开来源（可能已无访问权限）');
    }
  }

  async function handleRead(item: NotificationItem) {
    try {
      await readNotification(item.id);
      item.isRead = 1;
      unread.value = Math.max(0, unread.value - 1);
    } catch {
      message.error('操作失败');
    }
  }

  async function handleReadAll() {
    try {
      await readAllNotifications();
      message.success('已全部标记为已读');
      unread.value = 0;
      load();
    } catch {
      message.error('操作失败');
    }
  }

  onMounted(() => {
    load();
    loadUnread();
  });
</script>

<style lang="less" scoped>
  .clickable {
    cursor: pointer;
  }

  .notice-list {
    display: flex;
    flex-direction: column;
  }

  .notice-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
    padding: 10px 12px;
    border-bottom: 1px solid #f0f0f0;

    &:last-child {
      border-bottom: none;
    }

    &:hover {
      background: #fafafa;
    }

    &.unread {
      border-left: 3px solid #18a058;
      background: rgb(24 160 88 / 3%);
    }
  }

  .notice-main {
    min-width: 0;
    flex: 1;
  }

  .notice-title {
    font-size: 14px;
    color: #333;

    &.bold {
      font-weight: 600;
    }
  }

  .notice-content {
    margin-top: 2px;
    font-size: 13px;
    color: #666;
    overflow: hidden;
    text-overflow: ellipsis;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
  }

  .notice-time {
    flex: none;
    font-size: 12px;
    color: #999;
  }
</style>
