package perm

import (
	"context"
	"fmt"
	"strings"

	"github.com/cicbyte/byte-code/internal/consts"
	"github.com/gogf/gf/v2/frame/g"
)

// AdminRoleId 超级管理员角色 id（种子数据固定为 1）
const AdminRoleId = 1

// UserId 从请求上下文取当前登录用户 id（TokenAuth 中间件注入），未登录返回 0
func UserId(ctx context.Context) int {
	if v, ok := ctx.Value("userId").(int); ok {
		return v
	}
	return 0
}

// IsAdmin 判断用户是否拥有超级管理员角色
func IsAdmin(ctx context.Context, userId int) bool {
	count, err := g.DB().Model("sys_user_roles ur").
		InnerJoin("sys_roles r", "ur.role_id = r.id").
		Where("ur.user_id", userId).
		Where("r.id", AdminRoleId).
		Count()
	return err == nil && count > 0
}

// RequireMenuPerm 菜单权限字典校验（PRD design/permission-system-prd.md §3.2）：
// 超管直通；否则查 sys_user_roles→sys_role_menus→sys_menus(name=key)。
// 通过返回 nil，拒绝返回带权限显示名的错误。挂检点在业务方法顶部，
// 风格同项目侧 IsProjectOwner 检查
func RequireMenuPerm(ctx context.Context, key string) error {
	uid := UserId(ctx)
	if IsAdmin(ctx, uid) {
		return nil
	}
	v, err := g.DB().Model("sys_menus m").
		InnerJoin("sys_role_menus rm", "m.id = rm.menu_id").
		InnerJoin("sys_user_roles ur", "rm.role_id = ur.role_id").
		Where("ur.user_id", uid).
		Where("m.name", key).
		Where("m.status", 1).
		Fields("m.title").
		Value()
	if err == nil && v != nil {
		return nil
	}
	title := key
	if t, terr := g.DB().Model("sys_menus").Where("name", key).Fields("title").Value(); terr == nil && t != nil {
		title = t.String()
	}
	return fmt.Errorf("无权限：需管理员授予「%s」", title)
}

// IsProjectMember 判断用户是否为指定项目的成员（project_members）
func IsProjectMember(ctx context.Context, userId, projectId int) bool {
	if userId <= 0 || projectId <= 0 {
		return false
	}
	count, err := g.DB().Model("project_members").
		Where("user_id", userId).
		Where("project_id", projectId).
		Count()
	return err == nil && count > 0
}

// CanAccessProject 资源访问判定：超级管理员、项目成员或已准入的外部 Agent 放行
func CanAccessProject(ctx context.Context, userId, projectId int) bool {
	return IsAdmin(ctx, userId) || IsProjectMember(ctx, userId, projectId) ||
		IsAgentBound(ctx, userId, projectId)
}

// IsAgentBound 外部 Agent 的项目准入（agent_project_bindings）。
// agent_id 全局唯一且仅指向 type=ai 账号，人类用户 id 不会命中
func IsAgentBound(ctx context.Context, userId, projectId int) bool {
	if userId <= 0 || projectId <= 0 {
		return false
	}
	count, err := g.DB().Model("agent_project_bindings").
		Where("agent_id", userId).
		Where("project_id", projectId).
		Count()
	return err == nil && count > 0
}

// AgentCaps Agent 能力集字典（PRD §5.1）：key → 显示名。
// 绑定表 capabilities 列存逗号分隔 key；空 = 全部能力（存量绑定兼容）
var AgentCaps = map[string]string{
	"tasks_read":   "读任务",
	"tasks_write":  "写任务",
	"docs_read":    "读文档",
	"docs_write":   "写文档",
	"memory_read":  "读记忆",
	"memory_write": "写记忆",
	"feedback":     "投递反馈",
	"qa":           "维护问答库",
	"test_execute": "上报测试执行",
	"discuss":      "参与讨论",
}

// AgentRequire Agent 能力门禁（PRD §5.2）：人类调用直通（受角色/成员体系
// 约束）；agent 按 agent_project_bindings.capabilities 校验。空能力集视为
// 全能力（只有管理侧显式设置过能力集的 agent 才受限；NULL 与空串同义——
// 迁移 69 之前的存量绑定 capabilities 为 NULL，v0.3.0 曾把 NULL 误判成
// 无绑定行导致存量 agent 被锁，v0.3.1 修正）。无绑定行不再放行
// （#443）：仅当存在 project_members 行（早期手工加的存量 agent）按成员
// 资格放行，否则明确拒绝——未接入项目的 agent 不应能操作其任务。
// 调整即时生效（无缓存）
func AgentRequire(ctx context.Context, projectId int, cap string) error {
	uid := UserId(ctx)
	if uid == 0 {
		return nil // 未认证请求由外层 TokenAuth/协议认证兜底
	}
	v, err := g.DB().Model("sys_users").Where("id", uid).Fields("type").Value()
	if err != nil || v == nil || v.String() != "ai" {
		return nil
	}
	// 全局资源（project_id=0，如全局记忆读）：无「项目 0 的绑定」概念，
	// 以前靠无行兜底放行（#443 关洞时误伤 agent 全局读，此处显式豁免）；
	// 全局写操作由管理路由 AdminAuth 把关，不经过这里
	if projectId <= 0 {
		return nil
	}
	// COALESCE 把 NULL 归一为空串；One() 的空行才真正代表「无绑定行」——
	// Value() 的 nil 同时覆盖两种情况，无法区分（v0.3.0 回归根因）
	row, cerr := g.DB().Model("agent_project_bindings").
		Where("agent_id", uid).
		Where("project_id", projectId).
		Fields("COALESCE(capabilities, '') AS capabilities").One()
	if cerr != nil {
		// 查询失败 fail-closed：静默放行会让门禁形同虚设
		return fmt.Errorf("校验 Agent 能力失败: %w", cerr)
	}
	if row.IsEmpty() {
		// 无绑定行：members 行兼容存量手工 agent（全能力），两边都没有 = 未接入
		cnt, merr := g.DB().Model("project_members").
			Where("user_id", uid).
			Where("project_id", projectId).
			Count()
		if merr == nil && cnt > 0 {
			return nil
		}
		return fmt.Errorf("无权限：Agent 未接入该项目，请向 project owner 申请接入码加入")
	}
	list := strings.TrimSpace(row["capabilities"].String())
	if list == "" {
		return nil
	}
	for _, c := range strings.Split(list, ",") {
		if strings.TrimSpace(c) == cap {
			return nil
		}
	}
	label := AgentCaps[cap]
	if label == "" {
		label = cap
	}
	return fmt.Errorf("无权限：Agent 未被授予「%s」能力", label)
}

// SessionProjectId 认证阶段（TokenAuth bc_ 分支）解析的会话项目：
// X-Session（bcsh_，键=agent+project）。0=未携带会话（人类 token 或
// 无会话直连 API），此时不做会话级项目约束
func SessionProjectId(ctx context.Context) int {
	if v := ctx.Value("sessionProjectId"); v != nil {
		if p, ok := v.(int); ok {
			return p
		}
	}
	return 0
}

// AgentSessionGuard 「当前项目」约束（#443）：agent 请求携带会话时，任务类
// 操作的目标项目必须与会话项目一致——同一 agent 多项目绑定时防跨会话/
// 跨项目错领（在 A 项目会话里认领/操作 B 项目任务）。人类与无会话请求
// 不约束（后者由 AgentRequire 的准入门禁兜底）
func AgentSessionGuard(ctx context.Context, projectId int) error {
	uid := UserId(ctx)
	if uid == 0 {
		return nil
	}
	sp := SessionProjectId(ctx)
	if sp == 0 {
		return nil
	}
	v, err := g.DB().Model("sys_users").Where("id", uid).Fields("type").Value()
	if err != nil || v == nil || v.String() != "ai" {
		return nil
	}
	if sp != projectId {
		return fmt.Errorf("该任务属于其它项目：当前会话绑定项目 %d，请在对应该项目的目录/会话中操作", sp)
	}
	return nil
}

// AgentTaskGate 任务类操作门禁 = 准入/能力（AgentRequire）+ 会话项目约束
// （AgentSessionGuard）。任务域写操作统一走这里
func AgentTaskGate(ctx context.Context, projectId int, cap string) error {
	if err := AgentRequire(ctx, projectId, cap); err != nil {
		return err
	}
	return AgentSessionGuard(ctx, projectId)
}

// IsProjectOwner 判断用户是否为项目 owner（或超管）——
// 删除项目、管理成员等管理级操作的准入条件
func IsProjectOwner(ctx context.Context, userId, projectId int) bool {
	if IsAdmin(ctx, userId) {
		return true
	}
	count, err := g.DB().Model("project_members").
		Where("user_id", userId).
		Where("project_id", projectId).
		Where("role", consts.MemberRoleOwner).
		Count()
	return err == nil && count > 0
}

// IsProjectMaintainer 项目管理级判定（PRD §4）：owner、maintainer 或超管。
// 成员管理/接入码/关联与分组治理等管理入口走此档；转交负责人、删除项目
// 等危险操作仍用 IsProjectOwner
func IsProjectMaintainer(ctx context.Context, userId, projectId int) bool {
	if IsAdmin(ctx, userId) {
		return true
	}
	count, err := g.DB().Model("project_members").
		Where("user_id", userId).
		Where("project_id", projectId).
		WhereIn("role", []string{consts.MemberRoleOwner, consts.MemberRoleMaintainer}).
		Count()
	return err == nil && count > 0
}

// EntityFieldInt 取实体表指定整型列的值，记录不存在返回 0
func EntityFieldInt(ctx context.Context, table string, entityId int, column string) int {
	v, _ := g.DB().Model(table).Where("id", entityId).Fields(column).Value()
	if v == nil {
		return 0
	}
	return v.Int()
}

// EntityProjectId 解析实体所属项目 id（实体表须有 project_id 列），记录不存在返回 0
func EntityProjectId(ctx context.Context, table string, entityId int) int {
	return EntityFieldInt(ctx, table, entityId, "project_id")
}
