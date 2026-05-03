<template>
  <div class="console">
    <!-- 统计卡片 -->
    <n-grid cols="1 s:2 m:2 l:4 xl:5 2xl:5" responsive="screen" :x-gap="12" :y-gap="8">
      <n-grid-item>
        <n-card title="总需求数" size="small" :bordered="false">
          <template #header-extra>
            <n-icon size="24" color="#69c0ff"><FileTextOutlined /></n-icon>
          </template>
          <n-skeleton v-if="loading" :width="60" size="medium" />
          <CountTo v-else :startVal="0" :endVal="stats.totalRequirements" class="text-3xl" />
          <template #footer>
            <span class="text-gray-400">全部需求统计</span>
          </template>
        </n-card>
      </n-grid-item>

      <n-grid-item>
        <n-card title="总任务数" size="small" :bordered="false">
          <template #header-extra>
            <n-icon size="24" color="#b37feb"><CheckSquareOutlined /></n-icon>
          </template>
          <n-skeleton v-if="loading" :width="60" size="medium" />
          <CountTo v-else :startVal="0" :endVal="stats.totalTasks" class="text-3xl" />
          <template #footer>
            <span class="text-gray-400">全部任务统计</span>
          </template>
        </n-card>
      </n-grid-item>

      <n-grid-item>
        <n-card title="进行中" size="small" :bordered="false">
          <template #header-extra>
            <n-icon size="24" color="#5cdbd3"><SyncOutlined /></n-icon>
          </template>
          <n-skeleton v-if="loading" :width="60" size="medium" />
          <CountTo v-else :startVal="0" :endVal="stats.inProgressTasks" class="text-3xl" />
          <template #footer>
            <span class="text-gray-400">当前进行中的任务</span>
          </template>
        </n-card>
      </n-grid-item>

      <n-grid-item>
        <n-card title="待审核" size="small" :bordered="false">
          <template #header-extra>
            <n-icon size="24" color="#ffc069"><AuditOutlined /></n-icon>
          </template>
          <n-skeleton v-if="loading" :width="60" size="medium" />
          <CountTo v-else :startVal="0" :endVal="stats.reviewTasks" class="text-3xl" />
          <template #footer>
            <span class="text-gray-400">等待审核的任务</span>
          </template>
        </n-card>
      </n-grid-item>

      <n-grid-item>
        <n-card title="测试通过率" size="small" :bordered="false">
          <template #header-extra>
            <n-icon size="24" color="#95de64"><SafetyCertificateOutlined /></n-icon>
          </template>
          <n-skeleton v-if="loading" :width="60" size="medium" />
          <span v-else class="text-3xl">{{ stats.testPassRate }}%</span>
          <template #footer>
            <span class="text-gray-400">自动化测试通过率</span>
          </template>
        </n-card>
      </n-grid-item>
    </n-grid>

    <!-- AI 协作效率排行 -->
    <n-card title="AI 协作效率排行" class="mt-4" :bordered="false">
      <div ref="chartRef" style="height: 300px"></div>
    </n-card>

    <!-- 最近更新的任务 -->
    <n-card title="最近更新的任务" class="mt-4" :bordered="false">
      <n-spin :show="loading">
        <n-empty v-if="!loading && stats.recentTasks.length === 0" description="暂无任务" />
        <n-table v-else :bordered="false" :single-line="false" size="small">
          <thead>
            <tr>
              <th>任务标题</th>
              <th>状态</th>
              <th>更新时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="task in stats.recentTasks" :key="task.id">
              <td>{{ task.title }}</td>
              <td>
                <n-tag :type="statusTagType(task.status)" size="small">
                  {{ statusLabel(task.status) }}
                </n-tag>
              </td>
              <td>{{ task.updatedAt }}</td>
            </tr>
          </tbody>
        </n-table>
      </n-spin>
    </n-card>
  </div>
</template>

<script lang="ts" setup>
  import { ref, reactive, onMounted, nextTick } from 'vue';
  import * as echarts from 'echarts';
  import { getDashboardStats } from '@/api/platform/index';
  import type { DashboardStatsResult, RecentTaskItem, AiStatItem } from '@/api/platform/index';
  import { CountTo } from '@/components/CountTo/index';
  import {
    FileTextOutlined,
    CheckSquareOutlined,
    SyncOutlined,
    AuditOutlined,
    SafetyCertificateOutlined,
  } from '@vicons/antd';

  const loading = ref(true);
  const chartRef = ref<HTMLElement | null>(null);

  const stats = reactive<DashboardStatsResult>({
    totalRequirements: 0,
    totalTasks: 0,
    inProgressTasks: 0,
    reviewTasks: 0,
    testPassRate: 0,
    aiStats: [],
    recentTasks: [],
  });

  const statusMap: Record<string, { label: string; type: string }> = {
    open: { label: '待处理', type: 'default' },
    in_progress: { label: '进行中', type: 'info' },
    review: { label: '审核中', type: 'warning' },
    done: { label: '已完成', type: 'success' },
  };

  function statusLabel(status: string) {
    return statusMap[status]?.label || status;
  }

  function statusTagType(status: string): 'default' | 'info' | 'warning' | 'success' {
    return (statusMap[status]?.type as any) || 'default';
  }

  function initChart(data: AiStatItem[]) {
    if (!chartRef.value) return;
    const chart = echarts.init(chartRef.value);
    chart.setOption({
      tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: { type: 'value' },
      yAxis: {
        type: 'category',
        data: data.map((d) => d.aiName),
      },
      series: [
        {
          type: 'bar',
          data: data.map((d) => d.taskCount),
          itemStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 1, 0, [
              { offset: 0, color: '#1890ff' },
              { offset: 1, color: '#69c0ff' },
            ]),
          },
        },
      ],
    });
    window.addEventListener('resize', () => chart.resize());
  }

  onMounted(async () => {
    try {
      const res = await getDashboardStats();
      if (res) {
        Object.assign(stats, res);
        await nextTick();
        initChart(stats.aiStats);
      }
    } catch (e) {
      // ignore
    } finally {
      loading.value = false;
    }
  });
</script>
