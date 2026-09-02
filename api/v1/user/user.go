package user

import "github.com/gogf/gf/v2/frame/g"

// ==================== 用户管理（仅超管） ====================

type ListReq struct {
	g.Meta `path:"/admin/users" method:"get" tags:"用户管理" summary:"用户列表"`
	Keyword string `json:"keyword" in:"query"`
	Page    int    `json:"page" in:"query" d:"1"`
	Size    int    `json:"size" in:"query" d:"20"`
}

type ListRes struct {
	List  []Item `json:"list"`
	Total int    `json:"total"`
	Page  int    `json:"page"`
	Size  int    `json:"size"`
}

type Item struct {
	Id        int    `json:"id"`
	Username  string `json:"username"`
	RealName  string `json:"realName"`
	Email     string `json:"email"`
	Type      string `json:"type"`
	Status    int    `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type CreateReq struct {
	g.Meta   `path:"/admin/users" method:"post" tags:"用户管理" summary:"创建用户"`
	Username string `json:"username" v:"required|length:3,30#用户名不能为空|用户名长度3-30位"`
	Password string `json:"password" v:"required|length:8,20#密码不能为空|密码长度8-20位"`
	RealName string `json:"realName"`
	Email    string `json:"email" v:"email#邮箱格式不合法"`
	RoleIds  []int  `json:"roleIds"`
}

type CreateRes struct {
	Id int `json:"id"`
}

type UpdateReq struct {
	g.Meta   `path:"/admin/users/{id}" method:"put" tags:"用户管理" summary:"更新用户"`
	Id       int     `json:"id" v:"required" in:"path"`
	RealName *string `json:"realName"`
	Email    *string `json:"email"`
	Status   *int    `json:"status" v:"in:0,1#状态必须是0或1"`
	RoleIds  []int   `json:"roleIds"`
}

type UpdateRes struct {
	g.Meta `mime:"application/json"`
}

type ResetPasswordReq struct {
	g.Meta      `path:"/admin/users/{id}/reset-password" method:"put" tags:"用户管理" summary:"重置用户密码"`
	Id          int    `json:"id" v:"required" in:"path"`
	NewPassword string `json:"newPassword" v:"required|length:8,20#新密码不能为空|密码长度8-20位"`
}

type ResetPasswordRes struct {
	g.Meta `mime:"application/json"`
}

type DeleteReq struct {
	g.Meta `path:"/admin/users/{id}" method:"delete" tags:"用户管理" summary:"删除用户"`
	Id     int `json:"id" v:"required" in:"path"`
}

type DeleteRes struct {
	g.Meta `mime:"application/json"`
}
