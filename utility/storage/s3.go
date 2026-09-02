package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Storage S3存储实现
type S3Storage struct {
	client *minio.Client
	bucket string
}

// NewS3Storage 创建S3存储实例
func NewS3Storage(config StorageConfig) (*S3Storage, error) {
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
		Region: config.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("创建S3客户端失败: %w", err)
	}
	return &S3Storage{
		client: client,
		bucket: config.Bucket,
	}, nil
}

// NewS3StorageFromCtx 从数据库配置创建S3存储实例
func NewS3StorageFromCtx(ctx context.Context) (*S3Storage, error) {
	config, err := GetStorageConfig(ctx)
	if err != nil {
		return nil, err
	}
	return NewS3Storage(*config)
}

// GetStorageConfig 从数据库获取存储配置
func GetStorageConfig(ctx context.Context) (*StorageConfig, error) {
	var configStr string
	err := g.DB().Model("sys_config").Ctx(ctx).
		Where("key", "storage_config").
		Fields("value").Scan(&configStr)
	if err != nil {
		return nil, fmt.Errorf("获取存储配置失败: %w", err)
	}
	if configStr == "" {
		// 未配置存储时不再回退到硬编码默认值（含内网地址与弱凭据），
		// 引导用户先完成存储配置
		return nil, fmt.Errorf("存储服务未配置，请先在系统设置中完成 S3 存储配置")
	}
	config := &StorageConfig{}
	err = gconv.Scan(configStr, config)
	if err != nil {
		return nil, fmt.Errorf("解析存储配置失败: %w", err)
	}
	return config, nil
}

// Upload 上传文件到S3
func (s *S3Storage) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	// 确保bucket存在
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("检查bucket失败: %w", err)
	}
	if !exists {
		if err = s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("创建bucket失败: %w", err)
		}
	}

	uploadInfo, err := s.client.PutObject(ctx, s.bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("上传文件失败: %w", err)
	}
	g.Log().Infof(ctx, "文件上传成功: bucket=%s, key=%s, size=%d", s.bucket, uploadInfo.Key, uploadInfo.Size)
	return nil
}

// DownloadURL 获取预签名下载URL
func (s *S3Storage) DownloadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucket, key, expiry, reqParams)
	if err != nil {
		return "", fmt.Errorf("生成预签名URL失败: %w", err)
	}
	return presignedURL.String(), nil
}

// Delete 从S3删除文件
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}
	g.Log().Infof(ctx, "文件删除成功: bucket=%s, key=%s", s.bucket, key)
	return nil
}

// Ping 测试连接
func (s *S3Storage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("S3连接测试失败: %w", err)
	}
	return nil
}
