package aiuser

import "github.com/gogf/gf/v2/frame/g"

type AiUserCreateReq struct {
	g.Meta       `path:"/ai-users" method:"post" tags:"AI用户" summary:"创建AI用户"`
	Username     string `json:"username" v:"required#用户名不能为空"`
	RealName     string `json:"realName"`
	Capabilities string `json:"capabilities"`
}

type AiUserCreateRes struct {
	Id     int    `json:"id"`
	ApiKey string `json:"apiKey"`
}

type AiUserUpdateReq struct {
	g.Meta `path:"/ai-users/{id}" method:"put" tags:"AI用户" summary:"更新AI用户"`
	Id     int `json:"id" v:"required" in:"path"`
	// 指针语义：nil=不更新，非nil=更新（含零值/空串）。
	// Status 若为非指针 int，编辑弹窗只传 realName 时零值 0 会被
	// 误判为"禁用"并清空该 agent 全部访问（准入/会话/token）
	RealName     *string `json:"realName"`
	Capabilities *string `json:"capabilities"`
	Status       *int    `json:"status"`
}

type AiUserUpdateRes struct {
	g.Meta `mime:"application/json"`
}

type AiUserDeleteReq struct {
	g.Meta `path:"/ai-users/{id}" method:"delete" tags:"AI用户" summary:"删除AI用户"`
	Id     int `json:"id" v:"required" in:"path"`
}

type AiUserDeleteRes struct {
	g.Meta `mime:"application/json"`
}

type AiUserListReq struct {
	g.Meta `path:"/ai-users" method:"get" tags:"AI用户" summary:"AI用户列表"`
	Page   int `json:"page" in:"query" d:"1"`
	Size   int `json:"size" in:"query" d:"20"`
}

type AiUserListRes struct {
	List  []AiUserItem `json:"list"`
	Total int          `json:"total"`
}

type AiUserItem struct {
	Id           int    `json:"id"`
	Username     string `json:"username"`
	RealName     string `json:"realName"`
	Avatar       string `json:"avatar"`
	Capabilities string `json:"capabilities"`
	Status       int    `json:"status"`
	OwnerHumanId int      `json:"ownerHumanId"`
	CreatedAt    string   `json:"createdAt"`
	Projects     []string `json:"projects" dc:"已接入项目名列表（回填）"`
}

type AiUserResetKeyReq struct {
	g.Meta `path:"/ai-users/{id}/reset-key" method:"post" tags:"AI用户" summary:"重置API Key"`
	Id     int `json:"id" v:"required" in:"path"`
}

type AiUserResetKeyRes struct {
	ApiKey string `json:"apiKey"`
}

type AiLoginReq struct {
	g.Meta `path:"/auth/ai/login" method:"post" tags:"AI认证" summary:"AI用API Key登录"`
	ApiKey string `json:"apiKey" v:"required#API Key不能为空"`
}

type AiLoginRes struct {
	Token string `json:"token"`
}
