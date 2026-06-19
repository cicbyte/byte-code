<template>
  <RouterView>
    <template #default="{ Component, route }">
      <template v-if="mode === 'production'">
        <transition :name="getTransitionName" mode="out-in" appear>
          <keep-alive v-if="keepAliveComponents.length" :include="keepAliveComponents">
            <component :is="Component" :key="routeKey(route)" />
          </keep-alive>
          <component v-else :is="Component" :key="routeKey(route)" />
        </transition>
      </template>
      <template v-else>
        <keep-alive v-if="keepAliveComponents.length" :include="keepAliveComponents">
          <component :is="Component" :key="routeKey(route)" />
        </keep-alive>
        <component v-else :is="Component" :key="routeKey(route)" />
      </template>
    </template>
  </RouterView>
</template>

<script>
  import { defineComponent, computed, unref } from 'vue';
  import { useAsyncRouteStore } from '@/store/modules/asyncRoute';
  import { useProjectSetting } from '@/hooks/setting/useProjectSetting';

  export default defineComponent({
    name: 'MainView',
    components: {},
    props: {
      notNeedKey: {
        type: Boolean,
        default: false,
      },
      animate: {
        type: Boolean,
        default: true,
      },
    },
    setup() {
      const { isPageAnimate, pageAnimateType } = useProjectSetting();
      const asyncRouteStore = useAsyncRouteStore();
      // 需要缓存的路由组件
      const keepAliveComponents = computed(() => asyncRouteStore.keepAliveComponents);

      const getTransitionName = computed(() => {
        return unref(isPageAnimate) ? unref(pageAnimateType) : '';
      });

      const mode = import.meta.env.MODE;

      // Layout 是 matched[0]，MainView 渲染 matched[1] 的组件。
      // 用 matched[1] 实际路径段作为 key：
      //   /product/1/overview 与 /product/1/requirements 都映射到 /product/1，不重建 ProductWorkspace；
      //   /product/1 与 /product/2 不同 key，仍会重建。
      const routeKey = (route) => {
        const m = route.matched[1];
        if (!m) return route.fullPath;
        const segCount = m.path.split('/').filter(Boolean).length;
        const segs = route.path.split('/').filter(Boolean).slice(0, segCount);
        return '/' + segs.join('/');
      };

      return {
        keepAliveComponents,
        getTransitionName,
        mode,
        routeKey,
      };
    },
  });
</script>

<style lang="less" scoped></style>
