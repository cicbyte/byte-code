package middleware

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/cicbyte/byte-code/utility/usagewriter"
	"github.com/gogf/gf/v2/net/ghttp"
)

// MiddlewareUsageTrack 使用埋点：读+写全量记录 usage_events（异步批量落盘）。
// 与 MiddlewareAuditLog 的分工：审计只记写操作（合规留痕），本中间件记全部
// 请求（行为统计）——CLI 行为主体是读（tasks/notify/docs/memory），只审写
// 看不到使用画像。挂载在 TokenAuth 之后（需要 uid/clientType）。
func (s *sMiddleware) MiddlewareUsageTrack(r *ghttp.Request) {
	start := time.Now()

	// 请求体须在 Next 前取（GoFrame 缓冲后 handler 仍可读）；GET 无体
	var body []byte
	if r.Method == "POST" || r.Method == "PUT" {
		body = r.GetBody()
	}

	r.Middleware.Next()

	// 健康检查/探活不记：只有噪音没有分析价值
	if strings.Contains(r.URL.Path, "/health") {
		return
	}

	uid := perm.UserId(r.Context())
	if uid == 0 {
		return // 未认证请求（登录/token 校验失败的也有价值，但先保底只记有效会话）
	}

	usagewriter.Record(usagewriter.Event{
		ActorID:    uid,
		ActorType:  usageActorType(r.Context(), uid),
		Client:     usageClient(r),
		Method:     r.Method,
		Endpoint:   templateEndpoint(r.URL.Path),
		StatusCode: r.Response.Status,
		DurationMs: int(time.Since(start).Milliseconds()),
		ProjectId:  resolveProjectId(r.Context(), r.URL.Path),
		SessionId:  sessionIdOf(r),
		ErrorCode:  respErrorCode(r.Response.Buffer()),
		Params:     extractParams(r, body),
	})
}

// usageClient 客户端识别：bc_ API key（agent/CLI）→ cli；UA 带 bcode 也归 cli；
// 其余 web。未来 CLI 版本 UA（bcode/x.y.z）天然被覆盖
func usageClient(r *ghttp.Request) string {
	if strings.HasPrefix(r.GetHeader("token"), "bc_") {
		return "cli"
	}
	if strings.Contains(strings.ToLower(r.UserAgent()), "bcode") {
		return "cli"
	}
	return "web"
}

// sessionIdOf agent 会话头（bcsh_），人类侧为空
func sessionIdOf(r *ghttp.Request) string {
	s := r.GetHeader("X-Session")
	if len(s) > 64 {
		s = s[:64]
	}
	return s
}

// respErrorCode 从 GoFrame 响应壳 {code,...} 提取业务码（0=成功；
// 解析失败按 HTTP 状态兜底：5xx 记 500）
func respErrorCode(buf []byte) int {
	if len(buf) == 0 {
		return 0
	}
	var resp struct {
		Code int `json:"code"`
	}
	if json.Unmarshal(buf, &resp) == nil {
		return resp.Code
	}
	return 0
}

// usageActorType 复用审计侧口径（sys_users.type=ai → ai）
func usageActorType(ctx context.Context, userId int) string {
	return userActorType(ctx, userId)
}

// templateEndpoint 路径模板化：/api/v1/tasks/51/comments → /v1/tasks/{id}/comments。
// 数字段替换为 {id} 防维度爆炸；剥掉 /api 前缀统一口径
func templateEndpoint(path string) string {
	p := strings.TrimPrefix(path, "/api")
	parts := strings.Split(strings.Trim(p, "/"), "/")
	for i, seg := range parts {
		if seg == "" {
			continue
		}
		if isNumeric(seg) {
			parts[i] = "{id}"
		}
	}
	return "/" + strings.Join(parts, "/")
}

// ==================== 参数抽取 ====================
//
// 原则：只进白名单，大文本只记长度，未知键只记键名。
// 目的是防两类事故——隐私（记忆 value/反馈正文/artifacts 入统计表）与
// 基数爆炸（keyword 任意值、任意路径）。

// 枚举/数值类：值本身可聚合，直接记值
var paramValueWhitelist = map[string]bool{
	// 通用
	"page": true, "size": true, "status": true, "type": true, "priority": true,
	"scope": true, "action": true, "state": true, "role": true,
	// 任务
	"assigneeId": true, "sprintId": true, "requirementId": true, "tagId": true,
	"dueDate": true, "milestoneId": true, "parentId": true,
	// 反馈/QA/专题
	"source": true, "sourceTaskId": true, "convertedTaskId": true, "phaseId": true,
	// 文档/记忆
	"recursive": true, "verified": true, "ttl": true, "include": true,
}

// 大文本类：只记 长度（_len 后缀）
var paramLenOnly = map[string]bool{
	"title": true, "description": true, "content": true, "answer": true,
	"detail": true, "artifacts": true, "note": true, "reason": true,
	"comment": true, "question": true, "value": true, "explain": true,
	"summary": true, "detailText": true,
}

// 自由文本 query 键：只记 长度
var queryLenOnly = map[string]bool{
	"keyword": true, "q": true, "path": true, "prefix": true, "name": true,
}

func extractParams(r *ghttp.Request, body []byte) string {
	m := map[string]interface{}{}
	var otherKeys []string

	// query：白名单记值，自由文本记长度，其余记键名
	for k, vs := range r.URL.Query() {
		v := ""
		if len(vs) > 0 {
			v = vs[0]
		}
		if len(v) > 64 {
			v = v[:64]
		}
		switch {
		case paramValueWhitelist[k]:
			m[k] = coerceNum(v)
		case queryLenOnly[k]:
			m[k+"_len"] = len(v)
		default:
			otherKeys = append(otherKeys, k)
		}
	}

	// body：JSON 白名单小字段记值，大文本记长度，其余记键名
	if len(body) > 0 && len(body) < 64*1024 {
		var parsed map[string]interface{}
		if json.Unmarshal(body, &parsed) == nil {
			for k, v := range parsed {
				switch {
				case paramValueWhitelist[k]:
					m[k] = jsonScalar(v)
				case paramLenOnly[k]:
					m[k+"_len"] = lenScalar(v)
				default:
					otherKeys = append(otherKeys, k)
				}
			}
		}
	}

	if len(otherKeys) > 0 {
		// 去重 + 上限，防未知键自身爆维度
		seen := map[string]bool{}
		uniq := otherKeys[:0]
		for _, k := range otherKeys {
			if !seen[k] && len(k) < 64 {
				seen[k] = true
				uniq = append(uniq, k)
			}
		}
		m["other_keys"] = uniq
	}
	if len(m) == 0 {
		return ""
	}
	b, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(b)
}

// coerceNum 数字字符串保持数值形态（page=2 → 2），便于聚合
func coerceNum(s string) interface{} {
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return s
}

func jsonScalar(v interface{}) interface{} {
	switch t := v.(type) {
	case string:
		if len(t) > 64 {
			return t[:64]
		}
		return t
	case float64, bool, int:
		return t
	default:
		return "<obj>"
	}
}

func lenScalar(v interface{}) int {
	switch t := v.(type) {
	case string:
		return len(t)
	default:
		return -1
	}
}
