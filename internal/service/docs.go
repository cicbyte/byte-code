package service

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/docs"
	"github.com/gogf/gf/v2/net/ghttp"
)

// IDocs 项目记忆与文档中枢
type IDocs interface {
	// 文档（磁盘 vault 真相源）
	Tree(ctx context.Context, projectId int64, space string) ([]api.VaultTreeNode, error)
	ReadFile(ctx context.Context, projectId int64, path string) (*api.VaultFileGetRes, error)
	ServeRaw(ctx context.Context, r *ghttp.Request, projectId int64, path string) error
	WriteFile(ctx context.Context, projectId int64, path, content string) (*api.VaultFileWriteRes, error)
	PatchFile(ctx context.Context, projectId int64, path string, ops []api.VaultPatchOperation) (*api.VaultFilePatchRes, error)
	CreateFolder(ctx context.Context, projectId int64, path string) error
	Upload(ctx context.Context, projectId int64, path string, file *ghttp.UploadFile) (*api.VaultUploadRes, error)
	Move(ctx context.Context, projectId int64, from, to string) error
	Delete(ctx context.Context, projectId int64, path string) error
	UpdateMeta(ctx context.Context, projectId int64, req *api.VaultMetaUpdateReq) error
	Search(ctx context.Context, projectId int64, q, space string) ([]api.VaultSearchItem, error)
	Linked(ctx context.Context, projectId int64, target string) ([]api.VaultSearchItem, error)
	Refresh(ctx context.Context, projectId int64) (*api.VaultRefreshRes, error)

	// 版本历史（.history 快照）
	HistoryList(ctx context.Context, projectId int64, path string) ([]api.HistoryItem, error)
	HistoryRead(ctx context.Context, projectId int64, path, snapshot string) (string, error)
	HistoryRestore(ctx context.Context, projectId int64, path, snapshot string) error

	// KV 记忆（状态机存储）
	MemList(ctx context.Context, projectId int64, prefix, include string) ([]api.MemoryItem, error)
	MemGet(ctx context.Context, projectId int64, key string) (*api.MemoryGetRes, error)
	MemSet(ctx context.Context, projectId int64, key string, req *api.MemorySetReq) error
	MemVerify(ctx context.Context, projectId int64, key string, userId int64) error
	MemExpire(ctx context.Context, projectId int64, key string) error
	MemDelete(ctx context.Context, projectId int64, key string) error
	// MemMaterialize 物化腐化状态：TTL 到期→expired、超阈值未验证→stale（定时 + 读取惰性调用）
	MemMaterialize(ctx context.Context) error
}

var localDocs IDocs

func Docs() IDocs {
	if localDocs == nil {
		panic("implement not found for interface IDocs")
	}
	return localDocs
}

func RegisterDocs(i IDocs) {
	localDocs = i
}
