<template>
  <n-popover trigger="hover" placement="bottom" :width="280" :show-arrow="false" raw>
    <template #trigger>
      <div class="layout-header-trigger layout-header-trigger-min project-switch" title="切换项目">
        <n-icon size="16"><AppstoreOutlined /></n-icon>
        <span v-if="currentName" class="ps-name">{{ currentName }}</span>
        <n-icon size="11" class="ps-caret"><CaretDownOutlined /></n-icon>
      </div>
    </template>

    <div class="ps-panel">
      <div class="ps-head">
        <span>切换项目</span>
        <n-radio-group v-model:value="viewMode" size="tiny">
          <n-radio-button value="project">项目</n-radio-button>
          <n-radio-button value="group">分组</n-radio-button>
        </n-radio-group>
      </div>
      <n-spin :show="loading" size="small">
        <!-- 项目视图：平铺列表（原逻辑不变） -->
        <template v-if="viewMode === 'project'">
          <div
            v-for="p in projects"
            :key="p.id"
            class="ps-item"
            :class="{ cur: p.id === currentProjectId }"
            @click="onSelect(p.id)"
          >
            <span class="ps-item-name">{{ p.name }}</span>
            <span class="ps-item-code">{{ p.code }}</span>
            <svg v-if="p.id === currentProjectId" class="ps-check" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="m5 13 4 4L19 7"/></svg>
          </div>
          <div v-if="!loading && projects.length === 0" class="ps-empty">暂无项目</div>
        </template>

        <!-- 分组视图：按分组折叠展示 -->
        <template v-else>
          <template v-for="g in groups" :key="g.id">
            <div class="ps-group-head" @click="toggleGroup(g.id)">
              <n-icon size="10" class="ps-group-caret" :class="{ open: expandedGroups.has(g.id) }">
                <CaretRightOutlined />
              </n-icon>
              <span class="ps-group-name">{{ g.name }}</span>
              <span class="ps-group-count">{{ (g.projects || []).length }}</span>
            </div>
            <template v-if="expandedGroups.has(g.id)">
              <div
                v-for="p in g.projects"
                :key="p.id"
                class="ps-item ps-group-child"
                :class="{ cur: p.id === currentProjectId }"
                @click="onSelect(p.id)"
              >
                <span class="ps-item-name">{{ p.name }}</span>
                <span class="ps-item-code">{{ p.code }}</span>
                <svg v-if="p.id === currentProjectId" class="ps-check" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="m5 13 4 4L19 7"/></svg>
              </div>
            </template>
          </template>
          <div v-if="!loading && groups.length === 0" class="ps-empty">暂无分组</div>
        </template>
      </n-spin>
      <div class="ps-foot" @click="goList">管理项目 ›</div>
    </div>
  </n-popover>
</template>

<script lang="ts" setup>
  import { ref, reactive, computed, onMounted } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { AppstoreOutlined, CaretDownOutlined, CaretRightOutlined } from '@vicons/antd';
  import { getProjects } from '@/api/project/index';
  import { getGroups } from '@/api/project/group';
  import type { GroupItem } from '@/api/project/group';

  const route = useRoute();
  const router = useRouter();
  const loading = ref(false);
  const projects = ref<Array<{ id: number; name: string; code: string }>>([]);
  const groups = ref<GroupItem[]>([]);
  const viewMode = ref<'project' | 'group'>('project');
  const expandedGroups = reactive(new Set<number>());

  const currentProjectId = computed(() => Number(route.params.projectId) || 0);
  const currentName = computed(
    () => projects.value.find((p) => p.id === currentProjectId.value)?.name || '',
  );

  function toggleGroup(id: number) {
    if (expandedGroups.has(id)) expandedGroups.delete(id);
    else expandedGroups.add(id);
  }

  function onSelect(id: number) {
    if (id === currentProjectId.value) return;
    const map: Record<string, string> = {
      board: 'board', tasks: 'tasks', reviews: 'reviews', feedbacks: 'feedbacks',
      topics: 'topics', requirements: 'requirements', milestones: 'milestones',
      sprints: 'sprints', knowledge: 'knowledge', docs: 'docs', memories: 'memories',
      qas: 'qas', 'test-cases': 'test-cases', 'test-plans': 'test-plans', members: 'members',
    };
    const seg = route.path.split('/').filter(Boolean)[2] || 'overview';
    const target = map[seg] ? `/project/${id}/${map[seg]}` : `/project/${id}/overview`;
    router.push(target);
  }

  function goList() {
    router.push('/project/list');
  }

  onMounted(async () => {
    loading.value = true;
    try {
      const [pRes, gRes] = await Promise.all([
        getProjects({ size: 50 }),
        viewMode.value === 'group' ? getGroups() : Promise.resolve(null),
      ]);
      projects.value = pRes?.list || [];
      if (gRes) groups.value = gRes?.list || [];
    } catch {
      // ignore
    } finally {
      loading.value = false;
    }
  });
</script>

<style lang="less" scoped>
  .project-switch {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 0 10px;
    height: 100%;
    cursor: pointer;
    transition: background 0.2s;

    &:hover {
      background: var(--hover-bg, #f5f5f5);
    }

    .ps-name {
      max-width: 140px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      font-size: 13px;
      color: var(--text-1, #333);
    }

    .ps-caret {
      color: var(--text-3, #999);
      transition: transform 0.2s;
    }

    &:hover .ps-caret {
      color: var(--primary-color, #16a34a);
    }
  }
</style>

<style lang="less">
  .ps-panel {
    margin: -4px -8px;
  }

  .ps-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 6px 12px 8px;
    border-bottom: 1px solid var(--line, #efeff5);
    font-size: 13px;
    font-weight: 500;
  }

  .ps-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 7px 12px;
    cursor: pointer;
    transition: background 0.15s;

    &:hover {
      background: var(--hover-bg, #f8f8f6);
    }

    &.cur {
      background: rgba(22, 163, 74, 0.06);

      .ps-item-name {
        color: var(--primary-color, #16a34a);
        font-weight: 500;
      }
    }
  }

  .ps-item-name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 13px;
    color: var(--text-1, #333);
  }

  .ps-item-code {
    flex: none;
    font-family: 'JetBrains Mono', Consolas, monospace;
    font-size: 11px;
    color: var(--text-3, #999);
    background: var(--hover-bg, #f4f4f2);
    padding: 1px 5px;
    border-radius: 4px;
  }

  .ps-check {
    flex: none;
    width: 14px;
    height: 14px;
    color: var(--primary-color, #16a34a);
  }

  .ps-group-head {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px 4px;
    cursor: pointer;
    font-size: 12px;
    color: var(--text-2, #57606a);
    font-weight: 500;

    &:hover {
      color: var(--primary-color, #16a34a);
    }

    .ps-group-caret {
      transition: transform 0.2s;
      &.open {
        transform: rotate(90deg);
      }
    }

    .ps-group-name {
      flex: 1;
    }

    .ps-group-count {
      font-size: 10px;
      color: var(--text-3, #999);
    }
  }

  .ps-group-child {
    padding-left: 28px;
  }

  .ps-empty {
    padding: 16px;
    text-align: center;
    color: var(--text-3, #999);
    font-size: 12px;
  }

  .ps-foot {
    padding: 8px 12px;
    border-top: 1px solid var(--line, #efeff5);
    text-align: center;
    font-size: 12px;
    color: var(--text-3, #999);
    cursor: pointer;
    transition: color 0.15s;

    &:hover {
      color: var(--primary-color, #16a34a);
    }
  }
</style>
