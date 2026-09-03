<template>
  <div class="layout-root">
    <!-- 全宽 Header 浮卡：Logo + 面包屑 + 功能区，横贯整个宽度 -->
    <header class="layout-top-header">
      <PageHeader v-model:collapsed="collapsed" />
    </header>

    <!-- Header 下方：左侧菜单浮卡全高连续 + 内容区浮卡，三卡统一 8px 缝隙 -->
    <div class="layout-main">
      <!-- 自控侧栏（脱离 n-layout 独立渲染后 naive 的宽度折叠响应不可靠，改为纯 CSS transition） -->
      <aside
        v-if="
          !isMobile && isMixMenuNoneSub && (navMode === 'vertical' || navMode === 'horizontal-mix')
        "
        class="layout-sider"
        :class="{ collapsed }"
        :style="{ width: collapsed ? collapsedMenuWidth : menuWidth + 'px' }"
      >
        <div class="sider-menu">
          <ProjectMenu v-if="inProjectContext" :collapsed="collapsed" />
          <AsideMenu v-else v-model:collapsed="collapsed" v-model:location="getMenuLocation" />
        </div>
        <!-- 折叠控制放在菜单卡自身底部：控制点与被控对象一体，菜单滚动时按钮固定不动 -->
        <div class="sider-footer" :class="{ collapsed }" @click="collapsed = !collapsed">
          <n-icon size="15">
            <MenuFoldOutlined v-if="!collapsed" />
            <MenuUnfoldOutlined v-else />
          </n-icon>
          <span v-show="!collapsed" class="sider-footer-text">收起菜单</span>
        </div>
      </aside>

      <n-drawer
        v-model:show="showSideDrawer"
        :width="menuWidth"
        :placement="'left'"
        class="layout-side-drawer"
      >
        <div class="drawer-sider">
          <Logo :collapsed="false" />
          <ProjectMenu v-if="inProjectContext" :collapsed="false" />
          <AsideMenu v-else v-model:location="getMenuLocation" />
        </div>
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
  import { MenuFoldOutlined, MenuUnfoldOutlined } from '@vicons/antd';
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
  // 收起态卡宽（icon-only）
  const collapsedMenuWidth = '64px';
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

    .drawer-sider {
      min-height: 100vh;
      box-shadow: 2px 0 8px 0 rgb(29 35 41 / 5%);
      position: relative;
      z-index: 13;
    }
  }
</style>
<style lang="less" scoped>
  // 通栏 Header（贯穿、无圆角）+ 下方浮卡；画布 8px 缝隙、12px 圆角统一。
  // 视口锁高：页面级永不滚动（消灭全宽滚动条挤压对齐），滚动只在内容卡内部
  .layout-root {
    height: 100vh;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    background: var(--canvas, #f1f1ee);

    .layout-top-header {
      flex-shrink: 0;
      height: 64px;
      background: var(--panel-bg, #fff);
      border-bottom: 1px solid var(--line, #e9e9e7);
    }

    .layout-main {
      flex: auto;
      min-height: 0;
      display: flex;
      flex-direction: row;
    }

    .layout-sider {
      flex-shrink: 0;
      display: flex;
      flex-direction: column;
      align-self: stretch;
      // 悬浮栏卡片：左右各留 8px 缝（右侧缝即与内容区的间隔）
      margin: 8px 0 8px 8px;
      border-radius: var(--panel-radius, 12px);
      background: var(--panel-bg, #fff);
      box-shadow: var(--panel-shadow);
      overflow: hidden;
      transition: width 0.2s ease-in-out;

      .sider-menu {
        flex: 1;
        min-height: 0;
        overflow-y: auto;
      }

      .sider-footer {
        flex-shrink: 0;
        display: flex;
        align-items: center;
        gap: 8px;
        height: 40px;
        padding: 0 18px;
        cursor: pointer;
        border-top: 1px solid var(--line, #e9e9e7);
        color: var(--text-2, #57606a);
        font-size: 13px;

        &:hover {
          color: #16a34a;
          background: var(--hover-bg);
        }

        // 收起态（64px 卡宽）：仅图标居中
        &.collapsed {
          justify-content: center;
          padding: 0;
        }
      }
    }

    .layout-content {
      flex: auto;
      min-width: 0;
      margin: 8px 8px 8px 0;
      border-radius: var(--panel-radius, 12px);
      overflow-y: auto;
      overflow-x: hidden;
      // flex 列：铺满型页面（看板/vault）子项 flex:1 恰好撑满不触发本层滚动，
      // 滚动只发生在卡内；普通页面内容超高仍在本层滚
      display: flex;
      flex-direction: column;
    }

    &-default-background {
      background: var(--canvas, #f1f1ee);
    }

    .layout-content-main {
      // 顶部 4px 统一呼吸距（页面级标题卡已移除，标题由 Header 面包屑承载）；
      // flex 列 + min-height:0 让看板等整页型视图可 flex:1 铺满视口；
      // 普通页面内容超高时由父级 main（overflow:auto）承接滚动
      padding: 4px 8px 8px;
      flex: 1;
      min-height: 0;
      display: flex;
      flex-direction: column;

      // 页面根节点：宽度撑满；末位节点至少撑满剩余高度——
      // 表格型页面数据少时不留大片画布空白，数据多时自然长高照常滚动
      > * {
        width: 100%;

        &:last-child {
          flex: 1;
          min-height: 0;
        }
      }
    }
  }
</style>
