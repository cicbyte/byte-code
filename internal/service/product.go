package service

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/product"
)

type IProduct interface {
	CreateProduct(ctx context.Context, req *api.ProductCreateReq) (id int, err error)
	UpdateProduct(ctx context.Context, req *api.ProductUpdateReq) (err error)
	DeleteProduct(ctx context.Context, id int) (err error)
	GetProduct(ctx context.Context, id int) (res *api.ProductDetailRes, err error)
	ListProducts(ctx context.Context, req *api.ProductListReq) (res *api.ProductListRes, err error)

	CreateRequirement(ctx context.Context, req *api.RequirementCreateReq) (id int, err error)
	UpdateRequirement(ctx context.Context, req *api.RequirementUpdateReq) (err error)
	DeleteRequirement(ctx context.Context, id int) (err error)
	GetRequirement(ctx context.Context, id int) (res *api.RequirementDetailRes, err error)
	ListRequirements(ctx context.Context, req *api.RequirementListReq) (res *api.RequirementListRes, err error)

	CreateMilestone(ctx context.Context, req *api.MilestoneCreateReq) (id int, err error)
	ListMilestones(ctx context.Context, productId int) (res *api.MilestoneListRes, err error)
}

var localProduct IProduct

func Product() IProduct {
	if localProduct == nil {
		panic("implement not found for interface IProduct, forgot register?")
	}
	return localProduct
}

func RegisterProduct(i IProduct) {
	localProduct = i
}
