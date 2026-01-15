package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"easy-stream/internal/config"
)

// LocalStorage 本地存储
type LocalStorage struct {
	name    string
	baseDir string
}

// NewLocalStorage 创建本地存储
func NewLocalStorage(cfg config.StorageTarget) (*LocalStorage, error) {
	if cfg.LocalDir == "" {
		return nil, fmt.Errorf("localDir is required for local storage")
	}

	// 确保目录存在
	if err := os.MkdirAll(cfg.LocalDir, 0755); err != nil {
		return nil, fmt.Errorf("create local dir failed: %w", err)
	}

	return &LocalStorage{
		name:    cfg.Name,
		baseDir: cfg.LocalDir,
	}, nil
}

// Name 返回存储名称
func (s *LocalStorage) Name() string {
	return s.name
}

// Upload 上传文件到本地存储（实际是复制文件）
func (s *LocalStorage) Upload(ctx context.Context, localPath, remotePath string) (string, error) {
	destPath := filepath.Join(s.baseDir, remotePath)

	// 确保目标目录存在
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("create dest dir failed: %w", err)
	}

	// 复制文件
	if err := copyFile(localPath, destPath); err != nil {
		return "", err
	}

	return destPath, nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
