package controller

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/categories"
	consts "github.com/cicbyte/byte-code/internal/consts"
	service "github.com/cicbyte/byte-code/internal/service"
)

var Categories = categoriesController{}

type categoriesController struct {
	BaseController
}

func (c *categoriesController) Add(ctx context.Context, req *api.CategoriesAddReq) (res *api.CategoriesAddRes, err error) {
	res = new(api.CategoriesAddRes)
	err = service.Categories().Add(ctx, req)
	return
}

func (c *categoriesController) List(ctx context.Context, req *api.CategoriesListReq) (res *api.CategoriesListRes, err error) {
	res = new(api.CategoriesListRes)
	if req.PageSize == 0 {
		req.PageSize = consts.PageSize
	}
	if req.PageNum == 0 {
		req.PageNum = 1
	}
	total, categoriess, err := service.Categories().List(ctx, req)
	res.Total = total
	res.CurrentPage = req.PageNum
	res.CategoriesList = categoriess
	return
}

func (c *categoriesController) Get(ctx context.Context, req *api.CategoriesDetailReq) (res *api.CategoriesDetailRes, err error) {
	// 原实现丢弃了查询结果，接口恒返回空对象
	info, err := service.Categories().GetById(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	res = new(api.CategoriesDetailRes)
	res.CategoriesInfo = info
	return
}

func (c *categoriesController) Edit(ctx context.Context, req *api.CategoriesEditReq) (res *api.CategoriesEditRes, err error) {
	err = service.Categories().Edit(ctx, req)
	return
}

func (c *categoriesController) Delete(ctx context.Context, req *api.CategoriesDelReq) (res *api.CategoriesDelRes, err error) {
	err = service.Categories().Delete(ctx, req.Id)
	return
}

func (c *categoriesController) BatchDelete(ctx context.Context, req *api.CategoriesBatchDelReq) (res *api.CategoriesBatchDelRes, err error) {
	err = service.Categories().BatchDelete(ctx, req.Ids)
	return
}