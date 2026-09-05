<template>
  <div>
    <!-- 项目信息 -->
    <n-card :bordered="false" class="proCard">
      <n-spin :show="loading">
        <template v-if="project">
          <n-space align="center" :size="12" class="mb-4">
            <span class="text-xl font-semibold">{{ project.name }}</span>
            <n-tag :type="PROJECT_STATUS.tagType(project.status)" size="small">
              {{ PROJECT_STATUS.label(project.status) }}
            </n-tag>
          </n-space>
          <n-descriptions label-placement="left" :column="3" size="small">
            <n-descriptions-item label="创建人">{{ project.creatorName || '-' }}</n-descriptions-item>
            <n-descriptions-item label="创建时间">{{ project.createdAt }}</n-descriptions-item>
            <n-descriptions-item label="描述" :span="3">
              {{ project.description || '暂无描述' }}
            </n-descriptions-item>
          </n-descriptions>
        </template>
        <EmptyState type="generic" title="未找到项目信息" v-else-if="!loading" />
      </n-spin>
    </n-card>

    <!-- 统计行 -->
    <n-grid :x-gap="12" :y-gap="12" :cols="5" class="mt-3" responsive="screen" item-responsive>
      <n-grid-item span="5 m:1">
        <n-card :bordered="false" size="small">
          <n-statistic label="任务总数" :value="taskStats.total" />
        </n-card>
      </n-grid-item>
      <n-grid-item span="5 m:1">
        <n-card :bordered="false" size="small">
          <n-statistic label="待处理" :value="taskStats.open">
            <template #suffix><span class="text-xs text-gray-400">open</span></template>
          </n-statistic>
        </n-card>
      </n-grid-item>
      <n-grid-item span="5 m:1">
        <n-card :bordered="false" size="small">
          <n-statistic label="进行中" :value="taskStats.inProgress">
            <template #suffix><span class="text-xs text-gray-400">in_progress</span></template>
          </n-statistic>
        </n-card>
      </n-grid-item>
      <n-grid-item span="5 m:1">
        <n-card :bordered="false" size="small">
          <n-statistic label="审核中" :value="taskStats.review">
            <template #suffix><span class="text-xs text-gray-400">review</span></template>
          </n-statistic>
        </n-card>
      </n-grid-item>
      <n-grid-item span="5 m:1">
        <n-card :bordered="false" size="small">
          <n-statistic label="已完成" :value="taskStats.done">
            <template #suffix><span class="text-xs text-gray-400">done</span></template>
          </n-statistic>
        </n-card>
      </n-grid-item>
    </n-grid>

    <!-- 中部：当前 Sprint + 最近动态 -->
    <n-grid :x-gap="12" :y-gap="12" :cols="3" class="mt-3" responsive="screen" item-responsive>
      <n-grid-item span="3 m:1">
        <n-card :bordered="false" size="small" title="当前 Sprint" class="h-full">
          <template v-if="activeSprint">
            <n-space vertical :size="8">
              <n-space align="center" :size="8">
                <span class="font-medium">{{ activeSprint.name }}</span>
                <n-tag :type="SPRINT_STATUS.tagType(activeSprint.status)" size="small">
                  {{ SPRINT_STATUS.label(activeSprint.status) }}
                </n-tag>
              </n-space>
              <div class="text-xs text-gray-400">
                {{ day(activeSprint.startDate) }} ~ {{ day(activeSprint.endDate) }}
              </div>
              <n-progress
                type="line"
                :percentage="sprintProgress"
                :height="10"
                :show-indicator="false"
                color="#18a058"
              />
              <div class="text-xs text-gray-500">
                冲刺任务完成 {{ sprintDone }}/{{ sprintTotal }}（{{ sprintProgress }}%）
              </div>
              <div class="text-xs text-gray-400">目标：{{ activeSprint.goal || '未设定' }}</div>
            </n-space>
          </template>
          <EmptyState type="task" title="没有进行中的 Sprint" v-else compact />
        </n-card>
      </n-grid-item>

      <n-grid-item span="3 m:2">
        <n-card :bordered="false" size="small" title="最近动态" class="h-full">
          <EmptyState type="notify" title="暂无动态" v-if="activities.length === 0" compact />
          <n-space v-else vertical :size="10">
            <n-space v-for="a in activities" :key="a.id" align="center" :size="8">
              <n-tag size="tiny" :type="a.actorType === 'ai' ? 'warning' : 'info'" :bordered="false">
                {{ a.actorType === 'ai' ? 'Agent' : '用户' }}
              </n-tag>
              <span class="text-sm">
                <span class="font-medium">{{ a.actorName || '系统' }}</span>
                {{ actionText(a.action) }}
                <span class="text-gray-500">{{ a.targetName || a.targetType }}</span>
              </span>
              <span class="text-xs text-gray-400 ml-auto">{{ a.createdAt }}</span>
            </n-space>
          </n-space>
        </n-card>
      </n-grid-item>
    </n-grid>

    <!-- 底部：需求与成员 -->
    <n-grid :x-gap="12" :y-gap="12" :cols="3" class="mt-3" responsive="screen" item-responsive>
      <n-grid-item span="3 m:1">
        <n-card :bordered="false" size="small" title="需求概览" class="h-full">
          <n-space vertical :size="6">
            <n-statistic label="需求总数" :value="reqTotal" />
            <n-space :size="6">
              <n-tag
                v-for="c in reqCounts"
                :key="c.label"
                size="small"
                :type="REQ_STATUS.tagType(c.value)"
                :bordered="false"
              >{{ c.label }} {{ c.count }}</n-tag>
            </n-space>
          </n-space>
        </n-card>
      </n-grid-item>

      <n-grid-item span="3 m:2">
        <n-card :bordered="false" size="small" class="h-full">
          <template #header>
            成员
            <span class="text-xs text-gray-400">（{{ members.length }}）</span>
          </template>
          <EmptyState type="member" title="暂无成员" v-if="members.length === 0" compact />
          <n-space v-else :size="8">
            <n-tag
              v-for="m in members"
              :key="m.userId"
              size="small"
              round
              :bordered="false"
              :type="m.role === 'owner' ? 'success' : 'default'"
            >
              {{ m.realName || m.username }}{{ m.role === 'owner' ? ' · owner' : '' }}
            </n-tag>
          </n-space>
        </n-card>
      </n-grid-item>
    </n-grid>
  </div>
</template>

<script lang="ts" setup>
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { PROJECT_STATUS, SPRINT_STATUS, REQ_STATUS } from '@/enums/entities';
  import { ref, computed, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { getProject, getTasks, getSprints, getMembers, getRequirements } from '@/api/project/index';
  import type { ProjectItem, TaskItem, SprintItem, MemberItem } from '@/api/project/index';
  import { getActivities } from '@/api/platform/index';

  const route = useRoute();
  const projectId = computed(() => Number(route.params.projectId));
  const loading = ref(false);
  const project = ref<ProjectItem | null>(null);
  const tasks = ref<TaskItem[]>([]);
  const sprints = ref<SprintItem[]>([]);
  const members = ref<MemberItem[]>([]);
  const reqTotal = ref(0);
  const reqCounts = ref<Array<{ label: string; value: string; count: number }>>([]);
  const activities = ref<any[]>([]);

  const taskStats = computed(() => {
    const s = { total: tasks.value.length, open: 0, inProgress: 0, review: 0, done: 0 };
    for (const t of tasks.value) {
      if (t.status === 'open') s.open++;
      else if (t.status === 'in_progress') s.inProgress++;
      else if (t.status === 'review') s.review++;
      else if (t.status === 'done' || t.status === 'closed') s.done++;
    }
    return s;
  });

  // 进行中的 Sprint 优先，否则最近一个
  const activeSprint = computed(() => {
    const list = sprints.value;
    return list.find((s) => s.status === 'active') || list[0] || null;
  });

  const sprintTasks = computed(() =>
    activeSprint.value ? tasks.value.filter((t) => t.sprintId === activeSprint.value!.id) : []
  );
  const sprintTotal = computed(() => sprintTasks.value.length);
  const sprintDone = computed(
    () => sprintTasks.value.filter((t) => t.status === 'done' || t.status === 'closed').length
  );
  const sprintProgress = computed(() =>
    sprintTotal.value === 0 ? 0 : Math.round((sprintDone.value / sprintTotal.value) * 100)
  );

  function day(v: string): string {
    return (v || '').slice(0, 10);
  }

  const actionLabels: Record<string, string> = {
    create: '创建了', update: '更新了', delete: '删除了',
    'project.created': '创建了项目',
    'task.created': '创建了任务', 'task.claimed': '认领了任务', 'task.completed': '完成了任务',
    'sprint.created': '创建了 Sprint', 'sprint.deleted': '删除了 Sprint',
    'test_plan.created': '创建了测试计划', 'test_plan.updated': '更新了测试计划',
    'test_plan.deleted': '删除了测试计划', 'test_plan.added_cases': '向计划添加了用例',
    'test_case.created': '创建了测试用例', 'test_case.deleted': '删除了测试用例',
    'test_case.executed': '执行了测试用例',
    'attachment.uploaded': '上传了附件', 'attachment.deleted': '删除了附件',
  };
  function actionText(a: string): string {
    return actionLabels[a] || `执行了 ${a}`;
  }

  onMounted(async () => {
    loading.value = true;
    const pid = projectId.value;
    if (!pid) return;
    // 各数据源并行拉取，单项失败不拖垮整页
    const safe = <T,>(p: Promise<T>, fallback: T): Promise<T> => p.catch(() => fallback);
    const [p, t, sp, mb, rq, ac] = await Promise.all([
      safe(getProject(pid), null),
      safe(getTasks(pid, { size: 200 }), { list: [] as TaskItem[] }),
      safe(getSprints(pid), { list: [] as SprintItem[] }),
      safe(getMembers(pid), { list: [] as MemberItem[] }),
      safe(getRequirements(pid, { page: 1, size: 200 }), { list: [], total: 0 }),
      safe(getActivities({ projectId: pid, page: 1, size: 10 }), { list: [] }),
    ]);
    project.value = p || null;
    tasks.value = (t as any)?.list || [];
    sprints.value = (sp as any)?.list || [];
    members.value = (mb as any)?.list || [];
    reqTotal.value = (rq as any)?.total || (rq as any)?.list?.length || 0;
    // 需求状态计数（客户端聚合，量级允许）
    const rc = new Map<string, number>();
    for (const r of (rq as any)?.list || []) {
      rc.set(r.status, (rc.get(r.status) || 0) + 1);
    }
    reqCounts.value = [...rc.entries()].map(([value, count]) => ({
      label: REQ_STATUS.label(value), value, count,
    }));
    activities.value = (ac as any)?.list || [];
    loading.value = false;
  });
</script>
