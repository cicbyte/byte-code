<template>
  <div class="entity-nav-bar" v-if="entityContext.currentEntityType">
    <div class="entity-nav-bar-inner">
      <span class="entity-name">{{ entityContext.currentEntityName }}</span>
      <n-divider vertical />
      <n-space :size="4">
        <router-link
          v-for="tab in tabs"
          :key="tab.key"
          :to="tab.path"
          custom
          v-slot="{ isActive, navigate }"
        >
          <div
            class="nav-tab"
            :class="{ active: isActive }"
            @click="navigate"
          >
            {{ tab.label }}
          </div>
        </router-link>
      </n-space>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { useEntityContext } from '@/store/modules/entityContext';
  import { getEntityTabs } from '@/config/entityNavConfig';

  const entityContext = useEntityContext();

  const tabs = computed(() => {
    const type = entityContext.currentEntityType;
    const id = entityContext.currentProject?.id;
    if (!type || !id) return [];
    return getEntityTabs(type, id);
  });
</script>

<style lang="less" scoped>
  .entity-nav-bar {
    background: #fff;
    border-bottom: 1px solid #e8e8e8;
    box-shadow: 0 1px 4px rgb(0 21 41 / 6%);
    position: sticky;
    top: 72px;
    z-index: 10;

    .entity-nav-bar-inner {
      display: flex;
      align-items: center;
      height: 40px;
      padding: 0 16px;
    }

    .entity-name {
      font-size: 14px;
      font-weight: 600;
      color: #333;
      white-space: nowrap;
    }

    .nav-tab {
      display: inline-flex;
      align-items: center;
      height: 40px;
      padding: 0 12px;
      font-size: 13px;
      color: #666;
      cursor: pointer;
      transition: color 0.2s;
      text-decoration: none;
      border-bottom: 2px solid transparent;
      margin-bottom: -1px;

      &:hover {
        color: #18a058;
      }

      &.active {
        color: #18a058;
        font-weight: 500;
        border-bottom-color: #18a058;
      }
    }
  }
</style>
