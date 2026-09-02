<template>
  <div class="layout-root">
    <!-- 全宽 Header 浮卡：Logo + 面包屑 + 功能区，横贯整个宽度 -->
    <header class="layout-top-header">
      <PageHeader v-model:collapsed="collapsed" />
    </header>

    <!-- Header 下方：左侧菜单浮卡全高连续 + 内容区浮卡，三卡统一 8px 缝隙 -->
    <div class="layout-main">
      <n-layout-sider
        v-if="
          !isMobile && isMixMenuNoneSub && (navMode === 'vertical' || navMode === 'horizontal-mix')
        "
        show-trigger="bar"
        @collapse="collapsed = true"
        :collapsed="collapsed"
        @expand="collapsed = false"
        collapse-mode="width"
        :collapsed-width="64"
        :width="leftMenuWidth"
        :native-scrollbar="true"
        class="layout-sider"
      >
        <ProjectMenu v-if="inProjectContext" :collapsed="collapsed" />
        <AsideMenu v-else v-model:collapsed="collapsed" v-model:location="getMenuLocation" />
      </n-layout-sider>

      <n-drawer
        v-model:show="showSideDrawer"
        :width="menuWidth"
        :placement="'left'"
        class="layout-side-drawer"
      >
        <n-layout-sider
          :collapsed="false"
          :width="menuWidth"
          :native-scrollbar="true"
          class="layout-sider"
        >
          <Logo :collapsed="false" />
          <ProjectMenu v-if="inProjectContext" :collapsed="false" />
          <AsideMenu v-else v-model:location="getMenuLocation" />
        </n-layout-sider>
      </n-drawer>

      <!-- 内容区：唯一滚动容器，滚动条在卡片内侧，不挤压三卡对齐 -->
      <main class="layout-content" :class="{ 'layout-default-background': getDarkTheme === false }">
        <div class="layout-content-main">
          <MainView />
        </div>
      </main>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { ref, unref, computed, onMounted } from 'vue';
  import { Logo } from './components/Logo';
  import { MainView } from './components/Main';
  import { AsideMenu } from './components/Menu';
  import { PageHeader } from './components/Header';
  import ProjectMenu from './components/Menu/ProjectMenu.vue';
  import { useProjectSetting } from '@/hooks/setting/useProjectSetting';
  import { useDesignSetting } from '@/hooks/setting/useDesignSetting';
  import { useRoute } from 'vue-router';
  import { useProjectSettingStore } from '@/store/modules/projectSetting';

  const { getDarkTheme } = useDesignSetting();
  const { navMode, menuSetting } = useProjectSetting();

  const settingStore = useProjectSettingStore();

  const collapsed = ref<boolean>(false);
  const route = useRoute();
  // 项目工作台内左侧菜单整体替换为该项目专属导航。
  // 以路由参数同步判断（而非等 entityContext 的异步 API 返回），避免进入项目时全局菜单闪现
  const inProjectContext = computed(() => !!route.params.projectId);

  const { mobileWidth, menuWidth } = unref(menuSetting);

  const isMobile = computed<boolean>({
    get: () => settingStore.getIsMobile,
    set: (val) => settingStore.setIsMobile(val),
  });

  const isMixMenuNoneSub = computed(() => {
    const mixMenu = unref(menuSetting).mixMenu;
    const currentRoute = useRoute();
    if (unref(navMode) != 'horizontal-mix') return true;
    if (unref(navMode) === 'horizontal-mix' && mixMenu && currentRoute.meta.isRoot) {
      return false;
    }
    return true;
  });

  const leftMenuWidth = computed(() => {
    const { minMenuWidth, menuWidth } = unref(menuSetting);
    return collapsed.value ? minMenuWidth : menuWidth;
  });

  const showSideDrawer = computed({
    get: () => isMobile.value && collapsed.value,
    set: (val) => (collapsed.value = val),
  });

  const checkMobileMode = () => {
    if (document.body.clientWidth <= mobileWidth) {
      isMobile.value = true;
    } else {
      isMobile.value = false;
    }
    collapsed.value = false;
  };

  const watchWidth = () => {
    const Width = document.body.clientWidth;
    if (Width <= 950) {
      collapsed.value = true;
    } else collapsed.value = false;

    checkMobileMode();
  };

  onMounted(() => {
    checkMobileMode();
    window.addEventListener('resize', watchWidth);
  });
</script>

<style lang="less">
  .layout-side-drawer {
    background-color: rgb(0, 20, 40);

    .layout-sider {
      min-height: 100vh;
      box-shadow: 2px 0 8px 0 rgb(29 35 41 / 5%);
      position: relative;
      z-index: 13;
      transition: all 0.2s ease-in-out;
    }
  }
</style>
<style lang="less" scoped>
  // 统一浮卡几何：画布上 8px 缝隙、12px 圆角、面板阴影。
  // 视口锁高：页面级永不滚动（消灭全宽滚动条挤压三卡对齐），滚动只在内容卡内部
  .layout-root {
    height: 100vh;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    background: var(--canvas, #f1f1ee);

    .layout-top-header {
      flex-shrink: 0;
      height: 64px;
      margin: 8px 8px 0;
      border-radius: var(--panel-radius, 12px);
      background: var(--panel-bg, #fff);
      box-shadow: var(--panel-shadow);
    }

    .layout-main {
      flex: auto;
      min-height: 0;
      display: flex;
      flex-direction: row;
    }

    .layout-sider {
      flex-shrink: 0;
      transition: all 0.2s ease-in-out;
      // 悬浮栏卡片：左右各留 8px 缝（右侧缝即与内容区的间隔）
      margin: 8px 0 8px 8px;
      height: calc(100% - 16px);
      align-self: flex-start;
      border-radius: var(--panel-radius, 12px);
      background: var(--panel-bg, #fff);
      box-shadow: var(--panel-shadow);

      // n-layout-sider 独立使用时其内部滚动容器需要显式圆角裁切
      :deep(.n-layout-sider-scroll-container) {
        border-radius: inherit;
      }
    }

    .layout-content {
      flex: auto;
      min-width: 0;
      margin: 8px 8px 8px 0;
      border-radius: var(--panel-radius, 12px);
      overflow-y: auto;
      overflow-x: hidden;
    }

    &-default-background {
      background: var(--canvas, #f1f1ee);
    }

    .layout-content-main {
      padding: 0 8px 8px;
    }
  }
</style>
