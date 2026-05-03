package service

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/doc"
)

type IDoc interface {
	// 文档CRUD
	Create(ctx context.Context, req *api.DocCreateReq) (id int, err error)
	Update(ctx context.Context, req *api.DocUpdateReq) (err error)
	Delete(ctx context.Context, id int) (err error)
	List(ctx context.Context, req *api.DocListReq) (total int, list []api.DocItem, err error)
	Get(ctx context.Context, id int) (res *api.DocDetailRes, err error)
	Move(ctx context.Context, req *api.DocMoveReq) (err error)
	Tree(ctx context.Context, projectId int) (tree []api.DocTreeNode, err error)
	// 文档版本
	Versions(ctx context.Context, req *api.DocVersionListReq) (total int, list []api.DocVersionItem, err error)
	Revert(ctx context.Context, req *api.DocRevertReq) (err error)
	// 文档关联
	CreateRelation(ctx context.Context, req *api.DocRelationCreateReq) (id int, err error)
	DeleteRelation(ctx context.Context, docId int, targetType string, targetId int) (err error)
	ListRelations(ctx context.Context, docId int) (list []api.DocRelationItem, err error)
	// 搜索
	Search(ctx context.Context, req *api.SearchReq) (total int, list []api.SearchResult, err error)
}

var localDoc IDoc

func Doc() IDoc {
	if localDoc == nil {
		panic("implement not found for interface IDoc, forgot register?")
	}
	return localDoc
}

func RegisterDoc(i IDoc) {
	localDoc = i
}
