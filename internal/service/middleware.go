// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"github.com/gogf/gf/v2/net/ghttp"
)

type (
	IMiddleware interface {
		MiddlewareCORS(r *ghttp.Request)
		MiddlewareTokenAuth(r *ghttp.Request)
		// MiddlewareResponse 统一响应格式为 {code: 200, result: ..., message: "ok"}
		// 将 GoFrame 默认的 {code, message, data} 格式转换为前端期望的格式
		MiddlewareResponse(r *ghttp.Request)
		// MiddlewareAuditLog 审计日志中间件，拦截 POST/PUT/DELETE 请求记录操作
		MiddlewareAuditLog(r *ghttp.Request)
		// MiddlewareAdminAuth 管理接口鉴权，仅超级管理员可访问
		MiddlewareAdminAuth(r *ghttp.Request)
	}
)

var (
	localMiddleware IMiddleware
)

func Middleware() IMiddleware {
	if localMiddleware == nil {
		panic("implement not found for interface IMiddleware, forgot register?")
	}
	return localMiddleware
}

func RegisterMiddleware(i IMiddleware) {
	localMiddleware = i
}
