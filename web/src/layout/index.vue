<template>
  <n-layout class="layout" :position="fixedMenu" has-sider>
    <n-layout-sider
      v-if="
        !isMobile && isMixMenuNoneSub && (navMode === 'vertical' || navMode === 'horizontal-mix')
      "
      show-trigger="bar"
      @collapse="collapsed = true"
      :position="fixedMenu"
      @expand="collapsed = false"
      :collapsed="collapsed"
      collapse-mode="width"
      :collapsed-width="64"
      :width="leftMenuWidth"
      :native-scrollbar="false"
      :inverted="false"
      class="layout-sider"
    >
      <Logo :collapsed="collapsed" />
      <AsideMenu v-model:collapsed="collapsed" v-model:location="getMenuLocation" />
    </n-layout-sider>

    <n-drawer
      v-model:show="showSideDrawer"
      :width="menuWidth"
      :placement="'left'"
      class="layout-side-drawer"
    >
      <n-layout-sider
        :position="fixedMenu"
        :collapsed="false"
        :width="menuWidth"
        :native-scrollbar="false"
        :inverted="false"
        class="layout-sider"
      >
        <Logo :collapsed="collapsed" />
        <AsideMenu v-model:location="getMenuLocation" />
      </n-layout-sider>
    </n-drawer>

    <n-layout :inverted="inverted">
      <n-layout-header :inverted="false" :position="fixedHeader">
        <PageHeader v-model:collapsed="collapsed" :inverted="false" />
      </n-layout-header>

      <n-layout-content
        class="layout-content"
        :class="{ 'layout-default-background': getDarkTheme === false }"
      >
        <Entity-nav-bar />
        <div
          class="layout-content-main"
          :class="{
            'fluid-header': fixedHeader === 'static',
          }"
        >
          <div class="main-view mt-3">
            <MainView />
          </div>
        </div>
      </n-layout-content>
      <n-back-top :right="100" />
    </n-layout>
  </n-layout>
</template>

<script lang="ts" setup>
  import { ref, unref, computed, onMounted } from 'vue';
  import { Logo } from './components/Logo';
  import { MainView } from './components/Main';
  import { AsideMenu } from './components/Menu';
  import { PageHeader } from './components/Header';
  import EntityNavBar from './components/EntityNavBar/index.vue';
  import { useProjectSetting } from '@/hooks/setting/useProjectSetting';
  import { useDesignSetting } from '@/hooks/setting/useDesignSetting';
  import { useRoute } from 'vue-router';
  import { useProjectSettingStore } from '@/store/modules/projectSetting';

  const { getDarkTheme } = useDesignSetting();
  const {
    navMode,
    navTheme,
    headerSetting,
    menuSetting,
  } = useProjectSetting();

  const settingStore = useProjectSettingStore();

  const collapsed = ref<boolean>(false);

  const { mobileWidth, menuWidth } = unref(menuSetting);

  const isMobile = computed<boolean>({
    get: () => settingStore.getIsMobile,
    set: (val) => settingStore.setIsMobile(val),
  });

  const fixedHeader = computed(() => {
    const { fixed } = unref(headerSetting);
    return fixed ? 'absolute' : 'static';
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

  const fixedMenu = computed(() => {
    const { fixed } = unref(headerSetting);
    return fixed ? 'absolute' : 'static';
  });

  const inverted = computed(() => {
    return ['dark', 'header-dark'].includes(unref(navTheme));
  });

  const getHeaderInverted = computed(() => {
    return ['light', 'header-dark'].includes(unref(navTheme)) ? unref(inverted) : !unref(inverted);
  });

  const leftMenuWidth = computed(() => {
    const { minMenuWidth, menuWidth } = unref(menuSetting);
    return collapsed.value ? minMenuWidth : menuWidth;
  });

  const getMenuLocation = computed(() => {
    return 'left';
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
  .layout {
    display: flex;
    flex-direction: row;
    flex: auto;

    &-default-background {
      background: var(--canvas, #f1f1ee);
    }

    .layout-sider {
      position: relative;
      z-index: 13;
      transition: all 0.2s ease-in-out;
      // 悬浮栏卡片：白底圆角浮于画布之上，与内容区留 8px 细缝
      // 高度扣除上下边距，避免底边被视口裁切
      margin: 8px 0 8px 8px;
      min-height: calc(100vh - 16px);
      border-radius: var(--panel-radius, 12px);
      background: #fff;
      box-shadow: var(--panel-shadow);
    }

    .layout-sider-fix {
      position: fixed;
      top: 0;
      left: 0;
    }

    .ant-layout {
      overflow: hidden;
    }

    .layout-right-fix {
      overflow-x: hidden;
      padding-left: 200px;
      min-height: 100vh;
      transition: all 0.2s ease-in-out;
    }

    .layout-content {
      flex: auto;
      min-height: 100vh;
    }

    .n-layout-header.n-layout-header--absolute-positioned {
      z-index: 11;
    }

    .n-layout-footer {
      background: none;
    }
  }

  .layout-content-main {
    margin: 0 16px 16px;
    position: relative;
    // header 浮卡带 8px 上边距，内容需让出对应高度
    padding-top: 72px;
  }

  .fluid-header {
    padding-top: 0;
  }
</style>
