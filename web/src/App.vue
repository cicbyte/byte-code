<template>
  <NConfigProvider
    :locale="zhCN"
    :theme="getDarkTheme"
    :theme-overrides="getThemeOverrides"
    :date-locale="dateZhCN"
  >
    <AppProvider>
      <RouterView />
    </AppProvider>
  </NConfigProvider>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { zhCN, dateZhCN, darkTheme } from 'naive-ui';
  import { AppProvider } from '@/components/Application';
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
        bodyColor: '#f1f1ee',
        cardColor: '#ffffff',
        borderColor: '#e9e9e7',
        dividerColor: '#e9e9e7',
        textColorBase: '#1c1d21',
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

  const getDarkTheme = computed(() => (designStore.darkTheme ? darkTheme : undefined));
</script>
