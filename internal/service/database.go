package service

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/database"
)

type IDatabase interface {
	// 数据库连接配置
	GetDatabase(ctx context.Context, projectId int) (res *api.DatabaseGetRes, err error)
	SaveDatabase(ctx context.Context, req *api.DatabaseSaveReq) (err error)
	TestConnection(ctx context.Context, projectId int) (res *api.DatabaseTestConnRes, err error)

	// 表管理
	ListTables(ctx context.Context, projectId int) (res *api.TableListRes, err error)
	GetTable(ctx context.Context, id int) (res *api.TableDetailRes, err error)
	CreateTable(ctx context.Context, req *api.TableCreateReq) (id int, err error)
	UpdateTable(ctx context.Context, req *api.TableUpdateReq) (err error)
	DeleteTable(ctx context.Context, id int) (err error)

	// 列管理
	SaveColumns(ctx context.Context, req *api.ColumnsSaveReq) (err error)

	// Schema 变更历史
	ListSchemaChanges(ctx context.Context, req *api.SchemaChangeListReq) (res *api.SchemaChangeListRes, err error)
	GetSchemaChange(ctx context.Context, id int) (res *api.SchemaChangeDetailRes, err error)
}

var localDatabase IDatabase

func Database() IDatabase {
	if localDatabase == nil {
		panic("implement not found for interface IDatabase, forgot register?")
	}
	return localDatabase
}

func RegisterDatabase(i IDatabase) {
	localDatabase = i
}
