package perm

import (
	"context"

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

// CanAccessProject 资源访问判定：超级管理员或项目成员放行
func CanAccessProject(ctx context.Context, userId, projectId int) bool {
	return IsAdmin(ctx, userId) || IsProjectMember(ctx, userId, projectId)
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
		Where("role", "owner").
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
