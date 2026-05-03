package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"

	api "github.com/cicbyte/byte-code/api/v1/database"
	service "github.com/cicbyte/byte-code/internal/service"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/database/gdb"
)

func init() {
	service.RegisterDatabase(New())
}

func New() *sDatabase {
	return &sDatabase{}
}

type sDatabase struct{}

// ==================== 数据库连接配置 ====================

func (s *sDatabase) GetDatabase(ctx context.Context, projectId int) (res *api.DatabaseGetRes, err error) {
	var record map[string]interface{}
	err = g.DB().Model("project_databases").Ctx(ctx).Where("project_id", projectId).Scan(&record)
	if err != nil {
		return nil, fmt.Errorf("查询数据库配置失败: %v", err)
	}
	if record == nil {
		return &api.DatabaseGetRes{DbType: "none"}, nil
	}
	return &api.DatabaseGetRes{
		Id:               getInt(record, "id"),
		DbType:           getStr(record, "db_type"),
		DbName:           getStr(record, "db_name"),
		DbHost:           getStr(record, "db_host"),
		DbPort:           getInt(record, "db_port"),
		DbUser:           getStr(record, "db_user"),
		DbPassword:       getStr(record, "db_password"),
		DbOptions:        getStr(record, "db_options"),
		ConnectionStatus: getStr(record, "connection_status"),
		LastTestedAt:     getStr(record, "last_tested_at"),
	}, nil
}

func (s *sDatabase) SaveDatabase(ctx context.Context, req *api.DatabaseSaveReq) (err error) {
	count, _ := g.DB().Model("project_databases").Ctx(ctx).Where("project_id", req.ProjectId).Count()
	data := g.Map{
		"db_type":     req.DbType,
		"db_name":     req.DbName,
		"db_host":     req.DbHost,
		"db_port":     req.DbPort,
		"db_user":     req.DbUser,
		"db_password": req.DbPassword,
		"db_options":  req.DbOptions,
	}
	if count > 0 {
		_, err = g.DB().Model("project_databases").Ctx(ctx).Where("project_id", req.ProjectId).Data(data).Update()
	} else {
		data["project_id"] = req.ProjectId
		_, err = g.DB().Model("project_databases").Ctx(ctx).Data(data).Insert()
	}
	if err != nil {
		return fmt.Errorf("保存数据库配置失败: %v", err)
	}
	return nil
}

func (s *sDatabase) TestConnection(ctx context.Context, projectId int) (res *api.DatabaseTestConnRes, err error) {
	dbConfig, err := s.GetDatabase(ctx, projectId)
	if err != nil {
		return nil, err
	}
	if dbConfig.DbType == "none" {
		return &api.DatabaseTestConnRes{Success: false, Message: "未配置数据库"}, nil
	}

	var dsn string
	var driver string
	switch dbConfig.DbType {
	case "mysql":
		driver = "mysql"
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbConfig.DbUser, dbConfig.DbPassword, dbConfig.DbHost, dbConfig.DbPort, dbConfig.DbName)
	case "postgresql":
		driver = "postgres"
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbConfig.DbHost, dbConfig.DbPort, dbConfig.DbUser, dbConfig.DbPassword, dbConfig.DbName)
	case "sqlite":
		driver = "sqlite3"
		dsn = dbConfig.DbName
	default:
		return &api.DatabaseTestConnRes{Success: false, Message: "不支持的数据库类型"}, nil
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return &api.DatabaseTestConnRes{Success: false, Message: fmt.Sprintf("连接失败: %v", err)}, nil
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		return &api.DatabaseTestConnRes{Success: false, Message: fmt.Sprintf("连接失败: %v", err)}, nil
	}

	g.DB().Model("project_databases").Ctx(ctx).
		Where("project_id", projectId).
		Data(g.Map{"connection_status": "connected", "last_tested_at": "CURRENT_TIMESTAMP"}).
		Update()

	return &api.DatabaseTestConnRes{Success: true, Message: "连接成功"}, nil
}

// ==================== 表管理 ====================

func (s *sDatabase) ListTables(ctx context.Context, projectId int) (res *api.TableListRes, err error) {
	var tables []api.TableItem
	err = g.DB().Model("db_tables").Ctx(ctx).
		Where("project_id", projectId).
		Order("sort_order ASC, id ASC").
		Scan(&tables)
	if err != nil {
		return nil, fmt.Errorf("查询表列表失败: %v", err)
	}

	for i := range tables {
		count, _ := g.DB().Model("db_columns").Ctx(ctx).Where("table_id", tables[i].Id).Count()
		tables[i].ColumnCount = int(count)
	}

	return &api.TableListRes{List: tables}, nil
}

func (s *sDatabase) GetTable(ctx context.Context, id int) (res *api.TableDetailRes, err error) {
	var table api.TableItem
	err = g.DB().Model("db_tables").Ctx(ctx).Where("id", id).Scan(&table)
	if err != nil {
		return nil, fmt.Errorf("查询表失败: %v", err)
	}

	var columns []api.ColumnItem
	err = g.DB().Model("db_columns").Ctx(ctx).
		Where("table_id", id).
		Order("sort_order ASC, id ASC").
		Scan(&columns)
	if err != nil {
		return nil, fmt.Errorf("查询列失败: %v", err)
	}
	if columns == nil {
		columns = []api.ColumnItem{}
	}
	table.Columns = columns
	table.ColumnCount = len(columns)

	return &api.TableDetailRes{TableItem: table}, nil
}

func (s *sDatabase) CreateTable(ctx context.Context, req *api.TableCreateReq) (id int, err error) {
	result, err := g.DB().Model("db_tables").Ctx(ctx).Insert(g.Map{
		"project_id": req.ProjectId,
		"name":       req.Name,
		"comment":    req.Comment,
		"engine":     req.Engine,
		"charset":    req.Charset,
	})
	if err != nil {
		return 0, fmt.Errorf("创建表失败: %v", err)
	}
	lastId, _ := result.LastInsertId()

	s.recordSchemaChange(ctx, req.ProjectId, "create_table", req.Name, "创建表 "+req.Name, "", "")

	return int(lastId), nil
}

func (s *sDatabase) UpdateTable(ctx context.Context, req *api.TableUpdateReq) (err error) {
	data := g.Map{}
	if req.Name != "" {
		data["name"] = req.Name
	}
	if req.Comment != "" {
		data["comment"] = req.Comment
	}
	if req.Engine != "" {
		data["engine"] = req.Engine
	}
	if req.Charset != "" {
		data["charset"] = req.Charset
	}
	data["sort_order"] = req.SortOrder

	_, err = g.DB().Model("db_tables").Ctx(ctx).Where("id", req.Id).Data(data).Update()
	if err != nil {
		return fmt.Errorf("更新表失败: %v", err)
	}
	return nil
}

func (s *sDatabase) DeleteTable(ctx context.Context, id int) (err error) {
	var table map[string]interface{}
	err = g.DB().Model("db_tables").Ctx(ctx).Where("id", id).Scan(&table)
	if err != nil || table == nil {
		return fmt.Errorf("表不存在")
	}

	tableName := getStr(table, "name")
	projectId := getInt(table, "project_id")

	before := s.serializeTable(ctx, id)

	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_, e := tx.Model("db_columns").Ctx(ctx).Where("table_id", id).Delete()
		if e != nil {
			return e
		}
		_, e = tx.Model("db_tables").Ctx(ctx).Where("id", id).Delete()
		return e
	})
	if err != nil {
		return fmt.Errorf("删除表失败: %v", err)
	}

	s.recordSchemaChange(ctx, projectId, "drop_table", tableName, "删除表 "+tableName, before, "")

	return nil
}

// ==================== 列管理 ====================

func (s *sDatabase) SaveColumns(ctx context.Context, req *api.ColumnsSaveReq) (err error) {
	var table map[string]interface{}
	err = g.DB().Model("db_tables").Ctx(ctx).Where("id", req.TableId).Scan(&table)
	if err != nil || table == nil {
		return fmt.Errorf("表不存在")
	}

	tableName := getStr(table, "name")
	projectId := getInt(table, "project_id")

	before := s.serializeTable(ctx, req.TableId)

	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_, e := tx.Model("db_columns").Ctx(ctx).Where("table_id", req.TableId).Delete()
		if e != nil {
			return e
		}
		for i, col := range req.Columns {
			_, e = tx.Model("db_columns").Ctx(ctx).Insert(g.Map{
				"table_id":          req.TableId,
				"name":              col.Name,
				"type":              col.Type,
				"nullable":          col.Nullable,
				"default_value":     col.DefaultValue,
				"is_primary_key":    col.IsPrimaryKey,
				"is_auto_increment": col.IsAutoIncrement,
				"comment":           col.Comment,
				"sort_order":        i,
			})
			if e != nil {
				return e
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("保存列失败: %v", err)
	}

	after := s.serializeTable(ctx, req.TableId)
	desc := s.diffColumns(before, after, tableName)
	s.recordSchemaChange(ctx, projectId, "alter_table", tableName, desc, before, after)

	return nil
}

// ==================== Schema 变更历史 ====================

func (s *sDatabase) ListSchemaChanges(ctx context.Context, req *api.SchemaChangeListReq) (res *api.SchemaChangeListRes, err error) {
	model := g.DB().Model("schema_versions sv").Ctx(ctx).
		LeftJoin("sys_users u", "sv.operator_id = u.id").
		Fields("sv.*, COALESCE(u.real_name, u.username) as operator_name").
		Where("sv.project_id", req.ProjectId).
		Order("sv.id DESC")

	total, _ := model.Count()

	var list []api.SchemaChangeItem
	err = model.Page(req.Page, req.Size).Scan(&list)
	if err != nil {
		return nil, fmt.Errorf("查询变更历史失败: %v", err)
	}
	if list == nil {
		list = []api.SchemaChangeItem{}
	}

	return &api.SchemaChangeListRes{List: list, Total: int(total)}, nil
}

func (s *sDatabase) GetSchemaChange(ctx context.Context, id int) (res *api.SchemaChangeDetailRes, err error) {
	var item api.SchemaChangeItem
	err = g.DB().Model("schema_versions sv").Ctx(ctx).
		LeftJoin("sys_users u", "sv.operator_id = u.id").
		Fields("sv.*, COALESCE(u.real_name, u.username) as operator_name").
		Where("sv.id", id).
		Scan(&item)
	if err != nil {
		return nil, fmt.Errorf("查询变更详情失败: %v", err)
	}
	return &api.SchemaChangeDetailRes{SchemaChangeItem: item}, nil
}

// ==================== 辅助方法 ====================

func (s *sDatabase) serializeTable(ctx context.Context, tableId int) string {
	var columns []api.ColumnItem
	g.DB().Model("db_columns").Ctx(ctx).
		Where("table_id", tableId).
		Order("sort_order ASC, id ASC").
		Scan(&columns)
	if columns == nil {
		columns = []api.ColumnItem{}
	}
	data, _ := json.Marshal(columns)
	return string(data)
}

func (s *sDatabase) recordSchemaChange(ctx context.Context, projectId int, changeType, tableName, description, before, after string) {
	userId := 0
	if uid := ctx.Value("userId"); uid != nil {
		userId = uid.(int)
	}

	maxVer, _ := g.DB().Model("schema_versions").Ctx(ctx).
		Where("project_id", projectId).
		Fields("COALESCE(MAX(version), 0)").Value()
	version := maxVer.Int() + 1

	g.DB().Model("schema_versions").Ctx(ctx).Insert(g.Map{
		"project_id":         projectId,
		"version":            version,
		"change_type":        changeType,
		"change_description": description,
		"table_name":         tableName,
		"sql_statement":      "",
		"schema_before":      before,
		"schema_after":       after,
		"operator_id":        userId,
	})
}

func (s *sDatabase) diffColumns(before, after string, tableName string) string {
	var beforeCols []api.ColumnItem
	var afterCols []api.ColumnItem
	json.Unmarshal([]byte(before), &beforeCols)
	json.Unmarshal([]byte(after), &afterCols)

	beforeMap := make(map[string]api.ColumnItem)
	for _, c := range beforeCols {
		beforeMap[c.Name] = c
	}
	afterMap := make(map[string]api.ColumnItem)
	for _, c := range afterCols {
		afterMap[c.Name] = c
	}

	var added, removed, modified []string

	for name, col := range afterMap {
		if _, exists := beforeMap[name]; !exists {
			added = append(added, fmt.Sprintf("%s(%s)", name, col.Type))
		}
	}
	sort.Strings(added)

	for name, col := range beforeMap {
		if _, exists := afterMap[name]; !exists {
			removed = append(removed, fmt.Sprintf("%s(%s)", name, col.Type))
		}
	}
	sort.Strings(removed)

	for name, after := range afterMap {
		if before, exists := beforeMap[name]; exists {
			if before.Type != after.Type || before.Nullable != after.Nullable ||
				before.DefaultValue != after.DefaultValue || before.IsPrimaryKey != after.IsPrimaryKey ||
				before.Comment != after.Comment {
				modified = append(modified, fmt.Sprintf("%s: %s→%s", name, before.Type, after.Type))
			}
		}
	}
	sort.Strings(modified)

	parts := []string{}
	if len(added) > 0 {
		parts = append(parts, "添加列: "+joinStrings(added, ", "))
	}
	if len(removed) > 0 {
		parts = append(parts, "删除列: "+joinStrings(removed, ", "))
	}
	if len(modified) > 0 {
		parts = append(parts, "修改列: "+joinStrings(modified, ", "))
	}
	if len(parts) == 0 {
		return fmt.Sprintf("更新表 %s", tableName)
	}
	return fmt.Sprintf("表 %s: %s", tableName, joinStrings(parts, "; "))
}

func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

func getStr(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok && v != nil {
		switch val := v.(type) {
		case int:
			return val
		case int64:
			return int(val)
		case float64:
			return int(val)
		}
	}
	return 0
}
