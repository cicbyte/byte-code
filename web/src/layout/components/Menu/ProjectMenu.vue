<template>
  <div class="project-menu">
    <n-menu
      :value="activeKey"
      :options="menuOptions"
      :collapsed="collapsed"
      :collapsed-width="64"
      :collapsed-icon-size="20"
      :indent="24"
      :expanded-keys="expanded"
      @update:value="handleSelect"
      @update:expanded-keys="(keys: string[]) => (expanded = keys)"
    />
  </div>
</template>

<script lang="ts" setup>
  import { computed, h, ref, watch } from 'vue';
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
    CarryOutOutlined,
    AimOutlined,
    ReadOutlined,
    SafetyOutlined,
    ControlOutlined,
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

  // 二级项图标按路径段映射——纯视觉，缺省兜底；标题与顺序不在此定义
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

  // 一级分组图标（组名与路由 meta.group 对应）
  const groupIcons: Record<string, any> = {
    任务: CarryOutOutlined,
    规划: AimOutlined,
    知识: ReadOutlined,
    测试: SafetyOutlined,
    管理: ControlOutlined,
  };

  const renderIcon = (icon: any) =>
    () =>
      h(NIcon, null, {
        default: () => h(icon),
      });

  // 菜单由路由表派生：标题/顺序/分组的单一真相源在路由 meta（title/group），
  // 无 group 的子页为顶层单项（项目概览），同 group 连续子页归入该组二级
  const menuOptions = computed(() => {
    const pid = projectId.value;
    if (!pid) return [];
    const ws = route.matched.find((r) => r.name === 'project_workspace');
    if (!ws) return [];
    type Node = { label: string; key: string; icon: any; children?: Node[] };
    const out: Node[] = [];
    const groupNodes = new Map<string, Node>();
    for (const c of ws.children) {
      if (!c.meta?.title || c.redirect) continue;
      const seg = String(c.path);
      const item: Node = {
        label: String(c.meta!.title),
        key: `/project/${pid}/${seg}`,
        icon: renderIcon(iconBySegment[seg] || AppstoreOutlined),
      };
      const g = c.meta.group ? String(c.meta.group) : '';
      if (!g) {
        out.push(item);
        continue;
      }
      let node = groupNodes.get(g);
      if (!node) {
        node = {
          label: g,
          key: `group:${g}`,
          icon: renderIcon(groupIcons[g] || AppstoreOutlined),
          children: [],
        };
        groupNodes.set(g, node);
        out.push(node);
      }
      node.children!.push(item);
    }
    return out;
  });

  // 展平后的叶子（activeKey 查找与组定位用）
  const flatLeaves = computed(() => {
    const leaves: { key: string; group?: string }[] = [];
    for (const o of menuOptions.value as any[]) {
      if (o.children) {
        for (const k of o.children) leaves.push({ key: k.key, group: o.key });
      } else {
        leaves.push({ key: o.key });
      }
    }
    return leaves;
  });

  const activeKey = computed(() => {
    const item = flatLeaves.value.find((m) => route.path.startsWith(m.key));
    return item?.key ?? route.path;
  });

  // 当前路由所在组自动展开——用 watch(route.path) 而非 watchEffect：
  // watchEffect 把 expanded 也作为响应式依赖，用户收起分组后它立即重跑
  // 又把当前组加回去，导致「点已展开的组收不起」死循环
  const expanded = ref<string[]>([]);
  watch(
    () => route.path,
    (path) => {
      const item = flatLeaves.value.find((m) => path.startsWith(m.key));
      if (item?.group && !expanded.value.includes(item.group)) {
        expanded.value = [...expanded.value, item.group];
      }
    },
    { immediate: true }
  );

  function handleSelect(key: string) {
    router.push(key);
  }
</script>

<style lang="less" scoped>
  .project-menu {
    height: 100%;
  }
</style>
