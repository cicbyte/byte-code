package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/cicbyte/byte-code/internal/service"
	"github.com/cicbyte/byte-code/utility/perm"
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

// MiddlewareCORS 跨域处理。凭证未使用（token 走自定义 header），
// 默认通配符源；如未来引入 cookie 会话，须在配置 cors.origins 里
// 给出显式域名列表（非 * 时按请求 Origin 匹配回显，且不再返回通配符）
func (s *sMiddleware) MiddlewareCORS(r *ghttp.Request) {
	if allowOrigin := corsAllowedOrigin(r); allowOrigin != "" {
		r.Response.Header().Set("Access-Control-Allow-Origin", allowOrigin)
	}
	r.Response.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	r.Response.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, token, X-Api-Key")
	r.Response.Header().Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")

	if r.Method == "OPTIONS" {
		r.Response.WriteHeader(200)
		return
	}

	r.Middleware.Next()
}

// corsAllowedOrigin 依据配置 cors.origins（默认 *）返回允许的 Origin 值；
// 未命中返回空串（不加 CORS 头，由浏览器拦截）
func corsAllowedOrigin(r *ghttp.Request) string {
	cfg, err := g.Cfg().Get(r.Context(), "cors.origins")
	if err != nil || cfg == nil {
		return "*"
	}
	origins := cfg.Strings()
	if len(origins) == 0 {
		return "*"
	}
	for _, o := range origins {
		if o == "*" {
			return "*"
		}
	}
	reqOrigin := r.Header.Get("Origin")
	if reqOrigin != "" {
		for _, o := range origins {
			if o == reqOrigin {
				return reqOrigin
			}
		}
	}
	return ""
}

// extractToken 依次从 token 头、Authorization Bearer、X-Api-Key 提取凭据。
// 只接受请求头：URL query 传 token 会进入访问日志/浏览器历史/代理日志
func extractToken(r *ghttp.Request) string {
	if t := r.Header.Get("token"); t != "" {
		return t
	}
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimSpace(auth[len("Bearer "):])
	}
	return r.Header.Get("X-Api-Key")
}

func (s *sMiddleware) MiddlewareTokenAuth(r *ghttp.Request) {
	tokenStr := extractToken(r)
	if tokenStr == "" {
		r.Response.WriteHeader(http.StatusUnauthorized)
		r.Response.Header().Set("Content-Type", "application/json")
		r.Response.Write(jsonStr(401, nil, "未登录或登录已过期"))
		r.ExitAll()
		return
	}
	// API Key 直认证：bc_ 前缀走 AI 用户 key 校验（含防爆破），外部 agent 免登录流程；
	// userId 入 ctx 后项目成员校验/审计日志与普通登录一致，key 天然标识来源 agent
	if strings.HasPrefix(tokenStr, "bc_") {
		aiUid, keyErr := service.AiUser().VerifyApiKey(r.Context(), tokenStr)
		if keyErr != nil {
			r.Response.WriteHeader(http.StatusUnauthorized)
			r.Response.Header().Set("Content-Type", "application/json")
			// 透传具体原因（锁定提示 vs key 无效），便于调用方区分
			r.Response.Write(jsonStr(401, nil, keyErr.Error()))
			r.ExitAll()
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, "userId", int(aiUid))
		r.SetCtx(ctx)
		r.Middleware.Next()
		return
	}

	userId, err := service.Auth().ValidateToken(r.Context(), tokenStr)
	if err != nil {
		r.Response.WriteHeader(http.StatusUnauthorized)
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

// MiddlewareAdminAuth 管理接口鉴权：仅超级管理员可访问，
// 防止普通登录用户调用系统配置、存储凭据、AI 用户管理等管理功能
func (s *sMiddleware) MiddlewareAdminAuth(r *ghttp.Request) {
	if !perm.IsAdmin(r.Context(), perm.UserId(r.Context())) {
		r.Response.WriteHeader(http.StatusOK)
		r.Response.Header().Set("Content-Type", "application/json")
		r.Response.Write(jsonStr(403, nil, "无权限访问该功能"))
		r.ExitAll()
		return
	}
	r.Middleware.Next()
}

// projectDirectRe /api/v1/projects/{projectId}/... 路径直接提取项目 id
var projectDirectRe = regexp.MustCompile(`^/api/v1/projects/(\d+)`)

// projectEntityRules 实体路径前缀 -> 含 project_id 列的表，
// 用于把 /tasks/{id} 这类顶层实体操作解析回所属项目做成员校验
var projectEntityRules = []struct {
	re    *regexp.Regexp
	table string
}{
	{regexp.MustCompile(`^/api/v1/tasks/(\d+)`), "tasks"},
	{regexp.MustCompile(`^/api/v1/sprints/(\d+)`), "sprints"},
	{regexp.MustCompile(`^/api/v1/requirements/(\d+)`), "requirements"},
	{regexp.MustCompile(`^/api/v1/test-cases/(\d+)`), "test_cases"},
	{regexp.MustCompile(`^/api/v1/test-plans/(\d+)`), "test_plans"},
	{regexp.MustCompile(`^/api/v1/docs/(\d+)`), "docs"},
	{regexp.MustCompile(`^/api/v1/db-tables/(\d+)`), "db_tables"},
	{regexp.MustCompile(`^/api/v1/schema-changes/(\d+)`), "schema_versions"},
}

// testPlanCaseRe /test-plan-cases/{id} 需两跳解析（test_plan_cases.test_plan_id -> test_plans.project_id）
var testPlanCaseRe = regexp.MustCompile(`^/api/v1/test-plan-cases/(\d+)`)

// resolveProjectId 从请求路径解析所属项目 id：直接项目路径、两跳计划用例、
// 或按实体前缀表查 project_id。非项目资源路径返回 0（不校验）。
func resolveProjectId(ctx context.Context, path string) int {
	if m := projectDirectRe.FindStringSubmatch(path); m != nil {
		id, _ := strconv.Atoi(m[1])
		return id
	}
	if m := testPlanCaseRe.FindStringSubmatch(path); m != nil {
		id, _ := strconv.Atoi(m[1])
		planId := perm.EntityFieldInt(ctx, "test_plan_cases", id, "test_plan_id")
		return perm.EntityProjectId(ctx, "test_plans", planId)
	}
	for _, rule := range projectEntityRules {
		if m := rule.re.FindStringSubmatch(path); m != nil {
			id, _ := strconv.Atoi(m[1])
			return perm.EntityProjectId(ctx, rule.table, id)
		}
	}
	return 0
}

// MiddlewareProjectAuth 项目资源归属校验：项目及其下属实体的读写仅对
// 超级管理员与 project_members 成员开放，防止登录用户跨项目越权操作
func (s *sMiddleware) MiddlewareProjectAuth(r *ghttp.Request) {
	uid := perm.UserId(r.Context())
	if uid == 0 {
		r.Response.WriteHeader(http.StatusUnauthorized)
		r.Response.Header().Set("Content-Type", "application/json")
		r.Response.Write(jsonStr(401, nil, "未登录或登录已过期"))
		r.ExitAll()
		return
	}
	projectId := resolveProjectId(r.Context(), r.URL.Path)
	if projectId == 0 {
		// 非项目资源路径，或目标记录不存在（后者由业务层返回不存在）
		r.Middleware.Next()
		return
	}
	if perm.CanAccessProject(r.Context(), uid, projectId) {
		r.Middleware.Next()
		return
	}
	r.Response.WriteHeader(http.StatusOK)
	r.Response.Header().Set("Content-Type", "application/json")
	r.Response.Write(jsonStr(403, nil, "无权限访问该项目资源"))
	r.ExitAll()
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
