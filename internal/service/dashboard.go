package service

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/dashboard"
)

type IDashboard interface {
	Console(ctx context.Context) (res *api.ConsoleRes, err error)
}

var localDashboard IDashboard

func Dashboard() IDashboard {
	if localDashboard == nil {
		panic("implement not found for interface IDashboard, forgot register?")
	}
	return localDashboard
}

func RegisterDashboard(i IDashboard) {
	localDashboard = i
}
