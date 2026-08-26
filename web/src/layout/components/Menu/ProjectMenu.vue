<template>
  <div class="project-menu">
    <n-menu
      :value="activeKey"
      :options="menuOptions"
      :collapsed="collapsed"
      :collapsed-width="64"
      :collapsed-icon-size="18"
      :indent="18"
      @update:value="handleSelect"
    />
  </div>
</template>

<script lang="ts" setup>
  import { computed, h } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { NIcon } from 'naive-ui';
  import {
    HomeOutlined,
    AppstoreOutlined,
    UnorderedListOutlined,
    ProfileOutlined,
    FlagOutlined,
    ThunderboltOutlined,
    TeamOutlined,
    DatabaseOutlined,
  } from '@vicons/antd';
  import { useEntityContext } from '@/store/modules/entityContext';

  defineProps<{ collapsed?: boolean }>();

  const route = useRoute();
  const router = useRouter();
  const entityContext = useEntityContext();

  const projectId = computed(() => entityContext.currentProject?.id);

  const renderIcon = (icon: any) =>
    () =>
      h(NIcon, null, {
        default: () => h(icon),
      });

  const menuOptions = computed(() => {
    const pid = projectId.value;
    if (!pid) return [];
    return [
      { label: '概览', key: `/project/${pid}/overview`, icon: renderIcon(HomeOutlined) },
      { label: '任务看板', key: `/project/${pid}/board`, icon: renderIcon(AppstoreOutlined) },
      { label: '任务列表', key: `/project/${pid}/tasks`, icon: renderIcon(UnorderedListOutlined) },
      { label: '需求池', key: `/project/${pid}/requirements`, icon: renderIcon(ProfileOutlined) },
      { label: '里程碑', key: `/project/${pid}/milestones`, icon: renderIcon(FlagOutlined) },
      { label: 'Sprint', key: `/project/${pid}/sprints`, icon: renderIcon(ThunderboltOutlined) },
      { label: '成员', key: `/project/${pid}/members`, icon: renderIcon(TeamOutlined) },
      { label: '数据库', key: `/project/${pid}/database`, icon: renderIcon(DatabaseOutlined) },
    ];
  });

  const activeKey = computed(() => {
    // 精确匹配当前路由（子路径如 /project/1/tasks 时高亮对应项）
    const item = menuOptions.value.find((m) => route.path.startsWith(m.key));
    return item?.key ?? route.path;
  });

  function handleSelect(key: string) {
    router.push(key);
  }
</script>

<style lang="less" scoped>
  .project-menu {
    height: 100%;
    padding-top: 4px;
  }
</style>
