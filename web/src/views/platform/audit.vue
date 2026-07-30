<template>
  <div>
    <div class="n-layout-page-header">
      <n-card :bordered="false" title="审计日志" />
    </div>

    <n-card :bordered="false" class="mt-4 proCard">
      <!-- 筛选条件 -->
      <n-space class="mb-4" align="center">
        <n-input
          v-model:value="filters.action"
          placeholder="操作类型"
          style="width: 160px"
          clearable
          @keyup.enter="loadData"
        />
        <n-select
          v-model:value="filters.targetType"
          :options="targetTypeOptions"
          placeholder="目标类型"
          style="width: 160px"
          clearable
          @update:value="loadData"
        />
        <n-button type="primary" @click="loadData">查询</n-button>
      </n-space>

      <n-spin :show="loading">
        <n-empty v-if="!loading && logList.length === 0" description="暂无审计日志" />
        <n-table v-else :bordered="false" :single-line="false" size="small">
          <thead>
            <tr>
              <th>操作</th>
              <th>目标类型</th>
              <th>目标名称</th>
              <th>变更内容</th>
              <th>IP</th>
              <th>时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in logList" :key="item.id">
              <td>
                <n-tag size="small">{{ item.action }}</n-tag>
              </td>
              <td>{{ item.targetType }}</td>
              <td>{{ item.targetName }}</td>
              <td>
                <n-ellipsis style="max-width: 200px">{{ item.changes }}</n-ellipsis>
              </td>
              <td class="text-xs text-gray-400">{{ item.ipAddress }}</td>
              <td>{{ item.createdAt }}</td>
            </tr>
          </tbody>
        </n-table>
      </n-spin>

      <div class="mt-4 flex justify-end" v-if="total > pagination.size">
        <n-pagination
          v-model:page="pagination.page"
          :page-size="pagination.size"
          :item-count="total"
          @update:page="loadData"
        />
      </div>
    </n-card>
  </div>
</template>

<script lang="ts" setup>
  import { ref, reactive, onMounted } from 'vue';
  import { getAuditLogs } from '@/api/platform/index';
  import type { AuditLogItem } from '@/api/platform/index';

  const loading = ref(false);
  const logList = ref<AuditLogItem[]>([]);
  const total = ref(0);
  const pagination = reactive({ page: 1, size: 20 });
  const filters = reactive({
    action: '',
    targetType: null as string | null,
  });

  const targetTypeOptions = [
    { label: '需求', value: 'requirement' },
    { label: '项目', value: 'project' },
    { label: '任务', value: 'task' },
    { label: '测试用例', value: 'test_case' },
    { label: '文档', value: 'doc' },
    { label: 'AI 用户', value: 'ai_user' },
  ];

  async function loadData() {
    loading.value = true;
    try {
      const res = await getAuditLogs({
        action: filters.action || undefined,
        targetType: filters.targetType ?? undefined,
        page: pagination.page,
        size: pagination.size,
      });
      if (res) {
        logList.value = res.list || [];
        total.value = res.total || 0;
      }
    } catch (e) {
      // ignore
    } finally {
      loading.value = false;
    }
  }

  onMounted(() => {
    loadData();
  });
</script>
