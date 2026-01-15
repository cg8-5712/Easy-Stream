package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"easy-stream/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Storage S3兼容存储（支持AWS S3/腾讯COS/阿里OSS）
type S3Storage struct {
	name         string
	client       *s3.Client
	bucket       string
	pathPrefix   string
	customDomain string
	endpoint     string
}

// NewS3Storage 创建S3兼容存储
func NewS3Storage(cfg config.StorageTarget) (*S3Storage, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("bucket is required")
	}
	if cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		return nil, fmt.Errorf("accessKeyId and secretAccessKey are required")
	}

	// 创建凭证
	creds := credentials.NewStaticCredentialsProvider(
		cfg.AccessKeyID,
		cfg.SecretAccessKey,
		"",
	)

	// 加载配置
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithCredentialsProvider(creds),
		awsconfig.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config failed: %w", err)
	}

	// 创建S3客户端
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		// COS/OSS 需要使用路径风格
		o.UsePathStyle = true
	})

	return &S3Storage{
		name:         cfg.Name,
		client:       client,
		bucket:       cfg.Bucket,
		pathPrefix:   cfg.PathPrefix,
		customDomain: cfg.CustomDomain,
		endpoint:     cfg.Endpoint,
	}, nil
}

// Name 返回存储名称
func (s *S3Storage) Name() string {
	return s.name
}

// Upload 上传文件到S3
func (s *S3Storage) Upload(ctx context.Context, localPath, remotePath string) (string, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("open file failed: %w", err)
	}
	defer file.Close()

	// 构建对象键
	key := remotePath
	if s.pathPrefix != "" {
		key = filepath.Join(s.pathPrefix, remotePath)
	}

	// 上传文件
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return "", fmt.Errorf("upload to s3 failed: %w", err)
	}

	// 生成访问URL
	var url string
	if s.customDomain != "" {
		url = fmt.Sprintf("%s/%s", s.customDomain, key)
	} else if s.endpoint != "" {
		url = fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, key)
	} else {
		url = fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucket, key)
	}

	return url, nil
}
