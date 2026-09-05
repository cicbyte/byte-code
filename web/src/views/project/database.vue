<template>
  <div>
    <!-- 数据库连接配置 -->
    <n-card title="数据库配置" :bordered="false" class="mb-4">
      <n-form label-placement="left" :label-width="80">
        <n-form-item label="数据库类型">
          <n-select
            v-model:value="dbConfig.dbType"
            :options="dbTypeOptions"
            style="width: 200px"
            @update:value="handleDbTypeChange"
          />
        </n-form-item>
        <template v-if="dbConfig.dbType !== 'none'">
          <template v-if="dbConfig.dbType === 'mysql' || dbConfig.dbType === 'postgresql'">
            <n-form-item label="主机">
              <n-input v-model:value="dbConfig.dbHost" placeholder="localhost" />
            </n-form-item>
            <n-form-item label="端口">
              <n-input-number v-model:value="dbConfig.dbPort" :min="0" :max="65535" placeholder="如 3306" style="width: 150px" />
            </n-form-item>
          </template>
          <n-form-item :label="dbConfig.dbType === 'sqlite' ? '数据库路径' : '数据库名'">
            <n-input v-model:value="dbConfig.dbName" :placeholder="dbConfig.dbType === 'sqlite' ? 'data/app.db' : 'mydb'" />
          </n-form-item>
          <template v-if="dbConfig.dbType === 'mysql' || dbConfig.dbType === 'postgresql'">
            <n-form-item label="用户名">
              <n-input v-model:value="dbConfig.dbUser" placeholder="root" />
            </n-form-item>
            <n-form-item label="密码">
              <n-input v-model:value="dbConfig.dbPassword" type="password" show-password-on="click" placeholder="密码" />
            </n-form-item>
          </template>
        </template>
        <n-space v-if="dbConfig.dbType !== 'none'" class="mt-2">
          <n-button type="primary" :loading="saving" @click="handleSaveConfig">保存配置</n-button>
          <n-button :loading="testing" @click="handleTestConn">
            测试连接
            <template #icon><n-icon><ApiOutlined /></n-icon></template>
          </n-button>
          <n-tag v-if="connResult !== null" :type="connResult ? 'success' : 'error'" size="small">
            {{ connResult ? '连接成功' : connMessage }}
          </n-tag>
        </n-space>
        <EmptyState type="data" title="未配置数据库" v-if="dbConfig.dbType === 'none'" description="在数据库页完成配置后可管理表结构" />
      </n-form>
    </n-card>

    <!-- 表和列管理 -->
    <template v-if="dbConfig.dbType !== 'none'">
      <n-card title="数据表" :bordered="false" class="mb-4">
        <template #header-extra>
          <n-button type="primary" size="small" @click="handleCreateTable">
            <template #icon><n-icon><PlusOutlined /></n-icon></template>
            新建表
          </n-button>
        </template>

        <n-spin :show="tablesLoading">
          <EmptyState type="data" title="暂无数据表" v-if="!tablesLoading && tables.length === 0" description="同步或新建表后展示在这里" />
          <n-collapse v-else v-model:expanded-names="expandedTable" accordion>
            <n-collapse-item v-for="table in tables" :key="table.id" :name="table.id">
              <template #header>
                <n-space align="center" :size="8">
                  <span class="font-medium">{{ table.name }}</span>
                  <n-tag size="tiny" :bordered="false">{{ table.columnCount }} 列</n-tag>
                  <span v-if="table.comment" class="text-gray-400 text-xs">{{ table.comment }}</span>
                </n-space>
              </template>
              <template #header-extra>
                <n-space :size="4" @click.stop>
                  <n-button text type="info" size="small" @click="handleEditTable(table)">编辑</n-button>
                  <n-button text type="error" size="small" @click="handleDeleteTable(table)">删除</n-button>
                </n-space>
              </template>

              <!-- 列编辑器 -->
              <div v-if="editingTableId === table.id">
                <n-table :bordered="false" :single-line="false" size="small">
                  <thead>
                    <tr>
                      <th>字段名</th>
                      <th>类型</th>
                      <th>可空</th>
                      <th>默认值</th>
                      <th>主键</th>
                      <th>自增</th>
                      <th>注释</th>
                      <th style="width: 80px">操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="(col, idx) in editColumns" :key="col._uid">
                      <td><n-input v-model:value="col.name" size="small" placeholder="字段名" /></td>
                      <td style="width: 160px">
                        <n-select v-model:value="col.type" :options="columnTypeOptions" size="small" />
                      </td>
                      <td><n-switch v-model:value="col.nullable" :unchecked-value="0" :checked-value="1" size="small" /></td>
                      <td><n-input v-model:value="col.defaultValue" size="small" placeholder="默认值" /></td>
                      <td><n-switch v-model:value="col.isPrimaryKey" :unchecked-value="0" :checked-value="1" size="small" /></td>
                      <td><n-switch v-model:value="col.isAutoIncrement" :unchecked-value="0" :checked-value="1" size="small" /></td>
                      <td><n-input v-model:value="col.comment" size="small" placeholder="注释" /></td>
                      <td>
                        <n-button text type="error" size="small" @click="editColumns.splice(idx, 1)">删除</n-button>
                      </td>
                    </tr>
                  </tbody>
                </n-table>
                <n-space class="mt-2">
                  <n-button size="small" @click="addColumn">添加列</n-button>
                  <n-button type="primary" size="small" :loading="savingColumns" @click="handleSaveColumns(table.id)">保存列</n-button>
                  <n-button size="small" @click="editingTableId = null">取消</n-button>
                </n-space>
              </div>
              <div v-else>
                <n-button size="small" type="primary" ghost @click="handleEditColumns(table)">编辑字段</n-button>
              </div>
            </n-collapse-item>
          </n-collapse>
        </n-spin>
      </n-card>

      <!-- 变更历史 -->
      <n-card title="变更历史" :bordered="false">
        <n-spin :show="changesLoading">
          <EmptyState type="data" title="暂无变更记录" v-if="!changesLoading && changes.length === 0" description="Schema 变更历史会记录在这里" />
          <n-timeline v-else>
            <n-timeline-item
              v-for="change in changes"
              :key="change.id"
              :type="changeTypeColor(change.changeType)"
              :title="`v${change.version} ${change.changeDescription}`"
              :time="change.createdAt"
            >
              <template #header>
                <n-space :size="8" align="center">
                  <n-tag :type="changeTypeColor(change.changeType)" size="small">{{ changeTypeLabel(change.changeType) }}</n-tag>
                  <span>{{ change.changeDescription }}</span>
                  <span class="text-gray-400 text-xs">{{ change.operatorName }}</span>
                </n-space>
              </template>
            </n-timeline-item>
          </n-timeline>
        </n-spin>
        <div class="mt-4 flex justify-end" v-if="changesTotal > changesPagination.size">
          <n-pagination
            v-model:page="changesPagination.page"
            v-model:page-size="changesPagination.size"
            :item-count="changesTotal"
            @update:page="loadChanges"
          />
        </div>
      </n-card>
    </template>

    <!-- 新建/编辑表弹窗 -->
    <n-modal
      v-model:show="showTableModal"
      preset="dialog"
      :title="isEditTable ? '编辑表' : '新建表'"
      positive-text="确定"
      negative-text="取消"
      @positive-click="handleTableSubmit"
      style="width: 480px"
    >
      <n-form label-placement="left" :label-width="80" class="py-4">
        <n-form-item label="表名">
          <n-input v-model:value="tableForm.name" placeholder="请输入表名" />
        </n-form-item>
        <n-form-item label="注释">
          <n-input v-model:value="tableForm.comment" type="textarea" placeholder="表注释" :rows="2" />
        </n-form-item>
        <n-form-item label="引擎">
          <n-input v-model:value="tableForm.engine" placeholder="InnoDB" />
        </n-form-item>
        <n-form-item label="字符集">
          <n-input v-model:value="tableForm.charset" placeholder="utf8mb4" />
        </n-form-item>
      </n-form>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { ref, reactive, onMounted, computed } from 'vue';
  import { useRoute } from 'vue-router';
  import { useMessage, useDialog } from 'naive-ui';
  import { ApiOutlined, PlusOutlined } from '@vicons/antd';
  import {
    getDatabaseConfig,
    saveDatabaseConfig,
    testDatabaseConnection,
    getDbTables,
    createDbTable,
    updateDbTable,
    deleteDbTable,
    saveDbColumns,
    getSchemaChanges,
  } from '@/api/database/index';
  import type { DatabaseConfig, DbTable, DbColumn, SchemaChange } from '@/api/database/index';

  const route = useRoute();
  const message = useMessage();
  const dialog = useDialog();
  const projectId = computed(() => Number(route.params.projectId));

  const dbTypeOptions = [
    { label: '未配置', value: 'none' },
    { label: 'SQLite', value: 'sqlite' },
    { label: 'MySQL', value: 'mysql' },
    { label: 'PostgreSQL', value: 'postgresql' },
  ];

  const columnTypeOptions = [
    'INT', 'BIGINT', 'SMALLINT', 'TINYINT',
    'VARCHAR(255)', 'VARCHAR(100)', 'VARCHAR(50)', 'CHAR(36)',
    'TEXT', 'MEDIUMTEXT', 'LONGTEXT',
    'DATETIME', 'TIMESTAMP', 'DATE', 'TIME',
    'BOOLEAN', 'TINYINT(1)',
    'DECIMAL(10,2)', 'FLOAT', 'DOUBLE',
    'BLOB', 'JSON', 'ENUM',
  ].map(t => ({ label: t, value: t }));

  // 数据库配置
  const dbConfig = reactive<DatabaseConfig>({
    id: 0, dbType: 'none', dbName: '', dbHost: '', dbPort: 3306,
    dbUser: '', dbPassword: '', dbOptions: '', connectionStatus: 'disconnected', lastTestedAt: '',
  });
  const saving = ref(false);
  const testing = ref(false);
  const connResult = ref<boolean | null>(null);
  const connMessage = ref('');

  async function loadConfig() {
    try {
      const res = await getDatabaseConfig(projectId.value);
      if (res) Object.assign(dbConfig, res);
    } catch { /* ignore */ }
  }

  function handleDbTypeChange(val: string) {
    if (val === 'mysql') { dbConfig.dbPort = 3306; }
    else if (val === 'postgresql') { dbConfig.dbPort = 5432; }
  }

  async function handleSaveConfig() {
    saving.value = true;
    try {
      await saveDatabaseConfig(projectId.value, {
        dbType: dbConfig.dbType,
        dbName: dbConfig.dbName,
        dbHost: dbConfig.dbHost,
        dbPort: dbConfig.dbPort,
        dbUser: dbConfig.dbUser,
        dbPassword: dbConfig.dbPassword,
        dbOptions: dbConfig.dbOptions,
      });
      message.success('保存成功');
    } catch { message.error('保存失败'); }
    finally { saving.value = false; }
  }

  async function handleTestConn() {
    testing.value = true;
    connResult.value = null;
    try {
      const res = await testDatabaseConnection(projectId.value);
      connResult.value = res.success;
      connMessage.value = res.message;
    } catch {
      connResult.value = false;
      connMessage.value = '测试失败';
    }
    finally { testing.value = false; }
  }

  // 表管理
  const tables = ref<DbTable[]>([]);
  const tablesLoading = ref(false);
  const expandedTable = ref<number | null>(null);
  const showTableModal = ref(false);
  const isEditTable = ref(false);
  const editTableId = ref<number | null>(null);
  const tableForm = reactive({ name: '', comment: '', engine: '', charset: '' });

  async function loadTables() {
    tablesLoading.value = true;
    try {
      const res = await getDbTables(projectId.value);
      tables.value = res?.list || [];
    } catch { /* ignore */ }
    finally { tablesLoading.value = false; }
  }

  function handleCreateTable() {
    isEditTable.value = false;
    editTableId.value = null;
    tableForm.name = '';
    tableForm.comment = '';
    tableForm.engine = '';
    tableForm.charset = '';
    showTableModal.value = true;
  }

  function handleEditTable(table: DbTable) {
    isEditTable.value = true;
    editTableId.value = table.id;
    tableForm.name = table.name;
    tableForm.comment = table.comment;
    tableForm.engine = table.engine;
    tableForm.charset = table.charset;
    showTableModal.value = true;
  }

  async function handleTableSubmit() {
    if (!tableForm.name) { message.warning('请输入表名'); return false; }
    try {
      if (isEditTable.value && editTableId.value) {
        await updateDbTable(editTableId.value, { ...tableForm });
        message.success('更新成功');
      } else {
        await createDbTable(projectId.value, { ...tableForm });
        message.success('创建成功');
      }
      loadTables();
    } catch { message.error('操作失败'); return false; }
  }

  function handleDeleteTable(table: DbTable) {
    dialog.warning({
      title: '确认删除',
      content: `确定要删除表「${table.name}」及其所有字段吗？`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await deleteDbTable(table.id);
          message.success('删除成功');
          loadTables();
          loadChanges();
        } catch { message.error('删除失败'); }
      },
    });
  }

  // 列编辑器
  const editingTableId = ref<number | null>(null);
  const editColumns = ref<(DbColumn & { _uid: number })[]>([]);
  const savingColumns = ref(false);
  let uidCounter = 0;

  function addColumn() {
    editColumns.value.push({
      _uid: ++uidCounter,
      name: '', type: 'VARCHAR(255)', nullable: 1, defaultValue: '',
      isPrimaryKey: 0, isAutoIncrement: 0, comment: '', sortOrder: editColumns.value.length,
    });
  }

  async function handleEditColumns(table: DbTable) {
    editingTableId.value = table.id;
    try {
      const detail = await getDbTable(table.id);
      editColumns.value = detail?.columns?.length
        ? detail.columns.map(c => ({ ...c, _uid: ++uidCounter }))
        : [];
    } catch { editColumns.value = []; }
  }

  async function handleSaveColumns(tableId: number) {
    const valid = editColumns.value.filter(c => c.name.trim());
    if (valid.length === 0) { message.warning('请至少添加一个字段'); return; }
    savingColumns.value = true;
    try {
      await saveDbColumns(tableId, valid.map((c, i) => ({ ...c, sortOrder: i })));
      message.success('保存成功');
      editingTableId.value = null;
      loadTables();
      loadChanges();
    } catch { message.error('保存失败'); }
    finally { savingColumns.value = false; }
  }

  // 变更历史
  const changes = ref<SchemaChange[]>([]);
  const changesTotal = ref(0);
  const changesLoading = ref(false);
  const changesPagination = reactive({ page: 1, size: 10 });

  async function loadChanges() {
    changesLoading.value = true;
    try {
      const res = await getSchemaChanges(projectId.value, { page: changesPagination.page, size: changesPagination.size });
      changes.value = res?.list || [];
      changesTotal.value = res?.total || 0;
    } catch { /* ignore */ }
    finally { changesLoading.value = false; }
  }

  function changeTypeColor(type: string) {
    switch (type) {
      case 'create_table': return 'success';
      case 'alter_table': return 'warning';
      case 'drop_table': return 'error';
      default: return 'default';
    }
  }

  function changeTypeLabel(type: string) {
    switch (type) {
      case 'create_table': return '建表';
      case 'alter_table': return '修改';
      case 'drop_table': return '删表';
      default: return type;
    }
  }

  onMounted(() => {
    loadConfig();
    loadTables();
    loadChanges();
  });
</script>
