// Package vault 项目记忆与文档中枢 API 定义。
// 文档（vault）以磁盘为真相源；记忆（memories）为 KV 状态机存储。
package vault

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// ==================== 文档：目录与读取 ====================

type VaultTreeNode struct {
	Name     string          `json:"name" dc:"文件/目录名"`
	Path     string          `json:"path" dc:"vault 内相对路径"`
	IsDir    bool            `json:"isDir"`
	Size     int64           `json:"size"`
	ModTime  string          `json:"modTime"`
	Meta     *FileMeta       `json:"meta,omitempty" dc:"文件元数据（目录为空）"`
	Children []VaultTreeNode `json:"children,omitempty"`
}

type FileMeta struct {
	Title  string   `json:"title"`
	Space  string   `json:"space"`
	Type   string   `json:"type"`
	Status string   `json:"status"`
	Tags   []string `json:"tags"`
	Linked []string `json:"linked"`
}

type VaultTreeReq struct {
	g.Meta    `path:"/projects/{id}/docs/tree" method:"get" tags:"文档中枢" summary:"vault 目录树"`
	Id        int64 `json:"-" in:"path" v:"required#项目ID不能为空"`
	Space     string `json:"space" in:"query" dc:"空间过滤 knowledge/work（仅过滤顶层文件归属）"`
}

type VaultTreeRes struct {
	g.Meta `mime:"application/json"`
	Tree   []VaultTreeNode `json:"tree"`
}

type VaultFileGetReq struct {
	g.Meta `path:"/projects/{id}/docs/file" method:"get" tags:"文档中枢" summary:"读取文件（文本返回正文，二进制返回元数据）"`
	Id     int64  `json:"-" in:"path" v:"required#项目ID不能为空"`
	Path   string `json:"path" in:"query" v:"required#文件路径不能为空"`
}

type VaultFileGetRes struct {
	g.Meta `mime:"application/json"`
	Path     string    `json:"path"`
	Binary   bool      `json:"binary" dc:"二进制文件无正文，走 raw 下载"`
	Content  string    `json:"content" dc:"文本正文（含 frontmatter）"`
	FileMeta *FileMeta `json:"meta"`
	Size     int64     `json:"size"`
}

type VaultFileRawReq struct {
	g.Meta `path:"/projects/{id}/docs/raw" method:"get" tags:"文档中枢" summary:"原始文件下载（浏览器直出）"`
	Id     int64  `json:"-" in:"path" v:"required#项目ID不能为空"`
	Path   string `json:"path" in:"query" v:"required#文件路径不能为空"`
}

type VaultFileRawRes struct {
	g.Meta `mime:"application/octet-stream" summary:"原始文件流"`
}

// ==================== 文档：写入 ====================

type VaultFileWriteReq struct {
	g.Meta  `path:"/projects/{id}/docs/file" method:"put" tags:"文档中枢" summary:"写入文件（整文件覆盖，写前自动 .history 快照）"`
	Id      int64  `json:"-" in:"path" v:"required#项目ID不能为空"`
	Path    string `json:"path" v:"required#文件路径不能为空"`
	Content string `json:"content" dc:"完整文件内容（md 可含 frontmatter）"`
}

type VaultFileWriteRes struct {
	g.Meta  `mime:"application/json"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
}

type VaultPatchOperation struct {
	Type    string `json:"type" v:"required#操作类型不能为空" dc:"replace|append|prepend"`
	Search  string `json:"search" dc:"replace：要查找的精确文本"`
	Replace string `json:"replace" dc:"replace：替换文本"`
	Content string `json:"content" dc:"append/prepend：追加内容"`
}

type VaultFilePatchReq struct {
	g.Meta     `path:"/projects/{id}/docs/file/patch" method:"post" tags:"文档中枢" summary:"补丁式修改（search/replace + append/prepend）"`
	Id         int64               `json:"-" in:"path" v:"required#项目ID不能为空"`
	Path       string              `json:"path" v:"required#文件路径不能为空"`
	Operations []VaultPatchOperation `json:"operations" v:"required#操作列表不能为空"`
}

type VaultPatchItem struct {
	Index   int    `json:"index"`
	Type    string `json:"type"`
	Applied bool   `json:"applied"`
	Reason  string `json:"reason,omitempty" dc:"未应用原因（如 search 未命中）"`
}

type VaultFilePatchRes struct {
	g.Meta    `mime:"application/json"`
	Applied   int              `json:"applied" dc:"成功应用条数"`
	Skipped   int              `json:"skipped"`
	Items     []VaultPatchItem `json:"items"`
}

type VaultFolderCreateReq struct {
	g.Meta `path:"/projects/{id}/docs/folder" method:"post" tags:"文档中枢" summary:"创建目录（幂等）"`
	Id     int64  `json:"-" in:"path" v:"required#项目ID不能为空"`
	Path   string `json:"path" v:"required#目录路径不能为空"`
}

type VaultFolderCreateRes struct {
	g.Meta `mime:"application/json"`
}

type VaultUploadReq struct {
	g.Meta `path:"/projects/{id}/docs/upload" method:"post" mime:"multipart/form-data" tags:"文档中枢" summary:"上传文件（任意类型落盘）"`
	Id     int64                `json:"-" in:"path" v:"required#项目ID不能为空"`
	Path   string               `json:"path" dc:"目标目录（可含文件名，缺省用上传文件名）"`
	File   *ghttp.UploadFile    `json:"file" type:"file" v:"required#文件不能为空"`
}

type VaultUploadRes struct {
	g.Meta `mime:"application/json"`
	Path   string `json:"path"`
	Size   int64  `json:"size"`
}

type VaultFileMoveReq struct {
	g.Meta `path:"/projects/{id}/docs/file/move" method:"post" tags:"文档中枢" summary:"移动/重命名文件或目录"`
	Id     int64  `json:"-" in:"path" v:"required#项目ID不能为空"`
	From   string `json:"from" v:"required#源路径不能为空"`
	To     string `json:"to" v:"required#目标路径不能为空"`
}

type VaultFileMoveRes struct {
	g.Meta `mime:"application/json"`
}

type VaultFileDeleteReq struct {
	g.Meta `path:"/projects/{id}/docs/file" method:"delete" tags:"文档中枢" summary:"删除文件或目录（目录递归删除，先入 .history）"`
	Id     int64  `json:"-" in:"path" v:"required#项目ID不能为空"`
	Path   string `json:"path" in:"query" v:"required#文件路径不能为空"`
}

type VaultFileDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// ==================== 文档：元数据 ====================

type VaultMetaUpdateReq struct {
	g.Meta `path:"/projects/{id}/docs/meta" method:"put" tags:"文档中枢" summary:"更新 frontmatter 元数据（改写文件头，不动正文）"`
	Id     int64    `json:"-" in:"path" v:"required#项目ID不能为空"`
	Path   string   `json:"path" v:"required#文件路径不能为空"`
	Title  *string  `json:"title" dc:"不传=不更新"`
	Type   *string  `json:"type" dc:"不传=不更新"`
	Status *string  `json:"status" v:"in:,draft,published#状态必须是draft/published" dc:"不传=不更新"`
	Tags   []string `json:"tags" dc:"不传=不更新，空数组=清空"`
	Linked []string `json:"linked" dc:"不传=不更新，空数组=清空"`
}

type VaultMetaUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// ==================== 文档：查询 ====================

type VaultSearchReq struct {
	g.Meta `path:"/projects/{id}/docs/search" method:"get" tags:"文档中枢" summary:"搜索文档（标题/标签/路径 + md 正文）"`
	Id     int64  `json:"-" in:"path" v:"required#项目ID不能为空"`
	Q      string `json:"q" in:"query" v:"required#关键字不能为空"`
	Space  string `json:"space" in:"query" dc:"空间过滤"`
}

type VaultSearchItem struct {
	Path  string   `json:"path"`
	Title string   `json:"title"`
	Space string   `json:"space"`
	Type  string   `json:"type"`
	Tags  []string `json:"tags"`
}

type VaultSearchRes struct {
	g.Meta `mime:"application/json"`
	Items  []VaultSearchItem `json:"items"`
}

type VaultLinkedReq struct {
	g.Meta `path:"/projects/{id}/docs/linked" method:"get" tags:"文档中枢" summary:"反查关联文档（task:/req:/tc: 前缀）"`
	Id     int64  `json:"-" in:"path" v:"required#项目ID不能为空"`
	Task   int    `json:"task" in:"query" dc:"任务ID（与 req/tc 三选一）"`
	Req    int    `json:"req" in:"query" dc:"需求ID"`
	Tc     int    `json:"tc" in:"query" dc:"测试用例ID"`
}

type VaultLinkedRes struct {
	g.Meta `mime:"application/json"`
	Items  []VaultSearchItem `json:"items"`
}

type VaultRefreshReq struct {
	g.Meta `path:"/projects/{id}/docs/refresh" method:"post" tags:"文档中枢" summary:"手动触发增量重扫（agent 改完文件后调用）"`
	Id     int64 `json:"-" in:"path" v:"required#项目ID不能为空"`
}

type VaultRefreshRes struct {
	g.Meta  `mime:"application/json"`
	Changed int `json:"changed"`
	Deleted int `json:"deleted"`
}

// ==================== KV 记忆 ====================

type MemoryItem struct {
	Key            string `json:"key"`
	Value          string `json:"value"`
	Status         string `json:"status"`
	StaleDays      int    `json:"staleDays" dc:"距上次验证天数"`
	LastVerifiedAt string `json:"lastVerifiedAt"`
	ExpiresAt      string `json:"expiresAt,omitempty"`
	Source         string `json:"source" dc:"来源用户名"`
	UpdatedAt      string `json:"updatedAt"`
}

type MemoryListReq struct {
	g.Meta  `path:"/projects/{id}/memories" method:"get" tags:"项目记忆" summary:"记忆列表（默认 pending+active）"`
	Id      int64  `json:"-" in:"path" v:"required#项目ID不能为空"`
	Prefix  string `json:"prefix" in:"query" dc:"key 前缀过滤（点分层级，如 conventions.）"`
	Include string `json:"include" in:"query" dc:"额外包含的状态：stale,expired"`
}

type MemoryListRes struct {
	g.Meta `mime:"application/json"`
	List   []MemoryItem `json:"list"`
}

type MemoryGetReq struct {
	g.Meta `path:"/projects/{id}/memories/{key}" method:"get" tags:"项目记忆" summary:"读取单条记忆"`
	Id     int64  `json:"-" in:"path" v:"required#项目ID不能为空"`
	Key    string `json:"-" in:"path" v:"required#key不能为空"`
}

type MemoryGetRes struct {
	g.Meta `mime:"application/json"`
	MemoryItem
}

type MemorySetReq struct {
	g.Meta `path:"/projects/{id}/memories/{key}" method:"put" tags:"项目记忆" summary:"写入/更新记忆（upsert）"`
	Id     int64  `json:"-" in:"path" v:"required#项目ID不能为空"`
	Key    string `json:"-" in:"path" v:"required#key不能为空"`
	Value  string `json:"value" v:"max-length:65536#单值上限64KB"`
	Ttl    string `json:"ttl" dc:"有效期：30m/12h/7d，缺省永不过期"`
	Status string `json:"status" v:"in:,pending,active#状态必须是pending/active" dc:"缺省active"`
}

type MemorySetRes struct {
	g.Meta `mime:"application/json"`
}

type MemoryVerifyReq struct {
	g.Meta `path:"/projects/{id}/memories/{key}/verify" method:"post" tags:"项目记忆" summary:"验证保鲜（刷新 last_verified_at，pending→active）"`
	Id     int64  `json:"-" in:"path" v:"required#项目ID不能为空"`
	Key    string `json:"-" in:"path" v:"required#key不能为空"`
}

type MemoryVerifyRes struct {
	g.Meta `mime:"application/json"`
}

type MemoryExpireReq struct {
	g.Meta `path:"/projects/{id}/memories/{key}/expire" method:"post" tags:"项目记忆" summary:"主动废弃（status→expired）"`
	Id     int64  `json:"-" in:"path" v:"required#项目ID不能为空"`
	Key    string `json:"-" in:"path" v:"required#key不能为空"`
}

type MemoryExpireRes struct {
	g.Meta `mime:"application/json"`
}

type MemoryDeleteReq struct {
	g.Meta `path:"/projects/{id}/memories/{key}" method:"delete" tags:"项目记忆" summary:"删除记忆"`
	Id     int64  `json:"-" in:"path" v:"required#项目ID不能为空"`
	Key    string `json:"-" in:"path" v:"required#key不能为空"`
}

type MemoryDeleteRes struct {
	g.Meta `mime:"application/json"`
}
