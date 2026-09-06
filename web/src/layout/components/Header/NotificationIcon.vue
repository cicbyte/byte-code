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
        <EmptyState v-if="list.length === 0" type="notify" title="暂无通知" description="任务指派与评论提醒会送达这里" compact />
        <div
          v-for="item in list"
          :key="item.id"
          class="notif-item"
          :class="{ unread: item.isRead === 0 }"
          @click="markRead(item)"
        >
          <div class="notif-item-title">
            <n-tag size="tiny" :type="noticeTypeTagType(item.type)">{{ noticeTypeLabel(item.type) }}</n-tag>
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
  import { useRouter } from 'vue-router';
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { noticeTypeLabel, noticeTypeTagType } from '@/enums/notification';
  import { storage } from '@/utils/Storage';
  import { ACCESS_TOKEN } from '@/store/mutation-types';
  import { BellOutlined } from '@vicons/antd';
  import {
    getNotifications,
    getUnreadCount,
    NotificationItem,
    readAllNotifications,
    readNotification,
  } from '@/api/platform';

  const router = useRouter();

  // 轮询降级兜底：SSE 断开时仍能拉到通知（间隔放长到 5 分钟）
  const POLL_INTERVAL = 5 * 60 * 1000;

  const unreadCount = ref(0);
  const popoverShow = ref(false);
  const list = ref<NotificationItem[]>([]);
  const total = ref(0);
  const loading = ref(false);
  let timer: ReturnType<typeof setInterval> | null = null;

  // ==================== SSE 实时通知（fetch 流解析；EventSource 不支持自定义 header） ====================
  let sseAbort: AbortController | null = null;

  function startSSE() {
    sseAbort = new AbortController();
    // 走 Storage 封装读 token（带 expire 校验，key 不再侥幸匹配裸 localStorage）。
    // 勿 import user store 取 token：NotificationIcon→store/user→api→alova→store/user
    // 成环，这条边会改变模块初始化顺序导致页面数据加载静默失效（实测踩坑）
    const token = (storage.get(ACCESS_TOKEN, '') as string) || '';
    fetch('/api/v1/notifications/stream', {
      headers: { token },
      signal: sseAbort.signal,
    })
      .then(async (resp) => {
        if (!resp.ok || !resp.body) throw new Error('sse unavailable');
        const reader = resp.body.getReader();
        const decoder = new TextDecoder();
        let buf = '';
        for (;;) {
          const { done, value } = await reader.read();
          if (done) break;
          buf += decoder.decode(value, { stream: true });
          // SSE 事件以空行分隔；注释行（: ping）跳过
          let idx;
          while ((idx = buf.indexOf('\n\n')) >= 0) {
            const frame = buf.slice(0, idx);
            buf = buf.slice(idx + 2);
            const dataLine = frame.split('\n').find((l) => l.startsWith('data: '));
            if (!dataLine) continue;
            // 收到实时通知：刷未读数；面板打开时同步列表；同时广播为
            // window 事件——任务详情抽屉据此对正打开的任务做评论区实时刷新
            let payload: any = null;
            try {
              payload = JSON.parse(dataLine.slice(6));
            } catch {
              payload = null;
            }
            fetchUnread();
            if (popoverShow.value) loadList();
            window.dispatchEvent(new CustomEvent('bc-notification', { detail: payload }));
          }
        }
      })
      .catch(() => {
        // 连接断开/失败：什么都不做，轮询兜底仍在
      });
  }

  function stopSSE() {
    sseAbort?.abort();
    sseAbort = null;
  }

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
    popoverShow.value = show;
    if (show) {
      loadList();
      fetchUnread();
    }
  }

  async function markRead(item: NotificationItem) {
    // 点击=已读+跳转源实体（已读失败不阻断跳转）
    if (item.isRead === 0) {
      try {
        await readNotification(item.id);
        item.isRead = 1;
        fetchUnread();
      } catch {
        // 忽略单条已读失败
      }
    }
    jumpToSource(item);
  }

  // 跳转源实体：task → 任务列表并自动开抽屉（?task= 由宿主页消费）；
  // project → 项目概览。getTask 用动态 import——layout 组件静态引 api 层
  // 会进 alova↔store 循环依赖，破坏全站模块初始化（踩坑两次，勿改回）
  async function jumpToSource(item: NotificationItem) {
    try {
      if (item.sourceType === 'task' && item.sourceId) {
        const { getTask } = await import('@/api/project/index');
        const t = await getTask(item.sourceId);
        if (t?.projectId) {
          popoverShow.value = false;
          router.push(`/project/${t.projectId}/tasks?task=${item.sourceId}`);
        }
      } else if (item.sourceType === 'project' && item.sourceId) {
        popoverShow.value = false;
        router.push(`/project/${item.sourceId}/overview`);
      }
    } catch {
      // 无权访问或实体已删：停留原地
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


  onMounted(() => {
    fetchUnread();
    startSSE();
    timer = setInterval(fetchUnread, POLL_INTERVAL);
  });

  onUnmounted(() => {
    if (timer) clearInterval(timer);
    stopSSE();
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
    color: var(--text-3, #8b949e);
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
      background: var(--hover-bg);
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
    background: #d03050;
  }

  .notif-item-content {
    margin-top: 4px;
    overflow: hidden;
    color: var(--text-2, #57606a);
    font-size: 12px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .notif-item-time {
    margin-top: 2px;
    color: var(--text-3, #8b949e);
    font-size: 12px;
  }
</style>
