<template>
  <div class="layout-header">
    <!--顶部菜单-->
    <div
      class="layout-header-left"
      v-if="navMode === 'horizontal' || (navMode === 'horizontal-mix' && mixMenu)"
    >
      <div class="logo" v-if="navMode === 'horizontal'">
        <img :src="websiteConfig.logo" alt="" />
        <h2 v-show="!collapsed" class="title">{{ websiteConfig.title }}</h2>
      </div>
      <AsideMenu
        :collapsed="collapsed"
        v-model:location="getMenuLocation"
        :inverted="getInverted"
        mode="horizontal"
      />
    </div>
    <!--左侧菜单-->
    <div class="layout-header-left" v-else>
      <!-- 菜单收起 -->
      <div
        class="ml-1 layout-header-trigger layout-header-trigger-min"
        @click="handleMenuCollapsed"
      >
        <n-icon size="18" v-if="collapsed">
          <MenuUnfoldOutlined />
        </n-icon>
        <n-icon size="18" v-else>
          <MenuFoldOutlined />
        </n-icon>
      </div>
      <!-- 刷新 -->
      <div
        class="mr-1 layout-header-trigger layout-header-trigger-min"
        v-if="headerSetting.isReload"
        @click="reloadPage"
      >
        <n-icon size="18">
          <ReloadOutlined />
        </n-icon>
      </div>
      <!-- 面包屑 -->
      <n-breadcrumb v-if="crumbsSetting.show">
        <n-breadcrumb-item>
          <span class="link-text clickable" @click="goHome">
            <n-icon size="13"><HomeOutlined /></n-icon>
            首页
          </span>
        </n-breadcrumb-item>
        <n-breadcrumb-item
          v-for="(routeItem, idx) in breadcrumbList"
          :key="routeItem.name === RedirectName ? void 0 : routeItem.name"
        >
          <span
            v-if="routeItem.meta.title"
            class="link-text"
            :class="{ clickable: isNavigable(routeItem) && idx < breadcrumbList.length - 1 }"
            @click="navigateCrumb(routeItem)"
          >
            <component
              v-if="crumbsSetting.showIcon && routeItem.meta.icon"
              :is="routeItem.meta.icon"
            />
            {{ routeItem.meta.title }}
          </span>
        </n-breadcrumb-item>
      </n-breadcrumb>
      <span v-if="entityContext.currentEntityName" class="entity-breadcrumb-name">
        / {{ entityContext.currentEntityName }}
      </span>
    </div>
    <div class="layout-header-right">
      <div
        class="layout-header-trigger layout-header-trigger-min"
        v-for="item in iconList"
        :key="item.icon"
      >
        <n-tooltip placement="bottom">
          <template #trigger>
            <n-icon size="18">
              <component :is="item.icon" v-on="item.eventObject || {}" />
            </n-icon>
          </template>
          <span>{{ item.tips }}</span>
        </n-tooltip>
      </div>
      <!--明暗模式切换-->
      <ThemeToggle />
            <!--通知中心-->
      <NotificationIcon />
      <!--切换全屏-->
      <div class="layout-header-trigger layout-header-trigger-min">
        <n-tooltip placement="bottom">
          <template #trigger>
            <n-icon size="18">
              <component :is="fullscreenIcon" @click="toggleFullScreen" />
            </n-icon>
          </template>
          <span>全屏</span>
        </n-tooltip>
      </div>
      <!-- 个人中心 -->
      <div class="layout-header-trigger layout-header-trigger-min">
        <n-dropdown trigger="hover" @select="avatarSelect" :options="avatarOptions">
          <div class="avatar">
            <n-avatar :src="websiteConfig.logo">
              <template #icon>
                <UserOutlined />
              </template>
            </n-avatar>
            <n-divider vertical />
            <span>{{ username }}</span>
          </div>
        </n-dropdown>
      </div>
      <!--设置-->
      <div class="layout-header-trigger layout-header-trigger-min" @click="openSetting">
        <n-tooltip placement="bottom-end">
          <template #trigger>
            <n-icon size="18" style="font-weight: bold">
              <SettingOutlined />
            </n-icon>
          </template>
          <span>项目配置</span>
        </n-tooltip>
      </div>
    </div>
  </div>
  <!--项目配置-->
  <ProjectSetting ref="drawerSetting" />
  <!--全局搜索-->
  <SearchModal ref="searchRef" />
</template>

<script lang="ts">
  import { websiteConfig } from '@/config/website.config';
  import { useProjectSetting } from '@/hooks/setting/useProjectSetting';
  import { AsideMenu } from '@/layout/components/Menu';
  import { RedirectName } from '@/router/constant';
  import { useEntityContext } from '@/store/modules/entityContext';
  import { useUserStore } from '@/store/modules/user';
  import { NDialogProvider, useDialog, useMessage } from 'naive-ui';
  import { computed, defineComponent, reactive, ref, toRefs, unref } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import components from './components';
  import ProjectSetting from './ProjectSetting.vue';
  import NotificationIcon from './NotificationIcon.vue';
  import ThemeToggle from './ThemeToggle.vue';
  import SearchModal from './SearchModal.vue';
  import { ref as vueRef } from 'vue';
  import { HomeOutlined } from '@vicons/antd';

  export default defineComponent({
    name: 'PageHeader',
    components: { ...components, NDialogProvider, ProjectSetting, AsideMenu, NotificationIcon, ThemeToggle, SearchModal, HomeOutlined },
    props: {
      collapsed: {
        type: Boolean,
      },
      inverted: {
        type: Boolean,
      },
    },
    emits: ['update:collapsed'],
    setup(props, { emit }) {
      const userStore = useUserStore();
      const entityContext = useEntityContext();
      const message = useMessage();
      const dialog = useDialog();
      const searchRef = vueRef(null);
      const { navMode, navTheme, headerSetting, menuSetting, crumbsSetting } = useProjectSetting();

      const drawerSetting = ref();

      const state = reactive({
        username: userStore?.info?.username ?? '',
        fullscreenIcon: 'FullscreenOutlined',
        navMode,
        navTheme,
        headerSetting,
        crumbsSetting,
      });

      const getInverted = computed(() => {
        return ['light', 'header-dark'].includes(unref(navTheme))
          ? props.inverted
          : !props.inverted;
      });

      const mixMenu = computed(() => {
        return unref(menuSetting).mixMenu;
      });

      const getChangeStyle = computed(() => {
        const { collapsed } = props;
        const { minMenuWidth, menuWidth } = unref(menuSetting);
        return {
          left: collapsed ? `${minMenuWidth}px` : `${menuWidth}px`,
          width: `calc(100% - ${collapsed ? `${minMenuWidth}px` : `${menuWidth}px`})`,
        };
      });

      const getMenuLocation = computed(() => {
        return 'header';
      });

      const router = useRouter();
      const route = useRoute();

      const generator: any = (routerMap) => {
        return routerMap
          .filter((item) => item.meta?.title)
          .map((item) => {
            const currentMenu = {
              ...item,
              label: item.meta.title,
              key: item.name,
              disabled: item.path === '/',
            };
            if (item.children && item.children.length > 0) {
              currentMenu.children = generator(item.children);
            }
            return currentMenu;
          });
      };

      const breadcrumbList = computed(() => {
        return generator(route.matched);
      });

      // 面包屑导航：无悬浮下拉，层级可点击回跳；参数化路由按段数取当前实际路径前缀
      const goHome = () => {
        router.push('/dashboard/console');
      };

      const crumbPath = (pattern: string) => {
        if (!pattern.includes(':')) return pattern;
        const seg = pattern.split('/').filter(Boolean).length;
        const cur = route.path.split('/').filter(Boolean);
        return '/' + cur.slice(0, seg).join('/');
      };

      const isNavigable = (item) => {
        const p = crumbPath(item.path || '');
        return !!p && p !== '/' && p !== route.path;
      };

      const navigateCrumb = (item) => {
        if (isNavigable(item)) router.push(crumbPath(item.path));
      };

      // 刷新页面
      const reloadPage = () => {
        router.push({
          path: '/redirect' + unref(route).fullPath,
        });
      };

      // 退出登录
      const doLogout = () => {
        dialog.info({
          title: '提示',
          content: '您确定要退出登录吗',
          positiveText: '确定',
          negativeText: '取消',
          onPositiveClick: () => {
            userStore.logout().then(() => {
              message.success('成功退出登录');
              router
                .replace({
                  name: 'Login',
                  query: {
                    redirect: route.fullPath,
                  },
                })
                .finally(() => location.reload());
            });
          },
          onNegativeClick: () => {},
        });
      };

      // 切换全屏图标
      const toggleFullscreenIcon = () =>
        (state.fullscreenIcon =
          document.fullscreenElement !== null ? 'FullscreenExitOutlined' : 'FullscreenOutlined');

      // 监听全屏切换事件
      document.addEventListener('fullscreenchange', toggleFullscreenIcon);

      // 全屏切换
      const toggleFullScreen = () => {
        if (!document.fullscreenElement) {
          document.documentElement.requestFullscreen();
        } else {
          if (document.exitFullscreen) {
            document.exitFullscreen();
          }
        }
      };

      // 图标列表
      const iconList = [
        {
          icon: 'SearchOutlined',
          tips: '搜索',
          eventObject: {
            click: () => searchRef.value?.open(),
          },
        },
        {
          icon: 'GithubOutlined',
          tips: 'github',
          eventObject: {
            click: () => window.open('https://github.com/cicbyte/byte-code'),
          },
        },
      ];
      const avatarOptions = [
        {
          label: '个人设置',
          key: 1,
        },
        {
          label: '退出登录',
          key: 2,
        },
      ];

      //头像下拉菜单
      const avatarSelect = (key) => {
        switch (key) {
          case 1:
            router.push({ name: 'Setting' });
            break;
          case 2:
            doLogout();
            break;
        }
      };

      function openSetting() {
        const { openDrawer } = drawerSetting.value;
        openDrawer();
      }

      function handleMenuCollapsed() {
        emit('update:collapsed', !props.collapsed);
      }

      return {
        ...toRefs(state),
        iconList,
        searchRef,
        goHome,
        isNavigable,
        navigateCrumb,
        toggleFullScreen,
        doLogout,
        route,
        avatarOptions,
        getChangeStyle,
        avatarSelect,
        breadcrumbList,
        reloadPage,
        drawerSetting,
        openSetting,
        getInverted,
        getMenuLocation,
        mixMenu,
        websiteConfig,
        handleMenuCollapsed,
        entityContext,
        RedirectName,
      };
    },
  });
</script>

<style lang="less" scoped>
  .layout-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0;
    height: 64px;
    width: 100%;
    z-index: 11;
    // 悬浮栏卡片：与画布/侧栏留 8px 细缝
    margin: 8px 8px 0;
    width: calc(100% - 16px);
    border-radius: var(--panel-radius, 12px);
    background: var(--panel-bg, #fff);
    box-shadow: var(--panel-shadow);
    transition: box-shadow 0.3s ease-in-out;

    &-left {
      display: flex;
      align-items: center;

      .logo {
        display: flex;
        align-items: center;
        justify-content: center;
        height: 64px;
        line-height: 64px;
        overflow: hidden;
        white-space: nowrap;
        padding-left: 10px;

        img {
          width: auto;
          height: 32px;
          margin-right: 10px;
        }

        .title {
          margin-bottom: 0;
        }
      }

      ::v-deep(.ant-breadcrumb span:last-child .link-text) {
        color: #515a6e;
      }

      .n-breadcrumb {
        display: inline-block;
      }

      .entity-breadcrumb-name {
        font-size: 13px;
        color: #515a6e;
        font-weight: 500;
      }

      &-menu {
        color: var(--text-color);
      }
    }

    .link-text.clickable {
      color: var(--n-text-color-3, #97999d);
      cursor: pointer;
      transition: color 0.2s;

      &:hover {
        color: #16a34a;
      }
    }

    &-right {
      display: flex;
      align-items: center;
      margin-right: 20px;

      .avatar {
        display: flex;
        align-items: center;
        height: 64px;
      }

      > * {
        cursor: pointer;
      }
    }

    &-trigger {
      display: inline-block;
      width: 64px;
      height: 64px;
      text-align: center;
      cursor: pointer;
      transition: all 0.2s ease-in-out;

      .n-icon {
        display: flex;
        align-items: center;
        height: 64px;
        line-height: 64px;
      }

      &:hover {
        background: hsla(0, 0%, 100%, 0.08);
      }

      .anticon {
        font-size: 16px;
        color: #515a6e;
      }
    }

    &-trigger-min {
      width: auto;
      padding: 0 12px;
    }
  }

  .layout-header-light {
    background: #fff;
    color: #515a6e;

    .n-icon {
      color: #515a6e;
    }

    .layout-header-left {
      ::v-deep(.n-breadcrumb .n-breadcrumb-item:last-child .n-breadcrumb-item__link) {
        color: #515a6e;
      }
    }

    .layout-header-trigger {
      &:hover {
        background: #f8f8f9;
      }
    }
  }

  .layout-header-fix {
    position: fixed;
    top: 0;
    right: 0;
    left: 200px;
    z-index: 11;
  }

  //::v-deep(.menu-router-link) {
  //  color: #515a6e;
  //
  //  &:hover {
  //    color: #1890ff;
  //  }
  //}
</style>
