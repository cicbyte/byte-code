package controller

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/doc"
	"github.com/cicbyte/byte-code/internal/consts"
	service "github.com/cicbyte/byte-code/internal/service"
)

var Doc = docController{}

type docController struct {
	BaseController
}

func (c *docController) DocCreate(ctx context.Context, req *api.DocCreateReq) (res *api.DocCreateRes, err error) {
	res = new(api.DocCreateRes)
	id, err := service.Doc().Create(ctx, req)
	res.Id = id
	return
}

func (c *docController) DocUpdate(ctx context.Context, req *api.DocUpdateReq) (res *api.DocUpdateRes, err error) {
	res = new(api.DocUpdateRes)
	err = service.Doc().Update(ctx, req)
	return
}

func (c *docController) DocDelete(ctx context.Context, req *api.DocDeleteReq) (res *api.DocDeleteRes, err error) {
	res = new(api.DocDeleteRes)
	err = service.Doc().Delete(ctx, req.Id)
	return
}

func (c *docController) DocList(ctx context.Context, req *api.DocListReq) (res *api.DocListRes, err error) {
	res = new(api.DocListRes)
	if req.PageSize == 0 {
		req.PageSize = consts.PageSize
	}
	if req.PageNum == 0 {
		req.PageNum = 1
	}
	total, list, err := service.Doc().List(ctx, req)
	res.Total = total
	res.CurrentPage = req.PageNum
	res.List = list
	return
}

func (c *docController) DocDetail(ctx context.Context, req *api.DocDetailReq) (res *api.DocDetailRes, err error) {
	return service.Doc().Get(ctx, req.Id)
}

func (c *docController) DocMove(ctx context.Context, req *api.DocMoveReq) (res *api.DocMoveRes, err error) {
	res = new(api.DocMoveRes)
	err = service.Doc().Move(ctx, req)
	return
}

func (c *docController) DocTree(ctx context.Context, req *api.DocTreeReq) (res *api.DocTreeRes, err error) {
	res = new(api.DocTreeRes)
	tree, err := service.Doc().Tree(ctx, req.ProjectId)
	res.Tree = tree
	return
}

func (c *docController) DocVersionList(ctx context.Context, req *api.DocVersionListReq) (res *api.DocVersionListRes, err error) {
	res = new(api.DocVersionListRes)
	if req.PageSize == 0 {
		req.PageSize = consts.PageSize
	}
	if req.PageNum == 0 {
		req.PageNum = 1
	}
	total, list, err := service.Doc().Versions(ctx, req)
	res.Total = total
	res.CurrentPage = req.PageNum
	res.List = list
	return
}

func (c *docController) DocRevert(ctx context.Context, req *api.DocRevertReq) (res *api.DocRevertRes, err error) {
	res = new(api.DocRevertRes)
	err = service.Doc().Revert(ctx, req)
	return
}

func (c *docController) DocRelationCreate(ctx context.Context, req *api.DocRelationCreateReq) (res *api.DocRelationCreateRes, err error) {
	res = new(api.DocRelationCreateRes)
	id, err := service.Doc().CreateRelation(ctx, req)
	res.Id = id
	return
}

func (c *docController) DocRelationDelete(ctx context.Context, req *api.DocRelationDeleteReq) (res *api.DocRelationDeleteRes, err error) {
	res = new(api.DocRelationDeleteRes)
	err = service.Doc().DeleteRelation(ctx, req.Id, req.TargetType, req.TargetId)
	return
}

func (c *docController) DocRelationList(ctx context.Context, req *api.DocRelationListReq) (res *api.DocRelationListRes, err error) {
	res = new(api.DocRelationListRes)
	list, err := service.Doc().ListRelations(ctx, req.Id)
	res.List = list
	return
}

func (c *docController) Search(ctx context.Context, req *api.SearchReq) (res *api.SearchRes, err error) {
	res = new(api.SearchRes)
	if req.PageSize == 0 {
		req.PageSize = 20
	}
	if req.PageNum == 0 {
		req.PageNum = 1
	}
	total, list, err := service.Doc().Search(ctx, req)
	res.Total = total
	res.CurrentPage = req.PageNum
	res.List = list
	return
}
