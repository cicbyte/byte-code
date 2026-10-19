<template>
  <div class="pro-table">
    <div class="pro-toolbar">
      <div class="pro-toolbar-left">
        <slot name="toolbar" />
      </div>
      <div class="pro-toolbar-right">
        <n-tooltip trigger="hover">
          <template #trigger>
            <span class="pro-tool-btn" data-test-id="protable.refresh-btn" @click="reload">
              <n-icon size="16"><ReloadOutlined /></n-icon>
            </span>
          </template>
          刷新
        </n-tooltip>
        <n-dropdown trigger="click" :options="densityOptions" @select="onDensity">
          <n-tooltip trigger="hover">
            <template #trigger>
              <span class="pro-tool-btn" data-test-id="protable.density-btn">
                <n-icon size="16"><ColumnHeightOutlined /></n-icon>
              </span>
            </template>
            密度
          </n-tooltip>
        </n-dropdown>
        <n-popover trigger="click" placement="bottom-end">
          <template #trigger>
            <n-tooltip trigger="hover">
              <template #trigger>
                <span class="pro-tool-btn" data-test-id="protable.columns-btn">
                  <n-icon size="16"><SettingOutlined /></n-icon>
                </span>
              </template>
              列设置
            </n-tooltip>
          </template>
          <div class="pro-col-setting">
            <div class="pro-col-head">
              <n-checkbox
                :checked="hiddenKeys.length === 0"
                :indeterminate="hiddenKeys.length > 0 && hiddenKeys.length < settableKeys.length"
                @update:checked="onCheckAll"
              >
                列显示
              </n-checkbox>
              <a class="pro-col-reset" @click="resetColumns">重置</a>
            </div>
            <n-checkbox-group :value="visibleKeys" @update:value="onVisibleChange">
              <div v-for="k in settableKeys" :key="k" class="pro-col-item">
                <n-checkbox :value="k" :label="String(colTitle(k))" />
              </div>
            </n-checkbox-group>
          </div>
        </n-popover>
      </div>
    </div>

    <n-data-table
      v-bind="$attrs"
      :columns="visibleColumns"
      :data="list"
      :loading="loading"
      :pagination="paginationOn"
      :size="size"
      :striped="striped"
      :row-key="resolvedRowKey"
      :row-props="rowProps"
      remote
      @update:page="onPage"
      @update:page-size="onPageSize"
    >
      <template v-for="(_, name) in $slots" :key="String(name)" #[name]="slotProps">
        <slot :name="name" v-bind="slotProps ?? {}" />
      </template>
    </n-data-table>
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, onMounted } from 'vue';
  import type { DataTableColumns, DataTableBaseColumn } from 'naive-ui';
  import { ReloadOutlined, ColumnHeightOutlined, SettingOutlined } from '@vicons/antd';

  /**
   * ProTable：n-data-table 的轻封装（后台管理列表页统一入口）。
   * - request({page,pageSize}) → {list,itemCount} 托管加载与远端分页
   * - 工具栏：刷新 / 密度（紧凑-默认-宽松）/ 列显示设置；按 tableKey 持久化
   *   （localStorage：protable:{key}:hidden / density）
   * - 列定义沿用 naive DataTableColumns，可加 hideable:false（如操作列不参与隐藏）
   * - 其余 props/slots 透传 n-data-table（row-props/max-height/empty 等）
   */
  export type ProColumn<T = any> = DataTableBaseColumn<T> & { hideable?: boolean };

  const props = withDefaults(
    defineProps<{
      /** 列配置持久化键（页面内唯一） */
      tableKey: string;
      columns: DataTableColumns<any>;
      /** 远端请求：返回 {list,itemCount} */
      request: (p: { page: number; pageSize: number }) => Promise<{ list: any[]; itemCount: number }>;
      rowKey?: (row: any) => number | string;
      /** 行属性（行点击/锚点等）；显式直绑，最常用透传不走 $attrs */
      rowProps?: (row: any, index: number) => Record<string, unknown>;
      pageSize?: number;
      striped?: boolean;
      pagination?: boolean;
    }>(),
    { rowKey: (row: any) => row.id, pageSize: 20, striped: false, pagination: true }
  );

  const list = ref<any[]>([]);
  const itemCount = ref(0);
  const page = ref(1);
  const pageSize = ref(props.pageSize);
  const loading = ref(false);

  async function fetch() {
    loading.value = true;
    try {
      if (!props.pagination) {
        const r = await props.request({ page: 1, pageSize: 10_000 });
        list.value = r.list || [];
        itemCount.value = r.itemCount || 0;
      } else {
        const r = await props.request({ page: page.value, pageSize: pageSize.value });
        list.value = r.list || [];
        itemCount.value = r.itemCount || 0;
      }
    } finally {
      loading.value = false;
    }
  }

  /** 重置到第一页并请求（筛选变化/新建后调用） */
  function reload() {
    page.value = 1;
    fetch();
  }

  function onPage(p: number) {
    page.value = p;
    fetch();
  }
  function onPageSize(s: number) {
    pageSize.value = s;
    page.value = 1;
    fetch();
  }

  defineExpose({ reload, refresh: fetch });

  onMounted(fetch);

  // naive 的 row-key 只接受函数：字符串键名（如 "id"）归一化为取值函数，
  // 否则行渲染静默失败（表头在、数据行零渲染）
  const resolvedRowKey = computed(() =>
    typeof props.rowKey === 'function' ? props.rowKey : (row: any) => row[props.rowKey as string]
  );

  // ---------- 密度（持久化） ----------
  const densityOptions = [
    { label: '紧凑', key: 'small' },
    { label: '默认', key: 'medium' },
    { label: '宽松', key: 'large' },
  ];
  const size = ref<'small' | 'medium' | 'large'>(
    (localStorage.getItem(`protable:${props.tableKey}:density`) as any) || 'medium'
  );
  function onDensity(k: string | number) {
    size.value = k as any;
    localStorage.setItem(`protable:${props.tableKey}:density`, String(k));
  }

  // ---------- 列显示/隐藏（持久化） ----------
  function colKey(c: any, i: number): string {
    return String(c.key ?? c.title ?? `col-${i}`);
  }
  function colTitle(k: string) {
    const c = props.columns.find((x: any, i) => colKey(x, i) === k);
    return (c as any)?.title ?? k;
  }
  const settableKeys = computed(() =>
    (props.columns as ProColumn[]).map((c, i) => ({ c, i })).filter((x) => x.c.hideable !== false).map((x) => colKey(x.c, x.i))
  );
  const hiddenKeys = ref<string[]>(readHidden());

  function readHidden(): string[] {
    try {
      return JSON.parse(localStorage.getItem(`protable:${props.tableKey}:hidden`) ?? '[]');
    } catch {
      return [];
    }
  }
  const visibleKeys = computed(() => settableKeys.value.filter((k) => !hiddenKeys.value.includes(k)));

  function persistHidden(keys: string[]) {
    hiddenKeys.value = keys;
    localStorage.setItem(`protable:${props.tableKey}:hidden`, JSON.stringify(keys));
  }
  function onVisibleChange(checked: (string | number)[]) {
    const vis = checked.map(String);
    persistHidden(settableKeys.value.filter((k) => !vis.includes(k)));
  }
  function onCheckAll(all: boolean) {
    persistHidden(all ? [] : [...settableKeys.value]);
  }
  function resetColumns() {
    persistHidden([]);
  }

  const visibleColumns = computed(() =>
    (props.columns as ProColumn[]).filter((c, i) => !hiddenKeys.value.includes(colKey(c, i)))
  );

  const paginationOn = computed(() =>
    props.pagination
      ? {
          page: page.value,
          pageSize: pageSize.value,
          itemCount: itemCount.value,
          showSizePicker: true,
          pageSizes: [10, 20, 50],
          prefix: ({ itemCount: n }: any) => `共 ${n} 条`,
        }
      : false
  );
</script>

<style lang="less" scoped>
  .pro-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
    margin-bottom: 12px;
  }

  .pro-toolbar-left {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .pro-toolbar-right {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .pro-tool-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    border-radius: 6px;
    color: var(--text-3, #8b949e);
    cursor: pointer;
    transition:
      background 0.15s,
      color 0.15s;

    &:hover {
      background: var(--hover-bg, rgba(0, 0, 0, 3%));
      color: var(--primary-color, #18a058);
    }
  }

  .pro-col-setting {
    width: 168px;
  }

  .pro-col-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 1px solid var(--line, #e9e9e7);
    padding-bottom: 8px;
    margin-bottom: 8px;
  }

  .pro-col-reset {
    font-size: 12px;
    cursor: pointer;
    color: var(--primary-color, #18a058);
  }

  .pro-col-item {
    display: flex;
    padding: 2px 0;
  }
</style>
