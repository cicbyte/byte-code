import { Alova } from '@/utils/http/alova/index';

// ==================== 类型定义 ====================

/** 数据库配置 */
export interface DatabaseConfig {
  id: number;
  dbType: string;
  dbName: string;
  dbHost: string;
  dbPort: number;
  dbUser: string;
  dbPassword: string;
  dbOptions: string;
  connectionStatus: string;
  lastTestedAt: string;
}

export interface DatabaseSaveData {
  dbType: string;
  dbName?: string;
  dbHost?: string;
  dbPort?: number;
  dbUser?: string;
  dbPassword?: string;
  dbOptions?: string;
}

/** 数据库列 */
export interface DbColumn {
  id?: number;
  tableId?: number;
  name: string;
  type: string;
  nullable: number;
  defaultValue: string;
  isPrimaryKey: number;
  isAutoIncrement: number;
  comment: string;
  sortOrder: number;
}

/** 数据库表 */
export interface DbTable {
  id: number;
  projectId: number;
  name: string;
  comment: string;
  engine: string;
  charset: string;
  sortOrder: number;
  columns: DbColumn[];
  columnCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface DbTableCreateData {
  name: string;
  comment?: string;
  engine?: string;
  charset?: string;
}

export interface DbTableUpdateData {
  name?: string;
  comment?: string;
  engine?: string;
  charset?: string;
  sortOrder?: number;
}

/** Schema 变更 */
export interface SchemaChange {
  id: number;
  projectId: number;
  version: number;
  changeType: string;
  changeDescription: string;
  tableName: string;
  sqlStatement: string;
  schemaBefore: string;
  schemaAfter: string;
  operatorId: number;
  operatorName: string;
  status: string;
  createdAt: string;
}

export interface SchemaChangeListResult {
  list: SchemaChange[];
  total: number;
}

// ==================== API ====================

/** 获取数据库配置 */
export function getDatabaseConfig(projectId: number) {
  return Alova.Get<DatabaseConfig>(`/v1/projects/${projectId}/database`);
}

/** 保存数据库配置 */
export function saveDatabaseConfig(projectId: number, data: DatabaseSaveData) {
  return Alova.Post(`/v1/projects/${projectId}/database`, data);
}

/** 测试数据库连接 */
export function testDatabaseConnection(projectId: number) {
  return Alova.Post<{ success: boolean; message: string }>(`/v1/projects/${projectId}/database/test`);
}

/** 表列表 */
export function getDbTables(projectId: number) {
  return Alova.Get<{ list: DbTable[] }>(`/v1/projects/${projectId}/db-tables`);
}

/** 表详情 */
export function getDbTable(id: number) {
  return Alova.Get<DbTable>(`/v1/db-tables/${id}`);
}

/** 创建表 */
export function createDbTable(projectId: number, data: DbTableCreateData) {
  return Alova.Post<{ id: number }>(`/v1/projects/${projectId}/db-tables`, data);
}

/** 更新表 */
export function updateDbTable(id: number, data: DbTableUpdateData) {
  return Alova.Put(`/v1/db-tables/${id}`, data);
}

/** 删除表 */
export function deleteDbTable(id: number) {
  return Alova.Delete(`/v1/db-tables/${id}`);
}

/** 批量保存列 */
export function saveDbColumns(tableId: number, columns: DbColumn[]) {
  return Alova.Post(`/v1/db-tables/${tableId}/columns`, { columns });
}

/** 变更历史列表 */
export function getSchemaChanges(projectId: number, params?: { page?: number; size?: number }) {
  return Alova.Get<SchemaChangeListResult>(`/v1/projects/${projectId}/schema-changes`, { params });
}

/** 变更详情 */
export function getSchemaChangeDetail(id: number) {
  return Alova.Get<SchemaChange>(`/v1/schema-changes/${id}`);
}
