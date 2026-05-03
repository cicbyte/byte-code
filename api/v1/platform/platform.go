package platform

import "github.com/gogf/gf/v2/frame/g"

// 标签
type TagCreateReq struct {
	g.Meta `path:"/tags" method:"post" tags:"标签" summary:"创建标签"`
	Name   string `json:"name" v:"required#标签名不能为空"`
	Color  string `json:"color" v:"required#颜色不能为空"`
}

type TagCreateRes struct {
	Id int `json:"id"`
}

type TagUpdateReq struct {
	g.Meta `path:"/tags/{id}" method:"put" tags:"标签" summary:"更新标签"`
	Id     int    `json:"id" v:"required" in:"path"`
	Name   string `json:"name"`
	Color  string `json:"color"`
}

type TagUpdateRes struct {
	g.Meta `mime:"application/json"`
}

type TagDeleteReq struct {
	g.Meta `path:"/tags/{id}" method:"delete" tags:"标签" summary:"删除标签"`
	Id     int `json:"id" v:"required" in:"path"`
}

type TagDeleteRes struct {
	g.Meta `mime:"application/json"`
}

type TagListReq struct {
	g.Meta `path:"/tags" method:"get" tags:"标签" summary:"标签列表"`
}

type TagListRes struct {
	List []TagItem `json:"list"`
}

type TagItem struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	CreatorId int    `json:"creatorId"`
	CreatedAt string `json:"createdAt"`
}

type TagAttachReq struct {
	g.Meta      `path:"/tags/{id}/attach" method:"post" tags:"标签" summary:"给实体打标签"`
	Id          int    `json:"id" v:"required" in:"path"`
	EntityType  string `json:"entityType" v:"required|in:task,requirement,test_case#实体类型不能为空"`
	EntityId    int    `json:"entityId" v:"required#实体ID不能为空"`
}

type TagAttachRes struct {
	g.Meta `mime:"application/json"`
}

type TagDetachReq struct {
	g.Meta      `path:"/tags/{id}/detach" method:"delete" tags:"标签" summary:"移除实体标签"`
	Id          int    `json:"id" v:"required" in:"path"`
	EntityType  string `json:"entityType" v:"required" in:"query"`
	EntityId    int    `json:"entityId" v:"required" in:"query"`
}

type TagDetachRes struct {
	g.Meta `mime:"application/json"`
}

type TagEntitiesReq struct {
	g.Meta `path:"/tags/{id}/entities" method:"get" tags:"标签" summary:"按标签查询实体"`
	Id     int `json:"id" v:"required" in:"path"`
}

type TagEntitiesRes struct {
	Tasks        []int `json:"tasks"`
	Requirements []int `json:"requirements"`
	TestCases    []int `json:"testCases"`
}

// 活动
type ActivityListReq struct {
	g.Meta    `path:"/activities" method:"get" tags:"活动流" summary:"活动流列表"`
	Module    string `json:"module" in:"query"`
	ActorId   int    `json:"actorId" in:"query"`
	ProjectId int    `json:"projectId" in:"query"`
	Page      int    `json:"page" in:"query" d:"1"`
	Size      int    `json:"size" in:"query" d:"20"`
}

type ActivityListRes struct {
	List  []ActivityItem `json:"list"`
	Total int            `json:"total"`
}

type ActivityItem struct {
	Id         int    `json:"id"`
	ActorId    int    `json:"actorId"`
	ActorType  string `json:"actorType"`
	ActorName  string `json:"actorName"`
	Action     string `json:"action"`
	TargetType string `json:"targetType"`
	TargetId   int    `json:"targetId"`
	TargetName string `json:"targetName"`
	ProjectId  int    `json:"projectId"`
	Detail     string `json:"detail"`
	CreatedAt  string `json:"createdAt"`
}

// 通知
type NotificationListReq struct {
	g.Meta `path:"/notifications" method:"get" tags:"通知" summary:"通知列表"`
	Unread int `json:"unread" in:"query"`
	Page   int `json:"page" in:"query" d:"1"`
	Size   int `json:"size" in:"query" d:"20"`
}

type NotificationListRes struct {
	List  []NotificationItem `json:"list"`
	Total int                `json:"total"`
}

type NotificationItem struct {
	Id         int    `json:"id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Type       string `json:"type"`
	IsRead     int    `json:"isRead"`
	SourceType string `json:"sourceType"`
	SourceId   int    `json:"sourceId"`
	CreatedAt  string `json:"createdAt"`
}

type NotificationReadReq struct {
	g.Meta `path:"/notifications/{id}/read" method:"put" tags:"通知" summary:"标记已读"`
	Id     int `json:"id" v:"required" in:"path"`
}

type NotificationReadRes struct {
	g.Meta `mime:"application/json"`
}

type NotificationReadAllReq struct {
	g.Meta `path:"/notifications/read-all" method:"put" tags:"通知" summary:"全部已读"`
}

type NotificationReadAllRes struct {
	g.Meta `mime:"application/json"`
}

type NotificationUnreadCountReq struct {
	g.Meta `path:"/notifications/unread-count" method:"get" tags:"通知" summary:"未读数量"`
}

type NotificationUnreadCountRes struct {
	Count int `json:"count"`
}

// 全局搜索
type SearchReq struct {
	g.Meta `path:"/search" method:"get" tags:"搜索" summary:"全局搜索"`
	Q      string `json:"q" in:"query" v:"required#关键词不能为空"`
	Module string `json:"module" in:"query"`
	Page   int    `json:"page" in:"query" d:"1"`
	Size   int    `json:"size" in:"query" d:"20"`
}

type SearchRes struct {
	List  []SearchResult `json:"list"`
	Total int            `json:"total"`
}

type SearchResult struct {
	Module  string `json:"module"`
	Id      int    `json:"id"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

// 仪表盘
type DashboardStatsReq struct {
	g.Meta `path:"/stats/dashboard" method:"get" tags:"统计" summary:"全局仪表盘"`
}

type DashboardStatsRes struct {
	TotalRequirements int              `json:"totalRequirements"`
	TotalTasks        int              `json:"totalTasks"`
	InProgressTasks   int              `json:"inProgressTasks"`
	ReviewTasks       int              `json:"reviewTasks"`
	TestPassRate      float64          `json:"testPassRate"`
	AiStats           []AiStatItem     `json:"aiStats"`
	RecentTasks       []RecentTaskItem `json:"recentTasks"`
}

type AiStatItem struct {
	AiName    string `json:"aiName"`
	TaskCount int    `json:"taskCount"`
}

type RecentTaskItem struct {
	Id        int    `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	UpdatedAt string `json:"updatedAt"`
}

// 审计日志
type AuditLogListReq struct {
	g.Meta     `path:"/admin/audit-logs" method:"get" tags:"审计日志" summary:"审计日志列表"`
	TargetType string `json:"targetType" in:"query"`
	ActorId    int    `json:"actorId" in:"query"`
	Action     string `json:"action" in:"query"`
	ProjectId  int    `json:"projectId" in:"query"`
	Page       int    `json:"page" in:"query" d:"1"`
	Size       int    `json:"size" in:"query" d:"20"`
}

type AuditLogListRes struct {
	List  []AuditLogItem `json:"list"`
	Total int            `json:"total"`
}

type AuditLogItem struct {
	Id         int    `json:"id"`
	ActorId    int    `json:"actorId"`
	ActorType  string `json:"actorType"`
	Action     string `json:"action"`
	TargetType string `json:"targetType"`
	TargetId   int    `json:"targetId"`
	TargetName string `json:"targetName"`
	Changes    string `json:"changes"`
	IpAddress  string `json:"ipAddress"`
	ProjectId  int    `json:"projectId"`
	CreatedAt  string `json:"createdAt"`
}
