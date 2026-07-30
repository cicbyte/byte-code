package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/cicbyte/byte-code/internal/service"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func init() {
	service.RegisterMiddleware(New())
}

func New() *sMiddleware {
	return &sMiddleware{}
}

type sMiddleware struct{}

func (s *sMiddleware) MiddlewareCORS(r *ghttp.Request) {
	r.Response.Header().Set("Access-Control-Allow-Origin", "*")
	r.Response.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	r.Response.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, token")
	r.Response.Header().Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
	r.Response.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == "OPTIONS" {
		r.Response.WriteHeader(200)
		return
	}

	r.Middleware.Next()
}

func (s *sMiddleware) MiddlewareTokenAuth(r *ghttp.Request) {
	tokenStr := r.Header.Get("token")
	if tokenStr == "" {
		tokenStr = r.Get("token").String()
	}
	if tokenStr == "" {
		r.Response.WriteHeader(http.StatusOK)
		r.Response.Header().Set("Content-Type", "application/json")
		r.Response.Write(jsonStr(401, nil, "未登录或登录已过期"))
		r.ExitAll()
		return
	}
	userId, err := service.Auth().ValidateToken(r.Context(), tokenStr)
	if err != nil {
		r.Response.WriteHeader(http.StatusOK)
		r.Response.Header().Set("Content-Type", "application/json")
		r.Response.Write(jsonStr(401, nil, "登录已过期，请重新登录"))
		r.ExitAll()
		return
	}
	// 仍在使用初始默认密码的账号只放行改密/登出/用户信息接口，防止默认口令账号被直接使用
	mustChange, _ := g.DB().Model("sys_users").Where("id", userId).Fields("must_change_password").Value()
	if mustChange != nil && mustChange.Int() == 1 && !passwordChangeAllowed(r.URL.Path) {
		r.Response.WriteHeader(http.StatusOK)
		r.Response.Header().Set("Content-Type", "application/json")
		r.Response.Write(jsonStr(1001, nil, "首次登录请先修改初始密码"))
		r.ExitAll()
		return
	}
	ctx := r.Context()
	ctx = context.WithValue(ctx, "userId", userId)
	ctx = context.WithValue(ctx, "token", tokenStr)
	r.SetCtx(ctx)
	r.Middleware.Next()
}

// passwordChangeAllowed 强制改密状态下仍可访问的接口：修改密码、登出、获取用户信息
func passwordChangeAllowed(path string) bool {
	switch path {
	case "/api/account/password", "/api/login/logout", "/api/admin_info":
		return true
	}
	return false
}

// MiddlewareResponse 统一响应格式为 {code: 200, result: ..., message: "ok"}
// 将 GoFrame 默认的 {code, message, data} 格式转换为前端期望的格式
func (s *sMiddleware) MiddlewareResponse(r *ghttp.Request) {
	r.Middleware.Next()

	// 获取 GoFrame 写入的 buffer
	buffer := r.Response.Buffer()
	if len(buffer) == 0 {
		return
	}

	// 尝试解析 GoFrame 默认格式
	var gfResp struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(buffer, &gfResp); err != nil {
		// 不是 JSON 或格式不对，保持原样
		return
	}

	// 已经是目标格式则跳过
	var check map[string]interface{}
	if err := json.Unmarshal(buffer, &check); err == nil {
		if _, ok := check["result"]; ok {
			return
		}
	}

	// 转换: code=0 → 200, data → result
	code := 200
	if gfResp.Code != 0 {
		code = gfResp.Code
	}
	msg := "ok"
	if gfResp.Message != "" {
		msg = gfResp.Message
	}

	// 如果 data 是 null，设为 nil
	var result interface{} = gfResp.Data
	if string(gfResp.Data) == "null" {
		result = nil
	}

	resp, _ := json.Marshal(map[string]interface{}{
		"code":    code,
		"result":  result,
		"message": msg,
	})
	r.Response.ClearBuffer()
	r.Response.Write(resp)
}

func jsonStr(code int, result interface{}, message string) []byte {
	resp, _ := json.Marshal(map[string]interface{}{
		"code":    code,
		"result":  result,
		"message": message,
	})
	return resp
}
