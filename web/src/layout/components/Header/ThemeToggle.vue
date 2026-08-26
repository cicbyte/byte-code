<template>
  <div class="layout-header-trigger layout-header-trigger-min" @click="toggle">
    <n-tooltip placement="bottom">
      <template #trigger>
        <n-icon size="18">
          <!-- Lucide 风格太阳/月亮 -->
          <svg
            v-if="!isDark"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <circle cx="12" cy="12" r="4" />
            <path
              d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41"
            />
          </svg>
          <svg
            v-else
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
          </svg>
        </n-icon>
      </template>
      <span>{{ isDark ? '切换浅色模式' : '切换暗色模式' }}</span>
    </n-tooltip>
  </div>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  // 注意用 store 实例而非 @/hooks/setting/useDesignSetting——
  // hooks 版只返回 getter 包装对象，不含 actions
  import { useDesignSettingStore } from '@/store/modules/designSetting';

  const designSetting = useDesignSettingStore();
  const isDark = computed(() => designSetting.darkTheme);

  function toggle() {
    designSetting.toggleDarkTheme();
  }
</script>
