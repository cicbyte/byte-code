// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package v1

import (
	"context"

	"github.com/cicbyte/byte-code/api/v1/auth"
	"github.com/cicbyte/byte-code/api/v1/categories"
	"github.com/cicbyte/byte-code/api/v1/dashboard"
	"github.com/cicbyte/byte-code/api/v1/health"
	"github.com/cicbyte/byte-code/api/v1/menu"
	"github.com/cicbyte/byte-code/api/v1/role"
	"github.com/cicbyte/byte-code/api/v1/setting"
)

type IV1Auth interface {
	Login(ctx context.Context, req *auth.LoginReq) (res *auth.LoginRes, err error)
	AdminInfo(ctx context.Context, req *auth.AdminInfoReq) (res *auth.AdminInfoRes, err error)
	Logout(ctx context.Context, req *auth.LogoutReq) (res *auth.LogoutRes, err error)
}

type IV1Categories interface {
	CategoriesAdd(ctx context.Context, req *categories.CategoriesAddReq) (res *categories.CategoriesAddRes, err error)
	CategoriesDel(ctx context.Context, req *categories.CategoriesDelReq) (res *categories.CategoriesDelRes, err error)
	CategoriesBatchDel(ctx context.Context, req *categories.CategoriesBatchDelReq) (res *categories.CategoriesBatchDelRes, err error)
	CategoriesEdit(ctx context.Context, req *categories.CategoriesEditReq) (res *categories.CategoriesEditRes, err error)
	CategoriesList(ctx context.Context, req *categories.CategoriesListReq) (res *categories.CategoriesListRes, err error)
	CategoriesDetail(ctx context.Context, req *categories.CategoriesDetailReq) (res *categories.CategoriesDetailRes, err error)
	PublicCategoriesList(ctx context.Context, req *categories.PublicCategoriesListReq) (res *categories.PublicCategoriesListRes, err error)
}

type IV1Dashboard interface {
	Console(ctx context.Context, req *dashboard.ConsoleReq) (res *dashboard.ConsoleRes, err error)
}

type IV1Health interface {
	Health(ctx context.Context, req *health.HealthReq) (res *health.HealthRes, err error)
	HealthDetail(ctx context.Context, req *health.HealthDetailReq) (res *health.HealthDetailRes, err error)
}

type IV1Menu interface {
	Menus(ctx context.Context, req *menu.MenusReq) (res *menu.MenusRes, err error)
	MenuList(ctx context.Context, req *menu.MenuListReq) (res *menu.MenuListRes, err error)
}

type IV1Role interface {
	List(ctx context.Context, req *role.ListReq) (res *role.ListRes, err error)
}

type IV1Setting interface {
	GetProfile(ctx context.Context, req *setting.GetProfileReq) (res *setting.GetProfileRes, err error)
	UpdateProfile(ctx context.Context, req *setting.UpdateProfileReq) (res *setting.UpdateProfileRes, err error)
	ChangePassword(ctx context.Context, req *setting.ChangePasswordReq) (res *setting.ChangePasswordRes, err error)
	GetSystemConfig(ctx context.Context, req *setting.GetSystemConfigReq) (res *setting.GetSystemConfigRes, err error)
	UpdateSystemConfig(ctx context.Context, req *setting.UpdateSystemConfigReq) (res *setting.UpdateSystemConfigRes, err error)
}
