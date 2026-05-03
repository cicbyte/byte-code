package controller

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/aiuser"
	"github.com/cicbyte/byte-code/internal/service"
)

var AiUserCtrl = aiUserController{}

type aiUserController struct {
	BaseController
}

func (c *aiUserController) Create(ctx context.Context, req *api.AiUserCreateReq) (res *api.AiUserCreateRes, err error) {
	id, apiKey, err := service.AiUser().Create(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.AiUserCreateRes{Id: id, ApiKey: apiKey}, nil
}

func (c *aiUserController) Update(ctx context.Context, req *api.AiUserUpdateReq) (res *api.AiUserUpdateRes, err error) {
	err = service.AiUser().Update(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.AiUserUpdateRes{}, nil
}

func (c *aiUserController) Delete(ctx context.Context, req *api.AiUserDeleteReq) (res *api.AiUserDeleteRes, err error) {
	err = service.AiUser().Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &api.AiUserDeleteRes{}, nil
}

func (c *aiUserController) List(ctx context.Context, req *api.AiUserListReq) (res *api.AiUserListRes, err error) {
	return service.AiUser().List(ctx, req)
}

func (c *aiUserController) ResetKey(ctx context.Context, req *api.AiUserResetKeyReq) (res *api.AiUserResetKeyRes, err error) {
	apiKey, err := service.AiUser().ResetKey(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &api.AiUserResetKeyRes{ApiKey: apiKey}, nil
}

func (c *aiUserController) AiLogin(ctx context.Context, req *api.AiLoginReq) (res *api.AiLoginRes, err error) {
	token, err := service.AiUser().LoginByApiKey(ctx, req.ApiKey)
	if err != nil {
		return nil, err
	}
	return &api.AiLoginRes{Token: token}, nil
}
