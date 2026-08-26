<template>
  <div>
    <n-card :bordered="false" title="测试计划" class="proCard">
      <template #header-extra>
        <n-button type="primary" @click="handleCreate">
          <template #icon>
            <n-icon><PlusOutlined /></n-icon>
          </template>
          新建计划
        </n-button>
      </template>

      <n-spin :show="loading">
        <n-empty v-if="!loading && planList.length === 0" description="暂无测试计划" />
        <n-table v-else :bordered="false" :single-line="false" size="small">
          <thead>
            <tr>
              <th>名称</th>
              <th>描述</th>
              <th>状态</th>
              <th>创建人</th>
              <th>创建时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in planList" :key="item.id">
              <td>
                <n-button text type="info" @click="openPlanDetail(item)">{{ item.name }}</n-button>
              </td>
              <td>{{ item.description }}</td>
              <td>
                <n-tag :type="planStatusType(item.status)" size="small">
                  {{ planStatusLabel(item.status) }}
                </n-tag>
              </td>
              <td>{{ item.creatorName || '-' }}</td>
              <td>{{ item.createdAt }}</td>
              <td>
                <n-space size="small">
                  <n-button text type="info" @click="openPlanDetail(item)">详情</n-button>
                  <n-button text type="error" @click="handleDelete(item)">删除</n-button>
                </n-space>
              </td>
            </tr>
          </tbody>
        </n-table>
      </n-spin>

      <div class="mt-4 flex justify-end" v-if="total > pagination.pageSize">
        <n-pagination
          v-model:page="pagination.page"
          :page-size="pagination.pageSize"
          :item-count="total"
          @update:page="loadData"
        />
      </div>
    </n-card>

    <!-- 新建计划弹窗 -->
    <n-modal
      v-model:show="showCreateModal"
      preset="dialog"
      title="新建测试计划"
      positive-text="确定"
      negative-text="取消"
      @positive-click="handleCreateSubmit"
      style="width: 520px"
    >
      <n-form ref="createFormRef" :model="formData" :rules="formRules" label-placement="left" :label-width="80" class="py-4">
        <n-form-item label="名称" path="name">
          <n-input v-model:value="formData.name" placeholder="请输入计划名称" />
        </n-form-item>
        <n-form-item label="描述" path="description">
          <n-input v-model:value="formData.description" type="textarea" placeholder="请输入描述" :rows="3" />
        </n-form-item>
      </n-form>
    </n-modal>

    <!-- 计划详情弹窗 -->
    <n-modal
      v-model:show="showDetailModal"
      preset="card"
      :title="`测试计划：${currentPlan?.name || ''}`"
      style="width: 800px; max-height: 80vh"
    >
      <n-spin :show="detailLoading">
        <template v-if="planResults">
          <n-space class="mb-4" justify="space-around">
            <n-statistic label="总计" :value="planResults.total" />
            <n-statistic label="通过" :value="planResults.passed">
              <template #suffix><span class="text-green-500 text-xs">passed</span></template>
            </n-statistic>
            <n-statistic label="失败" :value="planResults.failed">
              <template #suffix><span class="text-red-500 text-xs">failed</span></template>
            </n-statistic>
            <n-statistic label="阻塞" :value="planResults.blocked">
              <template #suffix><span class="text-yellow-500 text-xs">blocked</span></template>
            </n-statistic>
          </n-space>
          <n-table :bordered="false" :single-line="false" size="small">
            <thead>
              <tr>
                <th>用例</th>
                <th>指派人</th>
                <th>结果</th>
                <th>实际结果</th>
                <th>执行时间</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="r in planResults.results" :key="r.id">
                <td>{{ r.testCaseTitle }}</td>
                <td>{{ r.assigneeName || '-' }}</td>
                <td>
                  <n-tag :type="resultType(r.status)" size="small">{{ r.status }}</n-tag>
                </td>
                <td>{{ r.actualResult || '-' }}</td>
                <td>{{ r.executedAt || '-' }}</td>
              </tr>
            </tbody>
          </n-table>
        </template>
        <n-empty v-else-if="!detailLoading" description="暂无执行结果" />
      </n-spin>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import { ref, reactive, onMounted } from 'vue';
  import { useMessage, useDialog } from 'naive-ui';
  import { PlusOutlined } from '@vicons/antd';
  import {
    getTestPlans,
    createTestPlan,
    deleteTestPlan,
    getTestPlanResults,
  } from '@/api/test/index';
  import type { TestPlanItem, TestPlanResultsResult } from '@/api/test/index';

  const message = useMessage();
  const dialog = useDialog();
  const loading = ref(false);
  const detailLoading = ref(false);
  const planList = ref<TestPlanItem[]>([]);
  const total = ref(0);
  const showCreateModal = ref(false);
  const showDetailModal = ref(false);
  const currentPlan = ref<TestPlanItem | null>(null);
  const planResults = ref<TestPlanResultsResult | null>(null);
  const createFormRef = ref<any>(null);

  const projectId = ref(1);
  const pagination = reactive({ page: 1, pageSize: 10 });

  const formData = reactive({ name: '', description: '' });
  const formRules = {
    name: { required: true, message: '请输入计划名称', trigger: 'blur' },
  };

  function planStatusType(s: string): 'default' | 'info' | 'success' {
    if (s === 'active') return 'info';
    if (s === 'completed') return 'success';
    return 'default';
  }

  function planStatusLabel(s: string) {
    const m: Record<string, string> = { planning: '规划中', active: '进行中', completed: '已完成' };
    return m[s] || s;
  }

  function resultType(s: string): 'success' | 'error' | 'warning' | 'default' {
    if (s === 'pass') return 'success';
    if (s === 'fail') return 'error';
    if (s === 'blocked') return 'warning';
    return 'default';
  }

  async function loadData() {
    loading.value = true;
    try {
      const res = await getTestPlans(projectId.value, {
        pageNum: pagination.page,
        pageSize: pagination.pageSize,
      });
      if (res) {
        planList.value = res.list || [];
        total.value = res.total || 0;
      }
    } catch (e) {
      // ignore
    } finally {
      loading.value = false;
    }
  }

  function handleCreate() {
    formData.name = '';
    formData.description = '';
    showCreateModal.value = true;
  }

  async function handleCreateSubmit() {
    try {
      await createFormRef.value?.validate();
    } catch {
      return false;
    }
    try {
      await createTestPlan(projectId.value, { ...formData });
      message.success('创建成功');
      showCreateModal.value = false;
      loadData();
    } catch (e) {
      message.error('创建失败');
      return false;
    }
  }

  async function openPlanDetail(plan: TestPlanItem) {
    currentPlan.value = plan;
    showDetailModal.value = true;
    detailLoading.value = true;
    try {
      const res = await getTestPlanResults(plan.id);
      planResults.value = res || null;
    } catch (e) {
      // ignore
    } finally {
      detailLoading.value = false;
    }
  }

  function handleDelete(item: TestPlanItem) {
    dialog.warning({
      title: '确认删除',
      content: `确定要删除计划「${item.name}」吗？`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await deleteTestPlan(item.id);
          message.success('删除成功');
          loadData();
        } catch (e) {
          message.error('删除失败');
        }
      },
    });
  }

  onMounted(() => {
    loadData();
  });
</script>
