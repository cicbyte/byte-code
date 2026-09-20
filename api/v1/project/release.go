package project

import (
	commonApi "github.com/cicbyte/byte-code/api/v1/common"
	"github.com/gogf/gf/v2/frame/g"
)

// ==================== 项目发布（Releases，#551） ====================
// 内部项目不走外部 CI：本地打包 → 推平台 → 团队内下载；
// 文件复用附件通道（entityType=release）

type ReleaseCreateReq struct {
	g.Meta    `path:"/projects/{projectId}/releases" method:"post" tags:"项目发布" summary:"创建发布"`
	ProjectId int    `json:"-" in:"path" v:"required#项目ID不能为空"`
	Version   string `json:"version" v:"required|max-length:64#版本号不能为空|版本号上限64字符"`
	Title     string `json:"title" v:"max-length:191#标题上限191字符" dc:"缺省同版本号"`
	Notes     string `json:"notes" v:"max-length:65535#发布说明上限64K" dc:"changelog / 安装说明（markdown）"`
	Channel   string `json:"channel" d:"stable" v:"in:stable,beta,nightly#渠道必须是stable/beta/nightly"`
}

type ReleaseCreateRes struct {
	g.Meta `mime:"application/json"`
	Id     int `json:"id"`
}

type ReleaseListReq struct {
	g.Meta    `path:"/projects/{projectId}/releases" method:"get" tags:"项目发布" summary:"发布列表"`
	ProjectId int `json:"-" in:"path" v:"required#项目ID不能为空"`
	Channel   string `json:"channel" in:"query" v:"in:,stable,beta,nightly#渠道不合法" dc:"渠道过滤"`
	commonApi.PageReq
}

type ReleaseListRes struct {
	g.Meta `mime:"application/json"`
	commonApi.ListRes
	List []ReleaseItem `json:"list"`
}

type ReleaseItem struct {
	Id             int    `json:"id"`
	ProjectId      int    `json:"projectId"`
	Version        string `json:"version"`
	Title          string `json:"title"`
	Notes          string `json:"notes"`
	Channel        string `json:"channel"`
	CreatedBy      int    `json:"createdBy"`
	CreatedByName  string `json:"createdByName,omitempty"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
	FileCount      int    `json:"fileCount" dc:"附件文件数"`
	TotalSizeBytes int64  `json:"totalSizeBytes" dc:"文件总字节数"`
}

type ReleaseUpdateReq struct {
	g.Meta `path:"/releases/{id}" method:"put" tags:"项目发布" summary:"更新发布（标题/说明/渠道）"`
	Id     int    `json:"-" in:"path" v:"required#发布ID不能为空"`
	Title  string `json:"title" v:"max-length:191#标题上限191字符"`
	Notes  string `json:"notes" v:"max-length:65535#发布说明上限64K"`
	// Channel 空=不变更
	Channel string `json:"channel" v:"in:,stable,beta,nightly#渠道必须是stable/beta/nightly"`
}

type ReleaseUpdateRes struct {
	g.Meta `mime:"application/json"`
}

type ReleaseDeleteReq struct {
	g.Meta `path:"/releases/{id}" method:"delete" tags:"项目发布" summary:"删除发布（级联删其附件记录）"`
	Id     int `json:"-" in:"path" v:"required#发布ID不能为空"`
}

type ReleaseDeleteRes struct {
	g.Meta `mime:"application/json"`
}
