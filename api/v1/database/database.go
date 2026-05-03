package database

import "github.com/gogf/gf/v2/frame/g"

// ==================== 数据库连接配置 ====================

type DatabaseGetReq struct {
	g.Meta     `path:"/projects/{projectId}/database" method:"get" tags:"数据库模型" summary:"获取数据库配置"`
	ProjectId  int `json:"projectId" v:"required" in:"path"`
}

type DatabaseGetRes struct {
	Id              int    `json:"id"`
	DbType          string `json:"dbType"`
	DbName          string `json:"dbName"`
	DbHost          string `json:"dbHost"`
	DbPort          int    `json:"dbPort"`
	DbUser          string `json:"dbUser"`
	DbPassword      string `json:"dbPassword"`
	DbOptions       string `json:"dbOptions"`
	ConnectionStatus string `json:"connectionStatus"`
	LastTestedAt    string `json:"lastTestedAt"`
}

type DatabaseSaveReq struct {
	g.Meta     `path:"/projects/{projectId}/database" method:"post" tags:"数据库模型" summary:"保存数据库配置"`
	ProjectId  int    `json:"projectId" v:"required" in:"path"`
	DbType     string `json:"dbType" v:"required|in:none,sqlite,mysql,postgresql#数据库类型不能为空|类型不合法"`
	DbName     string `json:"dbName"`
	DbHost     string `json:"dbHost"`
	DbPort     int    `json:"dbPort"`
	DbUser     string `json:"dbUser"`
	DbPassword string `json:"dbPassword"`
	DbOptions  string `json:"dbOptions"`
}

type DatabaseSaveRes struct {
	g.Meta `mime:"application/json"`
}

type DatabaseTestConnReq struct {
	g.Meta    `path:"/projects/{projectId}/database/test" method:"post" tags:"数据库模型" summary:"测试数据库连接"`
	ProjectId int `json:"projectId" v:"required" in:"path"`
}

type DatabaseTestConnRes struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ==================== 表管理 ====================

type TableListReq struct {
	g.Meta    `path:"/projects/{projectId}/db-tables" method:"get" tags:"数据库模型" summary:"表列表"`
	ProjectId int `json:"projectId" v:"required" in:"path"`
}

type TableItem struct {
	Id         int          `json:"id"`
	ProjectId  int          `json:"projectId"`
	Name       string       `json:"name"`
	Comment    string       `json:"comment"`
	Engine     string       `json:"engine"`
	Charset    string       `json:"charset"`
	SortOrder  int          `json:"sortOrder"`
	Columns    []ColumnItem `json:"columns"`
	ColumnCount int         `json:"columnCount"`
	CreatedAt  string       `json:"createdAt"`
	UpdatedAt  string       `json:"updatedAt"`
}

type TableListRes struct {
	List []TableItem `json:"list"`
}

type TableDetailReq struct {
	g.Meta `path:"/db-tables/{id}" method:"get" tags:"数据库模型" summary:"表详情"`
	Id     int `json:"id" v:"required" in:"path"`
}

type TableDetailRes struct {
	TableItem
}

type TableCreateReq struct {
	g.Meta    `path:"/projects/{projectId}/db-tables" method:"post" tags:"数据库模型" summary:"创建表"`
	ProjectId int    `json:"projectId" v:"required" in:"path"`
	Name      string `json:"name" v:"required#表名不能为空"`
	Comment   string `json:"comment"`
	Engine    string `json:"engine"`
	Charset   string `json:"charset"`
}

type TableCreateRes struct {
	Id int `json:"id"`
}

type TableUpdateReq struct {
	g.Meta   `path:"/db-tables/{id}" method:"put" tags:"数据库模型" summary:"更新表"`
	Id       int    `json:"id" v:"required" in:"path"`
	Name     string `json:"name"`
	Comment  string `json:"comment"`
	Engine   string `json:"engine"`
	Charset  string `json:"charset"`
	SortOrder int   `json:"sortOrder"`
}

type TableUpdateRes struct {
	g.Meta `mime:"application/json"`
}

type TableDeleteReq struct {
	g.Meta `path:"/db-tables/{id}" method:"delete" tags:"数据库模型" summary:"删除表"`
	Id     int `json:"id" v:"required" in:"path"`
}

type TableDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// ==================== 列管理 ====================

type ColumnItem struct {
	Id              int    `json:"id"`
	TableId         int    `json:"tableId"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	Nullable        int    `json:"nullable"`
	DefaultValue    string `json:"defaultValue"`
	IsPrimaryKey    int    `json:"isPrimaryKey"`
	IsAutoIncrement int    `json:"isAutoIncrement"`
	Comment         string `json:"comment"`
	SortOrder       int    `json:"sortOrder"`
}

type ColumnInput struct {
	Name            string `json:"name"`
	Type            string `json:"type"`
	Nullable        int    `json:"nullable"`
	DefaultValue    string `json:"defaultValue"`
	IsPrimaryKey    int    `json:"isPrimaryKey"`
	IsAutoIncrement int    `json:"isAutoIncrement"`
	Comment         string `json:"comment"`
	SortOrder       int    `json:"sortOrder"`
}

type ColumnsSaveReq struct {
	g.Meta  `path:"/db-tables/{tableId}/columns" method:"post" tags:"数据库模型" summary:"批量保存列"`
	TableId int            `json:"tableId" v:"required" in:"path"`
	Columns []ColumnInput  `json:"columns" v:"required#列数据不能为空"`
}

type ColumnsSaveRes struct {
	g.Meta `mime:"application/json"`
}

// ==================== Schema 变更历史 ====================

type SchemaChangeListReq struct {
	g.Meta    `path:"/projects/{projectId}/schema-changes" method:"get" tags:"数据库模型" summary:"变更历史列表"`
	ProjectId int    `json:"projectId" v:"required" in:"path"`
	Page      int    `json:"page" in:"query" d:"1"`
	Size      int    `json:"size" in:"query" d:"20"`
}

type SchemaChangeItem struct {
	Id                int    `json:"id"`
	ProjectId         int    `json:"projectId"`
	Version           int    `json:"version"`
	ChangeType        string `json:"changeType"`
	ChangeDescription string `json:"changeDescription"`
	TableName         string `json:"tableName"`
	SqlStatement      string `json:"sqlStatement"`
	SchemaBefore      string `json:"schemaBefore"`
	SchemaAfter       string `json:"schemaAfter"`
	OperatorId        int    `json:"operatorId"`
	OperatorName      string `json:"operatorName"`
	Status            string `json:"status"`
	CreatedAt         string `json:"createdAt"`
}

type SchemaChangeListRes struct {
	List  []SchemaChangeItem `json:"list"`
	Total int                `json:"total"`
}

type SchemaChangeDetailReq struct {
	g.Meta `path:"/schema-changes/{id}" method:"get" tags:"数据库模型" summary:"变更详情"`
	Id     int `json:"id" v:"required" in:"path"`
}

type SchemaChangeDetailRes struct {
	SchemaChangeItem
}
