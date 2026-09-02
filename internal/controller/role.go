package controller

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/role"
	service "github.com/cicbyte/byte-code/internal/service"
)

var Role = roleController{}

type roleController struct {
	BaseController
}

func (c *roleController) List(ctx context.Context, req *api.ListReq) (res *api.ListRes, err error) {
	return service.Role().List(ctx, req)
}

func (c *roleController) Create(ctx context.Context, req *api.CreateReq) (res *api.CreateRes, err error) {
	id, err := service.Role().Create(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.CreateRes{Id: id}, nil
}

func (c *roleController) Update(ctx context.Context, req *api.UpdateReq) (res *api.UpdateRes, err error) {
	err = service.Role().Update(ctx, req)
	res = new(api.UpdateRes)
	return
}

func (c *roleController) Delete(ctx context.Context, req *api.DeleteReq) (res *api.DeleteRes, err error) {
	err = service.Role().Delete(ctx, req.Id)
	res = new(api.DeleteRes)
	return
}

func (c *roleController) UpdateMenus(ctx context.Context, req *api.UpdateMenusReq) (res *api.UpdateMenusRes, err error) {
	err = service.Role().UpdateMenus(ctx, req.Id, req.MenuIds)
	res = new(api.UpdateMenusRes)
	return
}
