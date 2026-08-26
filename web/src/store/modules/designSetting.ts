import { defineStore } from 'pinia';
import { store } from '@/store';
import designSetting from '@/settings/designSetting';

const { appTheme, appThemeList } = designSetting;

// 暗色偏好持久化（用户手动切换后刷新保持）
const DARK_THEME_KEY = 'BYTECODE-DARK-THEME';
const storedDark = localStorage.getItem(DARK_THEME_KEY);
const darkTheme = storedDark === null ? designSetting.darkTheme : storedDark === '1';

interface DesignSettingState {
  //深色主题
  darkTheme: boolean;
  //系统风格
  appTheme: string;
  //系统内置风格
  appThemeList: string[];
}

export const useDesignSettingStore = defineStore({
  id: 'app-design-setting',
  state: (): DesignSettingState => ({
    darkTheme,
    appTheme,
    appThemeList,
  }),
  getters: {
    getDarkTheme(): boolean {
      return this.darkTheme;
    },
    getAppTheme(): string {
      return this.appTheme;
    },
    getAppThemeList(): string[] {
      return this.appThemeList;
    },
  },
  actions: {
    toggleDarkTheme() {
      this.darkTheme = !this.darkTheme;
      localStorage.setItem(DARK_THEME_KEY, this.darkTheme ? '1' : '0');
    },
  },
});

// Need to be used outside the setup
export function useDesignSetting() {
  return useDesignSettingStore(store);
}
