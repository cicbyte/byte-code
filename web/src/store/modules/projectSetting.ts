import { defineStore } from 'pinia';
import projectSetting from '@/settings/projectSetting';
import type { IHeaderSetting, IMenuSetting, ICrumbsSetting } from '/#/config';

const {
  navMode,
  navTheme,
  isMobile,
  headerSetting,
  menuSetting,
  crumbsSetting,
  isPageAnimate,
  pageAnimateType,
} = projectSetting;

interface ProjectSettingState {
  navMode: string;
  navTheme: string;
  headerSetting: IHeaderSetting;
  menuSetting: IMenuSetting;
  crumbsSetting: ICrumbsSetting;
  isPageAnimate: boolean;
  pageAnimateType: string;
  isMobile: boolean;
}

export const useProjectSettingStore = defineStore({
  id: 'app-project-setting',
  state: (): ProjectSettingState => ({
    navMode: navMode,
    navTheme,
    isMobile,
    headerSetting,
      menuSetting,
    crumbsSetting,
    isPageAnimate,
    pageAnimateType,
  }),
  getters: {
    getNavMode(): string {
      return this.navMode;
    },
    getNavTheme(): string {
      return this.navTheme;
    },
    getIsMobile(): boolean {
      return this.isMobile;
    },
    getHeaderSetting(): object {
      return this.headerSetting;
    },
    getMenuSetting(): object {
      return this.menuSetting;
    },
    getCrumbsSetting(): object {
      return this.crumbsSetting;
    },
    getIsPageAnimate(): boolean {
      return this.isPageAnimate;
    },
    getPageAnimateType(): string {
      return this.pageAnimateType;
    },
  },
  actions: {
    setNavTheme(value: string): void {
      this.navTheme = value;
    },
    setIsMobile(value: boolean): void {
      this.isMobile = value;
    },
  },
});
