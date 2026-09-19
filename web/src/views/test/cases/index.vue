<template>
  <div>

    <n-card :bordered="false" class="proCard">
      <!-- 筛选栏 -->
      <n-space class="mb-4" align="center">
        <n-select
          v-model:value="filters.category"
          :options="categoryOptions"
          placeholder="分类"
          style="width: 140px"
          clearable
          @update:value="onFilterChange"
        />
        <n-select
          v-model:value="filters.status"
          :options="statusOptions"
          placeholder="状态"
          style="width: 140px"
          clearable
          @update:value="onFilterChange"
        />
        <n-input
          v-model:value="filters.keyword"
          placeholder="搜索用例"
          style="width: 200px"
          clearable
          data-test-id="test-cases.search-input"
          @keyup.enter="onFilterChange"
        />
        <n-select
          v-model:value="sortBy"
          :options="sortOptions"
          size="small"
          style="width: 130px"
          data-test-id="test-cases.sort"
        />
        <n-button size="small" @click="sortDesc = !sortDesc" data-test-id="test-cases.sort-dir">
          <template #icon>
            <n-icon><SortDescendingOutlined v-if="sortDesc" /><SortAscendingOutlined v-else /></n-icon>
          </template>
          {{ sortDesc ? '降序' : '升序' }}
        </n-button>
        <n-button type="primary" @click="handleCreate" data-test-id="test-cases.create-btn">
          <template #icon>
            <n-icon><PlusOutlined /></n-icon>
          </template>
          新建用例
        </n-button>
      </n-space>

      <n-spin :show="loading">
        <EmptyState
          type="doc"
          title="暂无测试用例"
          description="创建用例沉淀测试资产"
          v-if="!loading && groupedCases.length === 0"
        />
        <n-table v-else :bordered="false" :single-line="false" size="small">
          <thead>
            <tr>
              <th class="col-idx">#</th>
              <th>标题</th>
              <th>分类</th>
              <th>优先级</th>
              <th>状态</th>
              <th>外部键</th>
              <th>创建人</th>
              <th>更新时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <template v-for="g in groupedCases" :key="g.module">
              <!-- 模块组头：稳定派生色条 + 计数 + 折叠 -->
              <tr
                class="module-head"
                :data-test-id="`test-cases.group-${g.module}`"
                @click="toggleGroup(g.module)"
              >
                <td colspan="9">
                  <n-space size="small" align="center">
                    <span class="module-dot" :style="{ background: moduleColor(g.module) }"></span>
                    <span class="font-medium">{{ g.module }}</span>
                    <n-tag size="small" :bordered="false">{{ g.cases.length }} 条</n-tag>
                    <span class="text-xs text-gray-400">
                      {{ collapsed.has(g.module) ? '展开' : '收起' }}
                    </span>
                  </n-space>
                </td>
              </tr>
              <template v-if="!collapsed.has(g.module)">
                <tr v-for="(item, __ix) in g.cases" :key="item.id" :data-test-id="`test-cases.row-${item.id}`">
                  <td class="col-idx">{{ __ix + 1 }}</td>
                  <td>
                    <span :title="item.externalKey">{{ item.title }}</span>
                  </td>
                  <td>{{ categoryLabel(item.category) }}</td>
                  <td>
                    <n-tag :type="casePriorityTagType(item.priority)" size="small">{{ item.priority }}</n-tag>
                  </td>
                  <td>
                    <n-tag :type="CASE_STATUS.tagType(item.status)" size="small">
                      {{ CASE_STATUS.label(item.status) }}
                    </n-tag>
                  </td>
                  <td class="ext-key">{{ item.externalKey || '-' }}</td>
                  <td>{{ item.creatorName || '-' }}</td>
                  <td>{{ item.updatedAt }}</td>
                  <td>
                    <n-space size="small">
                      <n-button text type="info" @click="handleEdit(item)">编辑</n-button>
                      <n-button text type="error" @click="handleDelete(item)">删除</n-button>
                    </n-space>
                  </td>
                </tr>
              </template>
            </template>
          </tbody>
        </n-table>
      </n-spin>
    </n-card>

    <!-- 新建/编辑弹窗 -->
    <n-modal
      v-model:show="showModal"
      preset="dialog"
      :title="isEdit ? '编辑用例' : '新建用例'"
      positive-text="确定"
      negative-text="取消"
      @positive-click="handleSubmit"
      style="width: 600px"
    >
      <n-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-placement="left"
        :label-width="100"
        class="py-4"
      >
        <n-form-item label="标题" path="title">
          <n-input v-model:value="formData.title" placeholder="请输入用例标题" />
        </n-form-item>
        <n-form-item label="分类" path="category">
          <n-input v-model:value="formData.category" placeholder="请输入分类" />
        </n-form-item>
        <n-form-item label="模块" path="module">
          <n-input v-model:value="formData.module" placeholder="请输入模块" />
        </n-form-item>
        <n-form-item label="优先级" path="priority">
          <n-select v-model:value="formData.priority" :options="priorityOptions" placeholder="请选择优先级" />
        </n-form-item>
        <n-form-item label="前置条件" path="preconditions">
          <n-input v-model:value="formData.preconditions" type="textarea" placeholder="请输入前置条件" :rows="2" />
        </n-form-item>
        <n-form-item label="步骤" path="steps">
          <n-input v-model:value="formData.steps" type="textarea" placeholder="请输入测试步骤" :rows="3" />
        </n-form-item>
        <n-form-item label="预期结果" path="expectedResult">
          <n-input v-model:value="formData.expectedResult" type="textarea" placeholder="请输入预期结果" :rows="2" />
        </n-form-item>
      </n-form>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { CASE_STATUS, CASE_CATEGORY_OPTIONS, casePriorityTagType } from '@/enums/test';
  import { ref, reactive, onMounted, computed } from 'vue';
  import { useRoute } from 'vue-router';
  import { useMessage, useDialog } from 'naive-ui';
  import { PlusOutlined } from '@vicons/antd';
  import {
    getTestCases,
    createTestCase,
    updateTestCase,
    deleteTestCase,
  } from '@/api/test/index';
  import type { TestCaseItem, TestCaseCreateData, TestCaseUpdateData } from '@/api/test/index';

  const message = useMessage();
  const dialog = useDialog();
  const showModal = ref(false);
  const isEdit = ref(false);
  const editId = ref<number | null>(null);
  const formRef = ref<any>(null);

  const route = useRoute();
  const projectId = computed(() => Number(route.params.projectId));
  const filters = reactive({ category: null as string | null, status: null as string | null, keyword: '' });

  function categoryLabel(c: string) {
    return categoryOptions.find((o) => o.value === c)?.label || c;
  }

  // 字典统一出口：enums/test.ts（用例状态/分类；优先级 P0-P3 字符串域）
  const categoryOptions = CASE_CATEGORY_OPTIONS;
  const statusOptions = CASE_STATUS.options;

  const priorityOptions = [
    { label: 'P0 - 阻塞', value: 'P0' },
    { label: 'P1 - 高', value: 'P1' },
    { label: 'P2 - 中', value: 'P2' },
    { label: 'P3 - 低', value: 'P3' },
  ];

  const formData = reactive<TestCaseCreateData & { module?: string; expectedResult?: string; preconditions?: string; steps?: string }>({
    title: '',
    category: '',
    module: '',
    priority: 'P2',
    preconditions: '',
    steps: '',
    expectedResult: '',
  });

  const formRules = {
    title: { required: true, message: '请输入用例标题', trigger: 'blur' },
    priority: { required: true, message: '请选择优先级', trigger: 'change' },
  };



  function resetForm() {
    formData.title = '';
    formData.category = '';
    formData.module = '';
    formData.priority = 'P2';
    formData.preconditions = '';
    formData.steps = '';
    formData.expectedResult = '';
    isEdit.value = false;
    editId.value = null;
  }

    // 筛选变更从第 1 页重查：第 N 页改筛选会请求空页显示"暂无"


// 整页直出（几百条量级分页碍事）：大 pageSize 一次拉全，管理动作在客户端做
  const loading = ref(false);
  const caseList = ref<TestCaseItem[]>([]);
  const CASES_PAGE_SIZE = 999;
  async function loadData() {
    loading.value = true;
    try {
      const res = await getTestCases(projectId.value, {
        category: filters.category ?? undefined,
        status: filters.status ?? undefined,
        keyword: filters.keyword || undefined,
        pageNum: 1,
        pageSize: CASES_PAGE_SIZE,
      });
      caseList.value = res?.list || [];
    } catch (e) {
      // ignore
    } finally {
      loading.value = false;
    }
  }
  function onFilterChange() {
    loadData();
  }
  function afterRemove() {
    loadData();
  }

  // ---------- 排序与模块分组（客户端） ----------
  const sortBy = ref<'priority' | 'status' | 'updatedAt' | 'title'>('priority');
  const sortDesc = ref(false);
  const collapsed = ref(new Set<string>());
  const sortOptions = [
    { label: '按优先级', value: 'priority' },
    { label: '按状态', value: 'status' },
    { label: '按更新时间', value: 'updatedAt' },
    { label: '按标题', value: 'title' },
  ];
  const PRIO_RANK: Record<string, number> = { P0: 0, P1: 1, P2: 2, P3: 3 };
  const STATUS_RANK: Record<string, number> = { active: 0, deprecated: 1 };

  function cmp(a: TestCaseItem, b: TestCaseItem): number {
    let d = 0;
    if (sortBy.value === 'priority') {
      d = (PRIO_RANK[a.priority] ?? 9) - (PRIO_RANK[b.priority] ?? 9);
    } else if (sortBy.value === 'status') {
      d = (STATUS_RANK[a.status] ?? 9) - (STATUS_RANK[b.status] ?? 9);
    } else if (sortBy.value === 'updatedAt') {
      d = (a.updatedAt || '').localeCompare(b.updatedAt || '');
    } else {
      d = a.title.localeCompare(b.title, 'zh');
    }
    if (d === 0) d = a.title.localeCompare(b.title, 'zh'); // 稳定次键
    return sortDesc.value ? -d : d;
  }

  const groupedCases = computed(() => {
    const groups = new Map<string, TestCaseItem[]>();
    for (const c of [...caseList.value].sort(cmp)) {
      const key = c.module?.trim() || '未分类';
      if (!groups.has(key)) groups.set(key, []);
      groups.get(key)!.push(c);
    }
    return [...groups.entries()]
      .sort((x, y) => x[0].localeCompare(y[0], 'zh'))
      .map(([module, cases]) => ({ module, cases }));
  });

  function toggleGroup(module: string) {
    if (collapsed.value.has(module)) {
      collapsed.value.delete(module);
    } else {
      collapsed.value.add(module);
    }
    collapsed.value = new Set(collapsed.value);
  }

  // 模块稳定派生色（8 色板按名字 hash，同名恒色）
  const MODULE_PALETTE = [
    '#2080f0', '#18a058', '#f0a020', '#d03050',
    '#7c29cd', '#0fa7a0', '#c26d1a', '#5a8b29',
  ];
  function moduleColor(module: string): string {
    let h = 0;
    for (const ch of module) h = (h * 31 + ch.codePointAt(0)!) >>> 0;
    return MODULE_PALETTE[h % MODULE_PALETTE.length];
  }

  function handleCreate() {
    resetForm();
    showModal.value = true;
  }

  function handleEdit(item: TestCaseItem) {
    isEdit.value = true;
    editId.value = item.id;
    formData.title = item.title;
    formData.category = item.category;
    formData.module = item.module;
    formData.priority = item.priority;
    formData.preconditions = item.preconditions;
    formData.steps = item.steps;
    formData.expectedResult = item.expectedResult;
    showModal.value = true;
  }

  function handleDelete(item: TestCaseItem) {
    dialog.warning({
      title: '确认删除',
      content: `确定要删除用例「${item.title}」吗？`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await deleteTestCase(item.id);
          message.success('删除成功');
          afterRemove();
        } catch (e) {
          message.error('删除失败');
        }
      },
    });
  }

  async function handleSubmit() {
    try {
      await formRef.value?.validate();
    } catch {
      return false;
    }
    try {
      if (isEdit.value && editId.value) {
        await updateTestCase(editId.value, { ...formData });
        message.success('更新成功');
      } else {
        await createTestCase(projectId.value, { ...formData });
        message.success('创建成功');
      }
      showModal.value = false;
      loadData();
    } catch (e) {
      message.error('操作失败');
      return false;
    }
  }

  onMounted(() => {
    loadData();
  });
</script>

<style scoped>
  .module-head {
    cursor: pointer;
    background: rgba(128, 128, 128, 0.06);
  }
  .module-head:hover {
    background: rgba(128, 128, 128, 0.12);
  }
  .module-dot {
    display: inline-block;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    flex-shrink: 0;
  }
  .ext-key {
    max-width: 260px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-family: 'JetBrains Mono', Consolas, monospace;
    font-size: 11px;
    color: rgba(128, 128, 128, 0.9);
  }
</style>
