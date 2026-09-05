<template>
  <div class="empty-state" :class="{ compact }">
    <div class="empty-visual">
      <!-- 背景装饰：淡绿圆斑 + 点阵 -->
      <span class="empty-blob" />
      <span class="empty-dots" />
      <svg
        class="empty-icon"
        viewBox="0 0 64 64"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        aria-hidden="true"
      >
        <!-- 各场景描边插画（1.5px 描边 + 品牌绿点缀） -->
        <template v-if="type === 'task'">
          <rect x="14" y="10" width="36" height="44" rx="5" stroke="currentColor" stroke-width="1.5" />
          <rect x="25" y="6" width="14" height="8" rx="3" stroke="currentColor" stroke-width="1.5" />
          <path d="M22 26h14M22 34h20M22 42h10" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
          <path d="M43 40l3 3 5-6" stroke="#18a058" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
        </template>
        <template v-else-if="type === 'doc'">
          <path d="M16 8h20l12 12v36a4 4 0 0 1-4 4H16a4 4 0 0 1-4-4V12a4 4 0 0 1 4-4z" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round" />
          <path d="M36 8v12h12" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round" />
          <path d="M21 30h16M21 38h22M21 46h12" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
          <path d="M44 24l1.5 4 4 1.5-4 1.5-1.5 4-1.5-4-4-1.5 4-1.5z" fill="#18a058" />
        </template>
        <template v-else-if="type === 'comment'">
          <path d="M12 14a6 6 0 0 1 6-6h28a6 6 0 0 1 6 6v20a6 6 0 0 1-6 6H30l-10 10V40h-2a6 6 0 0 1-6-6z" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round" />
          <circle cx="23" cy="24" r="2.4" fill="#18a058" />
          <circle cx="32" cy="24" r="2.4" fill="#18a058" opacity=".7" />
          <circle cx="41" cy="24" r="2.4" fill="#18a058" opacity=".4" />
        </template>
        <template v-else-if="type === 'notify'">
          <path d="M32 10a14 14 0 0 1 14 14v10l4 8H14l4-8V24a14 14 0 0 1 14-14z" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round" />
          <path d="M26 46a6 6 0 0 0 12 0" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
          <path d="M50 14l1.2 3.2 3.2 1.2-3.2 1.2L50 22.8l-1.2-3.2-3.2-1.2 3.2-1.2z" fill="#18a058" />
        </template>
        <template v-else-if="type === 'member'">
          <circle cx="24" cy="22" r="9" stroke="currentColor" stroke-width="1.5" />
          <path d="M8 50c0-9 7-15 16-15s16 6 16 15" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
          <circle cx="45" cy="24" r="6.5" stroke="#18a058" stroke-width="1.5" />
          <path d="M38 46c.5-6.5 5-10 11-10 4 0 8 2 9 8" stroke="#18a058" stroke-width="1.5" stroke-linecap="round" opacity=".6" />
        </template>
        <template v-else-if="type === 'data'">
          <path d="M10 50V22M10 50h44" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
          <rect x="18" y="34" width="7" height="16" rx="1.5" stroke="currentColor" stroke-width="1.5" />
          <rect x="31" y="26" width="7" height="24" rx="1.5" stroke="#18a058" stroke-width="1.5" />
          <rect x="44" y="16" width="7" height="34" rx="1.5" stroke="currentColor" stroke-width="1.5" />
        </template>
        <template v-else-if="type === 'search'">
          <circle cx="28" cy="28" r="15" stroke="currentColor" stroke-width="1.5" />
          <path d="M39 39l12 12" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
          <path d="M22 28a6 6 0 0 1 6-6" stroke="#18a058" stroke-width="2" stroke-linecap="round" />
        </template>
        <template v-else>
          <!-- generic：虚线圆 + sparkle -->
          <circle cx="32" cy="32" r="20" stroke="currentColor" stroke-width="1.5" stroke-dasharray="4 5" stroke-linecap="round" />
          <path d="M32 20c.9 4.5 3.5 7.1 8 8-4.5.9-7.1 3.5-8 8-.9-4.5-3.5-7.1-8-8 4.5-.9 7.1-3.5 8-8z" fill="#18a058" />
          <path d="M44 40c.4 2 1.6 3.2 3.6 3.6-2 .4-3.2 1.6-3.6 3.6-.4-2-1.6-3.2-3.6-3.6 2-.4 3.2-1.6 3.6-3.6z" fill="#18a058" opacity=".55" />
        </template>
      </svg>
    </div>
    <div class="empty-title">{{ title }}</div>
    <div v-if="description" class="empty-desc">{{ description }}</div>
    <n-button v-if="actionText" size="small" type="primary" ghost class="empty-action" @click="$emit('action')">
      {{ actionText }}
    </n-button>
  </div>
</template>

<script lang="ts" setup>
  withDefaults(
    defineProps<{
      /** 场景插画：task/doc/comment/notify/member/data/search/generic */
      type?: string;
      title: string;
      description?: string;
      actionText?: string;
      /** 紧凑模式：卡片内小空态（缩小插画与留白） */
      compact?: boolean;
    }>(),
    { type: 'generic', description: '', actionText: '', compact: false }
  );

  defineEmits<{ (e: 'action'): void }>();
</script>

<style lang="less" scoped>
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 32px 16px;
    text-align: center;
    color: #b8c0cc; // 插画描边主色（浅灰蓝）

    &.compact {
      padding: 18px 12px;

      .empty-visual {
        width: 52px;
        height: 52px;
      }

      .empty-title {
        font-size: 13px;
        margin-top: 8px;
      }

      .empty-desc {
        font-size: 12px;
      }
    }
  }

  .empty-visual {
    position: relative;
    width: 68px;
    height: 68px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .empty-icon {
    width: 100%;
    height: 100%;
    position: relative;
    z-index: 1;
  }

  // 背景装饰：右下淡绿圆斑
  .empty-blob {
    position: absolute;
    right: -4px;
    bottom: -6px;
    width: 34px;
    height: 34px;
    border-radius: 50%;
    background: radial-gradient(circle, rgb(24 160 88 / 12%) 0%, transparent 70%);
  }

  // 左上点阵
  .empty-dots {
    position: absolute;
    left: -8px;
    top: -4px;
    width: 22px;
    height: 22px;
    background-image: radial-gradient(#c8d0da 1.2px, transparent 1.2px);
    background-size: 6px 6px;
    opacity: 0.8;
  }

  .empty-title {
    margin-top: 12px;
    font-size: 14px;
    font-weight: 500;
    color: #4b5563;
  }

  .empty-desc {
    margin-top: 4px;
    font-size: 12.5px;
    color: #9ca3af;
    max-width: 320px;
    line-height: 1.6;
  }

  .empty-action {
    margin-top: 14px;
  }
</style>
