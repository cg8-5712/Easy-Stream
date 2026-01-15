package storage

import (
	"context"
	"fmt"

	"easy-stream/internal/config"
)

// Storage 存储接口
type Storage interface {
	Upload(ctx context.Context, localPath, remotePath string) (url string, err error)
	Name() string
}

// Manager 存储管理器
type Manager struct {
	storages []Storage
}

// NewManager 创建存储管理器
func NewManager(cfg config.StorageConfig) (*Manager, error) {
	m := &Manager{storages: make([]Storage, 0)}

	for _, target := range cfg.Targets {
		if !target.Enabled {
			continue
		}

		var s Storage
		var err error

		switch target.Type {
		case "local":
			s, err = NewLocalStorage(target)
		case "s3", "cos", "oss":
			s, err = NewS3Storage(target)
		default:
			return nil, fmt.Errorf("unknown storage type: %s", target.Type)
		}

		if err != nil {
			return nil, fmt.Errorf("init storage %s failed: %w", target.Name, err)
		}
		m.storages = append(m.storages, s)
	}

	return m, nil
}

// UploadToAll 上传到所有启用的存储
func (m *Manager) UploadToAll(ctx context.Context, localPath, remotePath string) map[string]string {
	results := make(map[string]string)
	for _, s := range m.storages {
		url, err := s.Upload(ctx, localPath, remotePath)
		if err != nil {
			results[s.Name()] = fmt.Sprintf("error: %v", err)
		} else {
			results[s.Name()] = url
		}
	}
	return results
}

// HasStorages 是否有启用的存储
func (m *Manager) HasStorages() bool {
	return len(m.storages) > 0
}
