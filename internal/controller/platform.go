package controller

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/platform"
	"github.com/cicbyte/byte-code/internal/service"
)

var PlatformCtrl = platformController{}

type platformController struct {
	BaseController
}

// ========== 标签 ==========

func (c *platformController) CreateTag(ctx context.Context, req *api.TagCreateReq) (res *api.TagCreateRes, err error) {
	id, err := service.Platform().CreateTag(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.TagCreateRes{Id: id}, nil
}

func (c *platformController) UpdateTag(ctx context.Context, req *api.TagUpdateReq) (res *api.TagUpdateRes, err error) {
	err = service.Platform().UpdateTag(ctx, req)
	return &api.TagUpdateRes{}, err
}

func (c *platformController) DeleteTag(ctx context.Context, req *api.TagDeleteReq) (res *api.TagDeleteRes, err error) {
	err = service.Platform().DeleteTag(ctx, req.Id)
	return &api.TagDeleteRes{}, err
}

func (c *platformController) ListTags(ctx context.Context, req *api.TagListReq) (res *api.TagListRes, err error) {
	return service.Platform().ListTags(ctx)
}

func (c *platformController) AttachTag(ctx context.Context, req *api.TagAttachReq) (res *api.TagAttachRes, err error) {
	err = service.Platform().AttachTag(ctx, req)
	return &api.TagAttachRes{}, err
}

func (c *platformController) DetachTag(ctx context.Context, req *api.TagDetachReq) (res *api.TagDetachRes, err error) {
	err = service.Platform().DetachTag(ctx, req.Id, req.EntityType, req.EntityId)
	return &api.TagDetachRes{}, err
}

func (c *platformController) GetTagEntities(ctx context.Context, req *api.TagEntitiesReq) (res *api.TagEntitiesRes, err error) {
	return service.Platform().GetTagEntities(ctx, req.Id)
}

// ========== 活动流 ==========

func (c *platformController) ListActivities(ctx context.Context, req *api.ActivityListReq) (res *api.ActivityListRes, err error) {
	return service.Platform().ListActivities(ctx, req)
}

// ========== 通知 ==========

func (c *platformController) ListNotifications(ctx context.Context, req *api.NotificationListReq) (res *api.NotificationListRes, err error) {
	return service.Platform().ListNotifications(ctx, req)
}

func (c *platformController) ReadNotification(ctx context.Context, req *api.NotificationReadReq) (res *api.NotificationReadRes, err error) {
	err = service.Platform().ReadNotification(ctx, req.Id)
	return &api.NotificationReadRes{}, err
}

func (c *platformController) ReadAllNotifications(ctx context.Context, req *api.NotificationReadAllReq) (res *api.NotificationReadAllRes, err error) {
	err = service.Platform().ReadAllNotifications(ctx)
	return &api.NotificationReadAllRes{}, err
}

func (c *platformController) UnreadCount(ctx context.Context, req *api.NotificationUnreadCountReq) (res *api.NotificationUnreadCountRes, err error) {
	count, err := service.Platform().UnreadCount(ctx)
	if err != nil {
		return nil, err
	}
	return &api.NotificationUnreadCountRes{Count: count}, nil
}

// ========== 全局搜索 ==========

func (c *platformController) Search(ctx context.Context, req *api.SearchReq) (res *api.SearchRes, err error) {
	return service.Platform().Search(ctx, req)
}

// ========== 仪表盘 ==========

func (c *platformController) DashboardStats(ctx context.Context, req *api.DashboardStatsReq) (res *api.DashboardStatsRes, err error) {
	return service.Platform().DashboardStats(ctx)
}

// ========== 审计日志 ==========

func (c *platformController) ListAuditLogs(ctx context.Context, req *api.AuditLogListReq) (res *api.AuditLogListRes, err error) {
	return service.Platform().ListAuditLogs(ctx, req)
}
