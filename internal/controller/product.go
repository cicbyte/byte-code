package controller

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/product"
	"github.com/cicbyte/byte-code/internal/service"
)

var ProductCtrl = productController{}

type productController struct {
	BaseController
}

func (c *productController) CreateProduct(ctx context.Context, req *api.ProductCreateReq) (res *api.ProductCreateRes, err error) {
	id, err := service.Product().CreateProduct(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.ProductCreateRes{Id: id}, nil
}

func (c *productController) UpdateProduct(ctx context.Context, req *api.ProductUpdateReq) (res *api.ProductUpdateRes, err error) {
	err = service.Product().UpdateProduct(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.ProductUpdateRes{}, nil
}

func (c *productController) DeleteProduct(ctx context.Context, req *api.ProductDeleteReq) (res *api.ProductDeleteRes, err error) {
	err = service.Product().DeleteProduct(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &api.ProductDeleteRes{}, nil
}

func (c *productController) ProductDetail(ctx context.Context, req *api.ProductDetailReq) (res *api.ProductDetailRes, err error) {
	return service.Product().GetProduct(ctx, req.Id)
}

func (c *productController) ProductList(ctx context.Context, req *api.ProductListReq) (res *api.ProductListRes, err error) {
	return service.Product().ListProducts(ctx, req)
}

// ==================== 需求 ====================

func (c *productController) CreateRequirement(ctx context.Context, req *api.RequirementCreateReq) (res *api.RequirementCreateRes, err error) {
	id, err := service.Product().CreateRequirement(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.RequirementCreateRes{Id: id}, nil
}

func (c *productController) UpdateRequirement(ctx context.Context, req *api.RequirementUpdateReq) (res *api.RequirementUpdateRes, err error) {
	err = service.Product().UpdateRequirement(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.RequirementUpdateRes{}, nil
}

func (c *productController) DeleteRequirement(ctx context.Context, req *api.RequirementDeleteReq) (res *api.RequirementDeleteRes, err error) {
	err = service.Product().DeleteRequirement(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &api.RequirementDeleteRes{}, nil
}

func (c *productController) RequirementDetail(ctx context.Context, req *api.RequirementDetailReq) (res *api.RequirementDetailRes, err error) {
	return service.Product().GetRequirement(ctx, req.Id)
}

func (c *productController) RequirementList(ctx context.Context, req *api.RequirementListReq) (res *api.RequirementListRes, err error) {
	return service.Product().ListRequirements(ctx, req)
}

// ==================== 里程碑 ====================

func (c *productController) CreateMilestone(ctx context.Context, req *api.MilestoneCreateReq) (res *api.MilestoneCreateRes, err error) {
	id, err := service.Product().CreateMilestone(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.MilestoneCreateRes{Id: id}, nil
}

func (c *productController) MilestoneList(ctx context.Context, req *api.MilestoneListReq) (res *api.MilestoneListRes, err error) {
	return service.Product().ListMilestones(ctx, req.ProductId)
}
