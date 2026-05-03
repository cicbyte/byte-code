package product

import "github.com/gogf/gf/v2/frame/g"

// ==================== 产品 CRUD ====================

type ProductCreateReq struct {
	g.Meta      `path:"/v1/products" method:"post" tags:"产品管理" summary:"创建产品"`
	Name        string `json:"name" v:"required#产品名称不能为空"`
	Description string `json:"description"`
}

type ProductCreateRes struct {
	Id int `json:"id"`
}

type ProductUpdateReq struct {
	g.Meta      `path:"/v1/products/{id}" method:"put" tags:"产品管理" summary:"更新产品"`
	Id          int    `json:"id" v:"required" in:"path"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type ProductUpdateRes struct {
	g.Meta `mime:"application/json"`
}

type ProductDeleteReq struct {
	g.Meta `path:"/v1/products/{id}" method:"delete" tags:"产品管理" summary:"删除产品"`
	Id     int `json:"id" v:"required" in:"path"`
}

type ProductDeleteRes struct {
	g.Meta `mime:"application/json"`
}

type ProductListReq struct {
	g.Meta `path:"/v1/products" method:"get" tags:"产品管理" summary:"产品列表"`
	Status string `json:"status" in:"query"`
	Page   int    `json:"page" in:"query" d:"1"`
	Size   int    `json:"size" in:"query" d:"20"`
}

type ProductListRes struct {
	List  []ProductItem `json:"list"`
	Total int           `json:"total"`
	Page  int           `json:"page"`
	Size  int           `json:"size"`
}

type ProductItem struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerId     int    `json:"ownerId"`
	OwnerName   string `json:"ownerName"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type ProductDetailReq struct {
	g.Meta `path:"/v1/products/{id}" method:"get" tags:"产品管理" summary:"产品详情"`
	Id     int `json:"id" v:"required" in:"path"`
}

type ProductDetailRes struct {
	ProductItem
}

// ==================== 需求 CRUD ====================

type RequirementCreateReq struct {
	g.Meta             `path:"/v1/products/{productId}/requirements" method:"post" tags:"需求管理" summary:"创建需求"`
	ProductId          int    `json:"productId" v:"required" in:"path"`
	ParentId           int    `json:"parentId"`
	Type               string `json:"type" v:"required|in:epic,story,task#类型不能为空|类型不合法" d:"story"`
	Title              string `json:"title" v:"required#标题不能为空"`
	Description        string `json:"description"`
	Priority           int    `json:"priority" d:"3"`
	AssigneeId         int    `json:"assigneeId"`
	MilestoneId        int    `json:"milestoneId"`
	AcceptanceCriteria string `json:"acceptanceCriteria"`
}

type RequirementCreateRes struct {
	Id int `json:"id"`
}

type RequirementUpdateReq struct {
	g.Meta             `path:"/v1/requirements/{id}" method:"put" tags:"需求管理" summary:"更新需求"`
	Id                 int    `json:"id" v:"required" in:"path"`
	Title              string `json:"title"`
	Description        string `json:"description"`
	Status             string `json:"status"`
	Priority           int    `json:"priority"`
	AssigneeId         int    `json:"assigneeId"`
	MilestoneId        int    `json:"milestoneId"`
	AcceptanceCriteria string `json:"acceptanceCriteria"`
	SortOrder          int    `json:"sortOrder"`
}

type RequirementUpdateRes struct {
	g.Meta `mime:"application/json"`
}

type RequirementDeleteReq struct {
	g.Meta `path:"/v1/requirements/{id}" method:"delete" tags:"需求管理" summary:"删除需求"`
	Id     int `json:"id" v:"required" in:"path"`
}

type RequirementDeleteRes struct {
	g.Meta `mime:"application/json"`
}

type RequirementListReq struct {
	g.Meta    `path:"/v1/products/{productId}/requirements" method:"get" tags:"需求管理" summary:"需求列表"`
	ProductId int    `json:"productId" v:"required" in:"path"`
	Type      string `json:"type" in:"query"`
	Status    string `json:"status" in:"query"`
	ParentId  int    `json:"parentId" in:"query"`
	Page      int    `json:"page" in:"query" d:"1"`
	Size      int    `json:"size" in:"query" d:"50"`
}

type RequirementListRes struct {
	List  []RequirementItem `json:"list"`
	Total int               `json:"total"`
}

type RequirementItem struct {
	Id                 int              `json:"id"`
	ProductId          int              `json:"productId"`
	ParentId           int              `json:"parentId"`
	Type               string           `json:"type"`
	Title              string           `json:"title"`
	Description        string           `json:"description"`
	Status             string           `json:"status"`
	Priority           int              `json:"priority"`
	AssigneeId         int              `json:"assigneeId"`
	AssigneeName       string           `json:"assigneeName"`
	CreatorId          int              `json:"creatorId"`
	CreatorName        string           `json:"creatorName"`
	MilestoneId        int              `json:"milestoneId"`
	SortOrder          int              `json:"sortOrder"`
	AcceptanceCriteria string           `json:"acceptanceCriteria"`
	Source             string           `json:"source"`
	Children           []RequirementItem `json:"children,omitempty"`
	CreatedAt          string           `json:"createdAt"`
	UpdatedAt          string           `json:"updatedAt"`
}

type RequirementDetailReq struct {
	g.Meta `path:"/v1/requirements/{id}" method:"get" tags:"需求管理" summary:"需求详情"`
	Id     int `json:"id" v:"required" in:"path"`
}

type RequirementDetailRes struct {
	RequirementItem
}

// ==================== 里程碑 ====================

type MilestoneCreateReq struct {
	g.Meta      `path:"/v1/products/{productId}/milestones" method:"post" tags:"里程碑" summary:"创建里程碑"`
	ProductId   int    `json:"productId" v:"required" in:"path"`
	Name        string `json:"name" v:"required#名称不能为空"`
	Description string `json:"description"`
	TargetDate  string `json:"targetDate"`
}

type MilestoneCreateRes struct {
	Id int `json:"id"`
}

type MilestoneListReq struct {
	g.Meta    `path:"/v1/products/{productId}/milestones" method:"get" tags:"里程碑" summary:"里程碑列表"`
	ProductId int `json:"productId" v:"required" in:"path"`
}

type MilestoneItem struct {
	Id          int    `json:"id"`
	ProductId   int    `json:"productId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TargetDate  string `json:"targetDate"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
}

type MilestoneListRes struct {
	List []MilestoneItem `json:"list"`
}
