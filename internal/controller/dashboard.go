package controller

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/dashboard"
	service "github.com/cicbyte/byte-code/internal/service"
)

var DashboardCtrl = dashboardController{}

type dashboardController struct {
	BaseController
}

func (c *dashboardController) Console(ctx context.Context, req *api.ConsoleReq) (res *api.ConsoleRes, err error) {
	return service.Dashboard().Console(ctx)
}
