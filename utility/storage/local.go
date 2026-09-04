package storage

// 本地磁盘存储后端：未配置 S3 时的默认实现，附件落 resource/attachments/。
// 与"零外部依赖"的产品定位对齐——S3 配置后自动切换（见 NewStorageFromCtx）

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalStorage 实现 storage.go 的 Storage 接口（S3 与本地磁盘共用）
type LocalStorage struct{}

const localRoot = "resource/attachments"

func NewLocalStorage() *LocalStorage {
	return &LocalStorage{}
}

// safeKey 规范 key：禁止穿越（附件 key 由服务端生成，此为纵深防御）
func safeKey(key string) (string, error) {
	clean := filepath.ToSlash(filepath.Clean("/" + key))
	clean = strings.TrimPrefix(clean, "/")
	if clean == "" || strings.Contains(clean, "..") {
		return "", fmt.Errorf("invalid storage key")
	}
	return clean, nil
}

// Upload 写入本地磁盘
func (l *LocalStorage) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	k, err := safeKey(key)
	if err != nil {
		return err
	}
	abs := filepath.Join(localRoot, filepath.FromSlash(k))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	f, err := os.Create(abs)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, reader); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}
	return nil
}

// DownloadURL 本地后端无预签名：返回 local:key 约定，由附件接口
// 翻译为鉴权下载端点（/attachments/{id}/file）后下发
func (l *LocalStorage) DownloadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	k, err := safeKey(key)
	if err != nil {
		return "", err
	}
	return "local:" + k, nil
}

// Delete 删除本地文件
func (l *LocalStorage) Delete(ctx context.Context, key string) error {
	k, err := safeKey(key)
	if err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(localRoot, filepath.FromSlash(k))); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除文件失败: %w", err)
	}
	return nil
}

// Ping 本地存储恒可用
func (l *LocalStorage) Ping() error {
	return os.MkdirAll(localRoot, 0o755)
}

// OpenPath 暴露本地文件的绝对路径（附件下载接口 ServeFile 用）
func (l *LocalStorage) OpenPath(key string) (string, error) {
	k, err := safeKey(key)
	if err != nil {
		return "", err
	}
	return filepath.Join(localRoot, filepath.FromSlash(k)), nil
}

// NewStorageFromCtx 统一存储工厂：配置了 S3 用 S3，否则回落本地磁盘。
// 附件功能不再强依赖外部 MinIO
func NewStorageFromCtx(ctx context.Context) (Storage, error) {
	config, err := GetStorageConfig(ctx)
	if err != nil || config == nil || config.Endpoint == "" {
		// 未配置 S3：本地磁盘（开箱即用）
		return NewLocalStorage(), nil
	}
	return NewS3Storage(*config)
}
