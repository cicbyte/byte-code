package storage

import (
	"context"
	"io"
	"time"
)

// Storage 存储接口
type Storage interface {
	// Upload 上传文件
	Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	// DownloadURL 获取预签名下载URL
	DownloadURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	// Delete 删除文件
	Delete(ctx context.Context, key string) error
}

// StorageConfig 存储配置
type StorageConfig struct {
	Endpoint  string `json:"endpoint"`
	Bucket    string `json:"bucket"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	Region    string `json:"region"`
	UseSSL    bool   `json:"useSSL"`
}

// TestConfig 测试存储连接
func TestConfig(config StorageConfig) error {
	s, err := NewS3Storage(config)
	if err != nil {
		return err
	}
	return s.Ping()
}
