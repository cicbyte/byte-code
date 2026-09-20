package project

import (
	commonApi "github.com/cicbyte/byte-code/api/v1/common"
	"github.com/gogf/gf/v2/frame/g"
)

// ==================== 项目发布（Releases） ====================
// 内部项目不走外部 CI：本地打包 → 推平台 → 团队内下载；
// 文件是独立体系（release_files）：按 项目/版本 组织、支持几百 MB 安装包、
// 公开令牌直链分享（/api/release-files/public/{token} 免鉴权下载）

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
	ProjectId int    `json:"-" in:"path" v:"required#项目ID不能为空"`
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
	FileCount      int    `json:"fileCount" dc:"文件数"`
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
	g.Meta `path:"/releases/{id}" method:"delete" tags:"项目发布" summary:"删除发布（级联删文件记录与存储对象）"`
	Id     int    `json:"-" in:"path" v:"required#发布ID不能为空"`
}

type ReleaseDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// ---------- 发布文件（独立体系） ----------

type ReleaseFileListReq struct {
	g.Meta `path:"/releases/{id}/files" method:"get" tags:"项目发布" summary:"发布文件列表"`
	Id     int  `json:"-" in:"path" v:"required#发布ID不能为空"`
}

type ReleaseFileListRes struct {
	g.Meta `mime:"application/json"`
	List   []ReleaseFileItem `json:"list"`
}

type ReleaseFileItem struct {
	Id             int    `json:"id"`
	ReleaseId      int    `json:"releaseId"`
	FileName       string `json:"fileName"`
	FileSize       int64  `json:"fileSize"`
	MimeType       string `json:"mimeType"`
	UploaderId     int    `json:"uploaderId"`
	UploaderName   string `json:"uploaderName,omitempty"`
	DownloadCount  int    `json:"downloadCount"`
	Shared         bool   `json:"shared" dc:"是否存在有效分享令牌"`
	ShareExpiresAt string `json:"shareExpiresAt,omitempty"`
	CreatedAt      string `json:"createdAt"`
}

type ReleaseFileUploadReq struct {
	g.Meta `path:"/releases/{id}/files" method:"post" mime:"multipart/form-data" tags:"项目发布" summary:"上传发布文件（单文件上限2GB）"`
	Id     int    `json:"-" in:"path" v:"required#发布ID不能为空"`
	File   string `p:"file" type:"file" v:"required#文件不能为空"`
}

type ReleaseFileUploadRes struct {
	g.Meta `mime:"application/json"`
	Id     int `json:"id"`
}

type ReleaseFileDeleteReq struct {
	g.Meta `path:"/release-files/{id}" method:"delete" tags:"项目发布" summary:"删除发布文件（含存储对象）"`
	Id     int    `json:"-" in:"path" v:"required#文件ID不能为空"`
}

type ReleaseFileDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// ReleaseFileShareReq 分享直链：幂等——已有未过期令牌直接返回；
// days>0 时重新生成并设定有效期（天），0=永久
type ReleaseFileShareReq struct {
	g.Meta `path:"/release-files/{id}/share" method:"post" tags:"项目发布" summary:"生成/获取分享直链"`
	Id     int `json:"-" in:"path" v:"required#文件ID不能为空"`
	Days   int `json:"days" d:"0" v:"min:0#天数不能为负" dc:"有效期（天），0=永久；非0重新生成令牌"`
}

type ReleaseFileShareRes struct {
	g.Meta `mime:"application/json"`
	// 免鉴权公开下载路径（拼站点地址即直链）
	Url       string `json:"url"`
	ExpiresAt string `json:"expiresAt" dc:"过期时间，空=永久"`
}

type ReleaseFileShareRevokeReq struct {
	g.Meta `path:"/release-files/{id}/share" method:"delete" tags:"项目发布" summary:"吊销分享直链"`
	Id     int `json:"-" in:"path" v:"required#文件ID不能为空"`
}

type ReleaseFileShareRevokeRes struct {
	g.Meta `mime:"application/json"`
}

// ReleaseFilePublicReq 公开下载（免鉴权）：凭分享令牌取文件流或 S3 预签名跳转
type ReleaseFilePublicReq struct {
	g.Meta `path:"/release-files/public/{token}" method:"get" tags:"项目发布" summary:"分享直链下载（免鉴权）"`
	Token  string `json:"-" in:"path" v:"required#令牌不能为空"`
}

type ReleaseFilePublicRes struct {
	g.Meta `mime:"application/octet-stream"`
}
