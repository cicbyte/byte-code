package service

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/platform"
)

type IPlatform interface {
	// 标签
	CreateTag(ctx context.Context, req *api.TagCreateReq) (id int, err error)
	UpdateTag(ctx context.Context, req *api.TagUpdateReq) (err error)
	DeleteTag(ctx context.Context, id int) (err error)
	ListTags(ctx context.Context) (res *api.TagListRes, err error)
	AttachTag(ctx context.Context, req *api.TagAttachReq) (err error)
	DetachTag(ctx context.Context, tagId int, entityType string, entityId int) (err error)
	GetTagEntities(ctx context.Context, tagId int) (res *api.TagEntitiesRes, err error)

	// 活动流
	ListActivities(ctx context.Context, req *api.ActivityListReq) (res *api.ActivityListRes, err error)

	// 通知
	ListNotifications(ctx context.Context, req *api.NotificationListReq) (res *api.NotificationListRes, err error)
	ReadNotification(ctx context.Context, id int) (err error)
	ReadAllNotifications(ctx context.Context) (err error)
	UnreadCount(ctx context.Context) (count int, err error)

	// 全局搜索
	Search(ctx context.Context, req *api.SearchReq) (res *api.SearchRes, err error)

	// 仪表盘
	DashboardStats(ctx context.Context) (res *api.DashboardStatsRes, err error)

	// 审计日志
	ListAuditLogs(ctx context.Context, req *api.AuditLogListReq) (res *api.AuditLogListRes, err error)
}

var localPlatform IPlatform

func Platform() IPlatform {
	if localPlatform == nil {
		panic("implement not found for interface IPlatform")
	}
	return localPlatform
}

func RegisterPlatform(i IPlatform) {
	localPlatform = i
}
