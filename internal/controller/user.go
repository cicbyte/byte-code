package controller

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/user"
	service "github.com/cicbyte/byte-code/internal/service"
)

var UserCtrl = userController{}

type userController struct {
	BaseController
}

func (c *userController) List(ctx context.Context, req *api.ListReq) (res *api.ListRes, err error) {
	return service.User().List(ctx, req)
}

func (c *userController) Create(ctx context.Context, req *api.CreateReq) (res *api.CreateRes, err error) {
	id, err := service.User().Create(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.CreateRes{Id: id}, nil
}

func (c *userController) Update(ctx context.Context, req *api.UpdateReq) (res *api.UpdateRes, err error) {
	err = service.User().Update(ctx, req)
	res = new(api.UpdateRes)
	return
}

func (c *userController) ResetPassword(ctx context.Context, req *api.ResetPasswordReq) (res *api.ResetPasswordRes, err error) {
	err = service.User().ResetPassword(ctx, req)
	res = new(api.ResetPasswordRes)
	return
}

func (c *userController) Delete(ctx context.Context, req *api.DeleteReq) (res *api.DeleteRes, err error) {
	err = service.User().Delete(ctx, req.Id)
	res = new(api.DeleteRes)
	return
}
