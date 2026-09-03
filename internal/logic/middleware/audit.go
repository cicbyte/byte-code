package middleware

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/cicbyte/byte-code/utility/auditwriter"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// MiddlewareAuditLog 审计写操作（POST/PUT/DELETE）：异步批量落盘 audit_logs。
// 挂载在 TokenAuth 之后；actor 类型取 sys_users.type。
// 项目归属必须在请求执行前解析——删除类操作执行后实体已不存在，事后解析必为空。
func (s *sMiddleware) MiddlewareAuditLog(r *ghttp.Request) {
	path := r.URL.Path
	projectId := resolveProjectId(r.Context(), path)

	r.Middleware.Next()

	if r.Method != "POST" && r.Method != "PUT" && r.Method != "DELETE" {
		return
	}

	uid := perm.UserId(r.Context())
	if uid == 0 {
		return
	}

	action := "create"
	if r.Method == "PUT" {
		action = "update"
	} else if r.Method == "DELETE" {
		action = "delete"
	}

	var target targetInfo
	if r.Method == "POST" {
		// POST 路径中的 id 是父资源（如 /projects/5/tasks 的 5 是项目），不能当作
		// 操作目标；目标类型取路径最后的资源段，新实体 id 从响应体 data.id 补全
		target = targetInfo{entityType: lastResourceSegment(path)}
		if target.entityId = createdIdFromBody(r.Response.Buffer()); target.entityType == "projects" {
			// 创建项目时项目即目标，归属项目就是它自己
			projectId = target.entityId
		}
	} else {
		target = parseTarget(path)
	}

	enrichVaultTarget(r, &target, path)

	auditwriter.Record(auditwriter.Entry{
		ActorID:    uid,
		ActorType:  userActorType(r.Context(), uid),
		Action:     action,
		TargetType: target.entityType,
		TargetID:   target.entityId,
		TargetName: target.name,
		IpAddress:  r.GetClientIp(),
		UserAgent:  r.UserAgent(),
		ProjectId:  projectId,
	})
}

// userActorType 操作者类型：AI 账号（sys_users.type=ai）记为 ai，其余 human
func userActorType(ctx context.Context, userId int) string {
	if v, _ := g.DB().Model("sys_users").Where("id", userId).Fields("type").Value(); v != nil && v.String() == "ai" {
		return "ai"
	}
	return "human"
}

// createdIdFromBody 从 GoFrame 默认响应 {code,message,data:{id}} 中提取新建实体 id
func createdIdFromBody(buf []byte) int {
	if len(buf) == 0 {
		return 0
	}
	var resp struct {
		Code int `json:"code"`
		Data struct {
			Id int `json:"id"`
		} `json:"data"`
	}
	if json.Unmarshal(buf, &resp) != nil || resp.Code != 0 {
		return 0
	}
	return resp.Data.Id
}

// lastResourceSegment 取路径最后的资源段作为目标类型：/api/v1/projects -> projects，
// /api/v1/projects/5/tasks -> tasks（末段为数字时取前一段）
func lastResourceSegment(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if !isNumeric(parts[i]) {
			return parts[i]
		}
	}
	return ""
}

type targetInfo struct {
	entityType string
	entityId   int
	name       string // vault/memories 等路径型资源的目标明细（文件路径 / 记忆 key）
}

// enrichVaultTarget vault/memories 资源的目标明细解析：
//   /projects/{pid}/vault/file?path=a/b.md    → vault + name=a/b.md
//   /projects/{pid}/memories/{key}/verify     → memories + name={key}
//   /projects/{pid}/vault/upload?path=dir     → vault + name=dir
// 目标类型取资源段后（file/meta/move 等动作词不作为类型）
func enrichVaultTarget(r *ghttp.Request, t *targetInfo, path string) {
	parts := strings.Split(strings.TrimPrefix(path, "/api/v1/"), "/")
	// projects/{pid}/vault|memories/...
	if len(parts) < 3 || parts[0] != "projects" {
		return
	}
	switch parts[2] {
	case "vault":
		t.entityType = "vault"
	case "memories":
		t.entityType = "memories"
	default:
		return
	}
	// 明细优先取 query（vault 的 path；memories 的 key 在路径段）
	if q := r.GetQuery("path").String(); q != "" {
		t.name = q
		return
	}
	if t.entityType == "memories" && len(parts) >= 4 {
		// memories/{key} 或 memories/{key}/verify|expire
		if p := parts[3]; !isNumeric(p) {
			t.name = p
		}
	}
}

// postEntityVerbs 紧跟实体 id 出现的动作词：审计目标仍归父实体
var postEntityVerbs = map[string]bool{
	"claim":    true,
	"burndown": true,
}

func parseTarget(path string) targetInfo {
	// /api/v1/projects/123/tasks/456
	// /api/v1/tasks/456
	// /api/v1/projects/1/memories/agent.note   ← 数字段（项目）后跟子资源
	parts := strings.Split(strings.TrimPrefix(path, "/api/v1/"), "/")
	info := targetInfo{}

	// 取最后一个数字作为 entityId
	lastNumIdx := -1
	for i := len(parts) - 1; i >= 0; i-- {
		if isNumeric(parts[i]) {
			info.entityId, _ = strconv.Atoi(parts[i])
			lastNumIdx = i
			break
		}
	}

	// 数字段后还有资源段时，该段才是真正的操作目标（memories/vault），
	// 数字段本身只是其父容器（projectId）；但 claim/burndown 这类紧跟实体 id 的
	// 动作词例外——操作目标仍是父实体（tasks/sprints）
	if lastNumIdx >= 0 && lastNumIdx+1 < len(parts) && !postEntityVerbs[parts[lastNumIdx+1]] {
		info.entityType = parts[lastNumIdx+1]
		return info
	}

	for i := 0; i < len(parts)-1; i++ {
		// 如果下一部分是数字，当前部分是实体类型
		if isNumeric(parts[i+1]) && i%2 == 0 {
			info.entityType = parts[i]
			// 取最后一个匹配
		}
	}

	return info
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}
