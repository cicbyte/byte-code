package controller

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/database"
	"github.com/cicbyte/byte-code/internal/service"
)

var DatabaseCtrl = databaseController{}

type databaseController struct {
	BaseController
}

// ==================== 数据库连接配置 ====================

func (c *databaseController) GetDatabase(ctx context.Context, req *api.DatabaseGetReq) (res *api.DatabaseGetRes, err error) {
	return service.Database().GetDatabase(ctx, req.ProjectId)
}

func (c *databaseController) SaveDatabase(ctx context.Context, req *api.DatabaseSaveReq) (res *api.DatabaseSaveRes, err error) {
	err = service.Database().SaveDatabase(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.DatabaseSaveRes{}, nil
}

func (c *databaseController) TestConnection(ctx context.Context, req *api.DatabaseTestConnReq) (res *api.DatabaseTestConnRes, err error) {
	return service.Database().TestConnection(ctx, req.ProjectId)
}

// ==================== 表管理 ====================

func (c *databaseController) ListTables(ctx context.Context, req *api.TableListReq) (res *api.TableListRes, err error) {
	return service.Database().ListTables(ctx, req.ProjectId)
}

func (c *databaseController) GetTable(ctx context.Context, req *api.TableDetailReq) (res *api.TableDetailRes, err error) {
	return service.Database().GetTable(ctx, req.Id)
}

func (c *databaseController) CreateTable(ctx context.Context, req *api.TableCreateReq) (res *api.TableCreateRes, err error) {
	id, err := service.Database().CreateTable(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.TableCreateRes{Id: id}, nil
}

func (c *databaseController) UpdateTable(ctx context.Context, req *api.TableUpdateReq) (res *api.TableUpdateRes, err error) {
	err = service.Database().UpdateTable(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.TableUpdateRes{}, nil
}

func (c *databaseController) DeleteTable(ctx context.Context, req *api.TableDeleteReq) (res *api.TableDeleteRes, err error) {
	err = service.Database().DeleteTable(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &api.TableDeleteRes{}, nil
}

// ==================== 列管理 ====================

func (c *databaseController) SaveColumns(ctx context.Context, req *api.ColumnsSaveReq) (res *api.ColumnsSaveRes, err error) {
	err = service.Database().SaveColumns(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.ColumnsSaveRes{}, nil
}

// ==================== Schema 变更历史 ====================

func (c *databaseController) ListSchemaChanges(ctx context.Context, req *api.SchemaChangeListReq) (res *api.SchemaChangeListRes, err error) {
	return service.Database().ListSchemaChanges(ctx, req)
}

func (c *databaseController) GetSchemaChange(ctx context.Context, req *api.SchemaChangeDetailReq) (res *api.SchemaChangeDetailRes, err error) {
	return service.Database().GetSchemaChange(ctx, req.Id)
}
