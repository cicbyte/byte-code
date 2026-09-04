<template>
  <div class="project-menu">
    <n-menu
      :value="activeKey"
      :options="menuOptions"
      :collapsed="collapsed"
      :collapsed-width="64"
      :collapsed-icon-size="20"
      :indent="24"
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
    BookOutlined,
    FileTextOutlined,
    BulbOutlined,
    FlagOutlined,
    ThunderboltOutlined,
    TeamOutlined,
    DatabaseOutlined,
    BugOutlined,
    ExperimentOutlined,
  } from '@vicons/antd';
  import { useEntityContext } from '@/store/modules/entityContext';

  defineProps<{ collapsed?: boolean }>();

  const route = useRoute();
  const router = useRouter();
  const entityContext = useEntityContext();

  // 优先取路由参数（同步可用，不闪屏）；entityContext 兜底（如子组件内导航）
  const projectId = computed(
    () => Number(route.params.projectId) || entityContext.currentProject?.id
  );

  // 图标按路径段映射——纯视觉，缺省兜底；标题与顺序不在此定义
  const iconBySegment: Record<string, any> = {
    overview: HomeOutlined,
    board: AppstoreOutlined,
    tasks: UnorderedListOutlined,
    requirements: ProfileOutlined,
    milestones: FlagOutlined,
    sprints: ThunderboltOutlined,
    knowledge: BookOutlined,
    docs: FileTextOutlined,
    memories: BulbOutlined,
    'test-cases': BugOutlined,
    'test-plans': ExperimentOutlined,
    members: TeamOutlined,
    database: DatabaseOutlined,
  };

  const renderIcon = (icon: any) =>
    () =>
      h(NIcon, null, {
        default: () => h(icon),
      });

  // 菜单由路由表派生：标题的单一真相源在路由 meta.title（与面包屑同源），
  // 顺序即路由 children 顺序——新增项目页只需在路由表加一条，此处零改动
  const menuOptions = computed(() => {
    const pid = projectId.value;
    if (!pid) return [];
    const ws = route.matched.find((r) => r.name === 'project_workspace');
    if (!ws) return [];
    return ws.children
      .filter((c) => c.meta?.title && !c.redirect)
      .map((c) => {
        const seg = String(c.path);
        return {
          label: String(c.meta!.title),
          key: `/project/${pid}/${seg}`,
          icon: renderIcon(iconBySegment[seg] || AppstoreOutlined),
        };
      });
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
  }
</style>
