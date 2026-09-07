<template>
  <n-popover trigger="hover" placement="bottom" :width="260" :show-arrow="false" raw>
    <template #trigger>
      <div class="layout-header-trigger layout-header-trigger-min project-switch" title="切换项目">
        <n-icon size="16"><AppstoreOutlined /></n-icon>
        <span v-if="currentName" class="ps-name">{{ currentName }}</span>
        <n-icon size="11" class="ps-caret"><CaretDownOutlined /></n-icon>
      </div>
    </template>

    <div class="ps-panel">
      <div class="ps-head">切换项目</div>
      <n-spin :show="loading" size="small">
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
      </n-spin>
      <div class="ps-foot" @click="goList">管理项目 ›</div>
    </div>
  </n-popover>
</template>

<script lang="ts" setup>
  import { ref, computed, onMounted } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { AppstoreOutlined, CaretDownOutlined } from '@vicons/antd';
  import { getProjects } from '@/api/project/index';

  const route = useRoute();
  const router = useRouter();
  const loading = ref(false);
  const projects = ref<Array<{ id: number; name: string; code: string }>>([]);

  const currentProjectId = computed(() => Number(route.params.projectId) || 0);
  const currentName = computed(
    () => projects.value.find((p) => p.id === currentProjectId.value)?.name || '',
  );

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

  async function load() {
    loading.value = true;
    try {
      const res = await getProjects({ size: 50 });
      projects.value = (res?.list || []).map((p: any) => ({ id: p.id, name: p.name, code: p.code }));
    } catch { /* ignore */ }
    finally { loading.value = false; }
  }

  onMounted(load);
</script>

<style lang="less" scoped>
  .project-switch {
    display: flex;
    align-items: center;
    gap: 5px;
    padding: 0 10px;

    .ps-name {
      font-size: 13px;
      font-weight: 600;
      max-width: 120px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .ps-caret { opacity: 0.45; }
  }
</style>

<style lang="less">
  /* Popover raw 模式：内容不带 naive 默认壳，完全自定义 */
  .ps-panel {
    background: #fff;
    border-radius: 14px;
    border: 1px solid rgba(0, 0, 0, 0.06);
    box-shadow: 0 4px 14px rgba(0, 0, 0, 0.1), 0 12px 36px rgba(0, 0, 0, 0.08);
    overflow: hidden;
    min-width: 240px;
  }
  .ps-head {
    font-size: 11px;
    font-weight: 700;
    color: #9aa1ab;
    letter-spacing: 1px;
    text-transform: uppercase;
    padding: 12px 14px 8px;
  }
  .ps-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 9px 14px;
    cursor: pointer;
    transition: background 0.12s;

    &:hover { background: rgba(22, 163, 74, 0.05); }
    &.cur {
      background: rgba(22, 163, 74, 0.08);
      .ps-item-name { font-weight: 700; color: #16a34a; }
      .ps-check { stroke: #16a34a; }
    }
  }
  .ps-item-name {
    flex: 1;
    min-width: 0;
    font-size: 13.5px;
    color: #1f2329;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .ps-item-code {
    flex: none;
    font-size: 11px;
    font-family: ui-monospace, Consolas, monospace;
    color: #9aa1ab;
    background: rgba(0, 0, 0, 0.04);
    border-radius: 6px;
    padding: 2px 7px;
  }
  .ps-check {
    flex: none;
    width: 15px;
    height: 15px;
  }
  .ps-empty {
    padding: 16px;
    text-align: center;
    font-size: 12px;
    color: #9aa1ab;
  }
  .ps-foot {
    font-size: 12px;
    font-weight: 600;
    color: #16a34a;
    text-align: center;
    padding: 10px;
    border-top: 1px solid rgba(0, 0, 0, 0.05);
    cursor: pointer;
    transition: background 0.12s;

    &:hover { background: rgba(22, 163, 74, 0.05); }
  }
</style>
