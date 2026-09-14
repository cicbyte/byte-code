package perm

import (
	"context"
	"fmt"

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
