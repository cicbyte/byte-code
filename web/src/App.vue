<template>
  <NConfigProvider
    :locale="zhCN"
    :theme="getDarkTheme"
    :theme-overrides="getThemeOverrides"
    :date-locale="dateZhCN"
  >
    <AppProvider>
      <RouterView />
      <ForceChangePassword />
    </AppProvider>
  </NConfigProvider>
</template>

<script lang="ts" setup>
  import { computed, watch } from 'vue';
  import { zhCN, dateZhCN, darkTheme } from 'naive-ui';
  import { AppProvider } from '@/components/Application';
  import ForceChangePassword from '@/components/ForceChangePassword/index.vue';
  import { useDesignSettingStore } from '@/store/modules/designSetting';
  import { lighten } from '@/utils/index';

  const designStore = useDesignSettingStore();

  /**
   * @type import('naive-ui').GlobalThemeOverrides
   */
  // 视觉基线参考 dev-docs/prototype（HiFox 风格：品牌绿 + 13px 密度 + 8/12px 圆角）
  const getThemeOverrides = computed(() => {
    const appTheme = designStore.appTheme;
    const lightenStr = lighten(designStore.appTheme, 6);
    // 浅色/暗色两套表面令牌，与 styles/index.less 的 CSS 变量保持一致
    const surface = designStore.darkTheme
      ? { bodyColor: '#1b1c1f', cardColor: '#232428', borderColor: '#333438', dividerColor: '#333438' }
      : { bodyColor: '#f1f1ee', cardColor: '#ffffff', borderColor: '#e9e9e7', dividerColor: '#e9e9e7' };
    return {
      common: {
        primaryColor: '#16a34a',
        primaryColorHover: '#18b551',
        primaryColorPressed: '#128a3f',
        primaryColorSuppl: '#16a34a',
        borderRadius: '8px',
        borderRadiusSmall: '6px',
        fontSize: '13px',
        fontSizeMedium: '13px',
        fontSizeSmall: '12px',
        ...surface,
      },
      Card: {
        borderRadius: '12px',
      },
      Button: {
        borderRadiusMedium: '8px',
      },
      LoadingBar: {
        colorLoading: appTheme,
      },
    };
  });

  // 同步 DOM 主题属性，驱动全局 CSS 变量（画布/浮卡/滚动条）在明暗间切换
  const applyDomTheme = () => {
    document.documentElement.setAttribute(
      'data-theme',
      designStore.darkTheme ? 'dark' : 'light'
    );
  };
  watch(
    () => designStore.darkTheme,
    () => applyDomTheme(),
    { immediate: true }
  );

  const getDarkTheme = computed(() => (designStore.darkTheme ? darkTheme : undefined));
</script>
