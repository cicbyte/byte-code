<template>
  <div class="project-menu">
    <div class="project-menu-back" @click="goList">
      <n-icon size="14"><ArrowLeftOutlined /></n-icon>
      <span>所有项目</span>
    </div>

    <div class="project-menu-title" :title="projectName">
      <n-icon size="15" class="project-menu-title-icon"><FolderOpenOutlined /></n-icon>
      <span class="project-menu-title-name">{{ projectName || '项目' }}</span>
    </div>

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
    ArrowLeftOutlined,
    FolderOpenOutlined,
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
  const projectName = computed(() => entityContext.currentEntityName);

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

  function goList() {
    router.push('/project/list');
  }
</script>

<style lang="less" scoped>
  .project-menu {
    display: flex;
    flex-direction: column;
    height: 100%;
    padding-top: 4px;

    .project-menu-back {
      display: flex;
      gap: 6px;
      align-items: center;
      padding: 8px 16px;
      color: var(--n-text-color-3, #97999d);
      font-size: 12px;
      cursor: pointer;
      transition: color 0.2s;

      &:hover {
        color: var(--n-text-color-1, #1c1d21);
      }
    }

    .project-menu-title {
      display: flex;
      gap: 8px;
      align-items: center;
      padding: 10px 16px 12px;
      font-size: 13px;
      font-weight: 600;

      .project-menu-title-icon {
        flex: none;
        color: #16a34a;
      }

      .project-menu-title-name {
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }
  }
</style>
