import { computed } from 'vue';
import { useProjectSettingStore } from '@/store/modules/projectSetting';

export function useProjectSetting() {
  const projectStore = useProjectSettingStore();

  const navMode = computed(() => projectStore.navMode);

  const navTheme = computed(() => projectStore.navTheme);

  const isMobile = computed(() => projectStore.isMobile);

  const headerSetting = computed(() => projectStore.headerSetting);

  const menuSetting = computed(() => projectStore.menuSetting);

  const crumbsSetting = computed(() => projectStore.crumbsSetting);


  const isPageAnimate = computed(() => projectStore.isPageAnimate);

  const pageAnimateType = computed(() => projectStore.pageAnimateType);

  return {
    navMode,
    navTheme,
    isMobile,
    headerSetting,
    menuSetting,
    crumbsSetting,
    isPageAnimate,
    pageAnimateType,
  };
}
