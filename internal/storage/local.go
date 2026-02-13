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

// Delete 删除本地存储中的文件
func (s *LocalStorage) Delete(ctx context.Context, remotePath string) error {
	destPath := filepath.Join(s.baseDir, remotePath)
	if err := os.Remove(destPath); err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在视为删除成功
		}
		return fmt.Errorf("delete local file failed: %w", err)
	}
	return nil
}

// UploadFromReader 从 Reader 上传文件到本地存储（流式传输）
func (s *LocalStorage) UploadFromReader(ctx context.Context, reader io.Reader, remotePath string, size int64) (string, error) {
	destPath := filepath.Join(s.baseDir, remotePath)

	// 确保目标目录存在
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("create dest dir failed: %w", err)
	}

	// 创建目标文件
	dstFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("create dest file failed: %w", err)
	}
	defer dstFile.Close()

	// 从 reader 复制到文件
	_, err = io.Copy(dstFile, reader)
	if err != nil {
		return "", fmt.Errorf("copy data failed: %w", err)
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
