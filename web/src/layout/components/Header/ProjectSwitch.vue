<template>
  <n-popover trigger="hover" placement="bottom" :width="300" :show-arrow="false" raw>
    <template #trigger>
      <div class="layout-header-trigger layout-header-trigger-min project-switch" title="切换项目">
        <n-icon size="16"><AppstoreOutlined /></n-icon>
        <span v-if="currentName" class="ps-name">{{ currentName }}</span>
        <n-icon size="11" class="ps-caret"><CaretDownOutlined /></n-icon>
      </div>
    </template>

    <div class="ps-panel">
      <!-- 头部：标题 + 视图切换 -->
      <div class="ps-head">
        <span class="ps-head-label">切换项目</span>
        <div class="ps-mode-switch">
          <button
            v-for="m in [
              { key: 'project', label: '项目', icon: '☰' },
              { key: 'group', label: '分组', icon: '⊞' },
            ]"
            :key="m.key"
            class="ps-mode-btn"
            :class="{ active: viewMode === m.key }"
            @click="viewMode = m.key as any"
          >
            <span class="ps-mode-icon">{{ m.icon }}</span>
            {{ m.label }}
          </button>
        </div>
      </div>

      <!-- 搜索框 -->
      <div class="ps-search">
        <n-input
          v-model:value="searchKw"
          size="tiny"
          placeholder="搜索项目…"
          clearable
          :bordered="true"
        >
          <template #prefix>
            <n-icon size="12" color="#999"><SearchOutlined /></n-icon>
          </template>
        </n-input>
      </div>

      <n-spin :show="loading" size="small">
        <div class="ps-list">
          <!-- 项目视图 -->
          <template v-if="viewMode === 'project'">
            <div
              v-for="p in filteredProjects"
              :key="p.id"
              class="ps-item"
              :class="{ cur: p.id === currentProjectId }"
              @click="onSelect(p.id)"
            >
              <div class="ps-item-dot" :class="{ cur: p.id === currentProjectId }" />
              <div class="ps-item-body">
                <span class="ps-item-name">{{ p.name }}</span>
                <span class="ps-item-code">{{ p.code }}</span>
              </div>
              <svg v-if="p.id === currentProjectId" class="ps-check" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="m5 13 4 4L19 7"/></svg>
            </div>
            <div v-if="!loading && filteredProjects.length === 0" class="ps-empty">
              {{ searchKw ? '无匹配项目' : '暂无项目' }}
            </div>
          </template>

          <!-- 分组视图 -->
          <template v-else>
            <template v-for="g in filteredGroups" :key="g.id">
              <div class="ps-group-head" @click="toggleGroup(g.id)">
                <n-icon size="10" class="ps-group-caret" :class="{ open: expandedGroups.has(g.id) }">
                  <CaretRightOutlined />
                </n-icon>
                <span class="ps-group-name">{{ g.name }}</span>
                <span class="ps-group-badge">{{ (g.projects || []).length }}</span>
              </div>
              <div v-if="expandedGroups.has(g.id)" class="ps-group-body">
                <div
                  v-for="p in g.projects"
                  :key="p.id"
                  class="ps-item"
                  :class="{ cur: p.id === currentProjectId }"
                  @click="onSelect(p.id)"
                >
                  <div class="ps-item-dot" :class="{ cur: p.id === currentProjectId }" />
                  <div class="ps-item-body">
                    <span class="ps-item-name">{{ p.name }}</span>
                    <span class="ps-item-code">{{ p.code }}</span>
                  </div>
                  <svg v-if="p.id === currentProjectId" class="ps-check" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="m5 13 4 4L19 7"/></svg>
                </div>
                <div v-if="(g.projects || []).length === 0" class="ps-empty slim">组内暂无项目</div>
              </div>
            </template>
            <div v-if="!loading && filteredGroups.length === 0" class="ps-empty">
              {{ searchKw ? '无匹配分组' : '暂无分组' }}
            </div>
          </template>
        </div>
      </n-spin>

      <div class="ps-foot" @click="goList">
        <n-icon size="12"><SettingOutlined /></n-icon>
        管理项目
        <span class="ps-foot-arrow">›</span>
      </div>
    </div>
  </n-popover>
</template>

<script lang="ts" setup>
  import { ref, reactive, computed, onMounted, watch } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import {
    AppstoreOutlined,
    CaretDownOutlined,
    CaretRightOutlined,
    SearchOutlined,
    SettingOutlined,
  } from '@vicons/antd';
  import { getProjects } from '@/api/project/index';
  import { getGroups } from '@/api/project/group';
  import type { GroupItem } from '@/api/project/group';

  const route = useRoute();
  const router = useRouter();
  const loading = ref(false);
  const projects = ref<Array<{ id: number; name: string; code: string }>>([]);
  const groups = ref<GroupItem[]>([]);
  const viewMode = ref<'project' | 'group'>('project');
  const searchKw = ref('');
  const expandedGroups = reactive(new Set<number>());

  const currentProjectId = computed(() => Number(route.params.projectId) || 0);
  const currentName = computed(
    () => projects.value.find((p) => p.id === currentProjectId.value)?.name || '',
  );

  const filteredProjects = computed(() => {
    if (!searchKw.value.trim()) return projects.value;
    const kw = searchKw.value.toLowerCase();
    return projects.value.filter((p) => p.name.toLowerCase().includes(kw) || p.code.includes(kw));
  });

  const filteredGroups = computed(() => {
    if (!searchKw.value.trim()) return groups.value;
    const kw = searchKw.value.toLowerCase();
    return groups.value.filter(
      (g) =>
        g.name.toLowerCase().includes(kw) ||
        (g.projects || []).some((p) => p.name.toLowerCase().includes(kw)),
    );
  });

  function toggleGroup(id: number) {
    if (expandedGroups.has(id)) expandedGroups.delete(id);
    else expandedGroups.add(id);
  }

  // 搜索时自动展开所有匹配的分组
  watch(searchKw, (kw) => {
    if (viewMode.value === 'group' && kw.trim()) {
      groups.value.forEach((g) => expandedGroups.add(g.id));
    }
  });

  function onSelect(id: number) {
    if (id === currentProjectId.value) return;
    const map: Record<string, string> = {
      board: 'board', tasks: 'tasks', reviews: 'reviews', feedbacks: 'feedbacks',
      topics: 'topics', requirements: 'requirements', milestones: 'milestones',
      sprints: 'sprints', knowledge: 'knowledge', docs: 'docs', memories: 'memories',
      qas: 'qas', 'test-cases': 'test-cases', 'test-plans': 'test-plans', members: 'members',
      settings: 'settings',
    };
    const seg = route.path.split('/').filter(Boolean)[2] || 'overview';
    const target = map[seg] ? `/project/${id}/${map[seg]}` : `/project/${id}/overview`;
    router.push(target);
  }

  function goList() {
    router.push('/project/list');
  }

  // 分组数据懒加载：切到分组视图时才拉
  watch(viewMode, async (mode) => {
    if (mode === 'group' && groups.value.length === 0 && !loading.value) {
      try {
        const res = await getGroups();
        groups.value = res?.list || [];
      } catch { /* ignore */ }
    }
  });

  onMounted(async () => {
    loading.value = true;
    try {
      const res = await getProjects({ size: 50 });
      projects.value = res?.list || [];
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
    margin: 0;
    background: #fff;
    border-radius: 12px;
    border: 1px solid rgba(0, 0, 0, 0.06);
    box-shadow:
      0 4px 16px rgba(0, 0, 0, 0.08),
      0 1px 4px rgba(0, 0, 0, 0.06);
    overflow: hidden;
  }

  .ps-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 14px 8px;
    border-bottom: 1px solid rgba(0, 0, 0, 0.04);
  }

  .ps-head-label {
    font-size: 13px;
    font-weight: 600;
    color: var(--text-1, #1f2328);
    letter-spacing: 0.3px;
  }

  .ps-mode-switch {
    display: flex;
    gap: 2px;
    padding: 2px;
    border-radius: 6px;
    background: rgba(0, 0, 0, 0.04);
  }

  .ps-mode-btn {
    display: flex;
    align-items: center;
    gap: 3px;
    padding: 3px 8px;
    border: none;
    border-radius: 5px;
    background: transparent;
    font-size: 11px;
    color: var(--text-3, #8b949e);
    cursor: pointer;
    transition: all 0.15s;
    outline: none;

    &.active {
      background: #fff;
      color: var(--primary-color, #16a34a);
      font-weight: 500;
      box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    }

    .ps-mode-icon {
      font-size: 10px;
      line-height: 1;
    }
  }

  .ps-search {
    padding: 6px 10px;
    border-bottom: 1px solid rgba(0, 0, 0, 0.04);

    .n-input {
      --n-height: 26px;
      --n-border-radius: 6px;
    }
  }

  .ps-list {
    max-height: 320px;
    overflow-y: auto;
    padding: 4px;

    &::-webkit-scrollbar {
      width: 4px;
    }
    &::-webkit-scrollbar-thumb {
      border-radius: 2px;
      background: rgba(0, 0, 0, 0.12);
    }
  }

  .ps-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 7px 10px;
    border-radius: 8px;
    cursor: pointer;
    transition: background 0.12s;

    &:hover {
      background: rgba(0, 0, 0, 0.035);
    }

    &.cur {
      background: rgba(22, 163, 74, 0.07);

      .ps-item-name {
        color: var(--primary-color, #16a34a);
        font-weight: 500;
      }

      .ps-item-dot {
        background: var(--primary-color, #16a34a);
        border-color: var(--primary-color, #16a34a);
      }
    }
  }

  .ps-item-dot {
    flex: none;
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: rgba(0, 0, 0, 0.15);
    transition: all 0.15s;

    &.cur {
      background: var(--primary-color, #16a34a);
    }
  }

  .ps-item-body {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .ps-item-name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 13px;
    color: var(--text-1, #1f2328);
  }

  .ps-item-code {
    flex: none;
    font-family: 'JetBrains Mono', Consolas, monospace;
    font-size: 10px;
    color: var(--text-3, #8b949e);
    background: rgba(0, 0, 0, 0.045);
    padding: 1px 5px;
    border-radius: 4px;
    letter-spacing: 0.3px;
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
    gap: 5px;
    padding: 8px 10px 4px;
    cursor: pointer;
    border-radius: 8px;
    transition: background 0.12s;

    &:hover {
      background: rgba(0, 0, 0, 0.03);

      .ps-group-name {
        color: var(--primary-color, #16a34a);
      }
    }
  }

  .ps-group-caret {
    color: var(--text-3, #8b949e);
    transition: transform 0.2s;
    &.open {
      transform: rotate(90deg);
      color: var(--primary-color, #16a34a);
    }
  }

  .ps-group-name {
    flex: 1;
    font-size: 12px;
    font-weight: 600;
    color: var(--text-2, #57606a);
    letter-spacing: 0.2px;
    transition: color 0.12s;
  }

  .ps-group-badge {
    flex: none;
    font-size: 10px;
    font-weight: 600;
    color: var(--text-3, #8b949e);
    background: rgba(0, 0, 0, 0.05);
    padding: 1px 6px;
    border-radius: 8px;
    min-width: 18px;
    text-align: center;
  }

  .ps-group-body {
    padding-left: 12px;
  }

  .ps-empty {
    padding: 20px 14px;
    text-align: center;
    color: var(--text-3, #8b949e);
    font-size: 12px;

    &.slim {
      padding: 10px 14px;
    }
  }

  .ps-foot {
    display: flex;
    align-items: center;
    gap: 5px;
    padding: 8px 14px;
    border-top: 1px solid rgba(0, 0, 0, 0.04);
    font-size: 12px;
    color: var(--text-3, #8b949e);
    cursor: pointer;
    transition: all 0.15s;

    &:hover {
      color: var(--primary-color, #16a34a);
      background: rgba(22, 163, 74, 0.03);

      .ps-foot-arrow {
        color: var(--primary-color, #16a34a);
        transform: translateX(1px);
      }
    }

    .ps-foot-arrow {
      margin-left: auto;
      font-size: 14px;
      transition: all 0.15s;
    }
  }
</style>
