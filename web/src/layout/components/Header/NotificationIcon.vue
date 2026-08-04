<template>
  <n-popover trigger="click" placement="bottom" :width="340" @update:show="onShowChange">
    <template #trigger>
      <div class="layout-header-trigger layout-header-trigger-min">
        <n-badge :value="unreadCount" :max="99">
          <n-icon size="18">
            <BellOutlined />
          </n-icon>
        </n-badge>
      </div>
    </template>
    <div class="notif-panel">
      <div class="notif-header">
        <span class="notif-title-bar">通知</span>
        <n-button text size="tiny" type="primary" :disabled="unreadCount === 0" @click="markAllRead">
          全部已读
        </n-button>
      </div>
      <n-spin :show="loading" size="small">
        <div v-if="list.length === 0" class="notif-empty">暂无通知</div>
        <div
          v-for="item in list"
          :key="item.id"
          class="notif-item"
          :class="{ unread: item.isRead === 0 }"
          @click="markRead(item)"
        >
          <div class="notif-item-title">
            <n-tag size="tiny" :type="tagType(item.type)">{{ typeLabel(item.type) }}</n-tag>
            <span class="notif-item-name">{{ item.title }}</span>
            <span v-if="item.isRead === 0" class="notif-dot"></span>
          </div>
          <div class="notif-item-content">{{ item.content }}</div>
          <div class="notif-item-time">{{ (item.createdAt || '').slice(0, 16) }}</div>
        </div>
      </n-spin>
    </div>
  </n-popover>
</template>

<script lang="ts" setup>
  import { onMounted, onUnmounted, ref } from 'vue';
  import { BellOutlined } from '@vicons/antd';
  import {
    getNotifications,
    getUnreadCount,
    NotificationItem,
    readAllNotifications,
    readNotification,
  } from '@/api/platform';

  const POLL_INTERVAL = 60 * 1000;

  const unreadCount = ref(0);
  const list = ref<NotificationItem[]>([]);
  const total = ref(0);
  const loading = ref(false);
  let timer: ReturnType<typeof setInterval> | null = null;

  async function fetchUnread() {
    try {
      const res = await getUnreadCount();
      unreadCount.value = res?.count || 0;
    } catch {
      // 未登录或会话过期由全局拦截器处理，这里静默即可
    }
  }

  async function loadList() {
    loading.value = true;
    try {
      const res = await getNotifications({ page: 1, size: 10 });
      list.value = res?.list || [];
      total.value = res?.total || 0;
    } catch {
      // 同上，静默
    } finally {
      loading.value = false;
    }
  }

  function onShowChange(show: boolean) {
    if (show) {
      loadList();
      fetchUnread();
    }
  }

  async function markRead(item: NotificationItem) {
    if (item.isRead !== 0) return;
    try {
      await readNotification(item.id);
      item.isRead = 1;
      fetchUnread();
    } catch {
      // 忽略单条已读失败
    }
  }

  async function markAllRead() {
    try {
      await readAllNotifications();
      await Promise.all([loadList(), fetchUnread()]);
    } catch {
      // 忽略
    }
  }

  function tagType(t: string): 'default' | 'error' | 'info' | 'success' | 'warning' {
    switch (t) {
      case 'warning':
        return 'warning';
      case 'success':
        return 'success';
      case 'error':
        return 'error';
      default:
        return 'info';
    }
  }

  function typeLabel(t: string): string {
    const m: Record<string, string> = { info: '通知', warning: '警告', success: '成功', error: '错误' };
    return m[t] || '通知';
  }

  onMounted(() => {
    fetchUnread();
    timer = setInterval(fetchUnread, POLL_INTERVAL);
  });

  onUnmounted(() => {
    if (timer) clearInterval(timer);
  });
</script>

<style lang="less" scoped>
  .notif-panel {
    margin: -4px -8px;
  }

  .notif-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 4px 12px 8px;
    border-bottom: 1px solid var(--n-border-color, #efeff5);
  }

  .notif-title-bar {
    font-weight: 500;
  }

  .notif-empty {
    padding: 24px 0;
    color: #999;
    text-align: center;
  }

  .notif-item {
    padding: 8px 12px;
    cursor: pointer;
    border-bottom: 1px solid var(--n-border-color, #efeff5);

    &:last-child {
      border-bottom: none;
    }

    &:hover {
      background: rgba(0, 0, 0, 0.02);
    }

    &.unread .notif-item-name {
      font-weight: 500;
    }
  }

  .notif-item-title {
    display: flex;
    gap: 6px;
    align-items: center;
  }

  .notif-item-name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .notif-dot {
    flex: none;
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: #f00;
  }

  .notif-item-content {
    margin-top: 4px;
    overflow: hidden;
    color: #666;
    font-size: 12px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .notif-item-time {
    margin-top: 2px;
    color: #999;
    font-size: 12px;
  }
</style>
