package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"easy-stream/internal/config"
)

// Storage 存储接口
type Storage interface {
	Upload(ctx context.Context, localPath, remotePath string) (url string, err error)
	UploadFromReader(ctx context.Context, reader io.Reader, remotePath string, size int64) (url string, err error)
	Delete(ctx context.Context, remotePath string) error
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

// UploadFromReaderToAll 从 Reader 上传到所有启用的存储（流式传输）
func (m *Manager) UploadFromReaderToAll(ctx context.Context, reader io.Reader, remotePath string, size int64) map[string]string {
	results := make(map[string]string)

	// 对于多个存储，需要使用 TeeReader 或多次读取
	// 这里简化处理：如果只有一个存储，直接使用 reader
	// 如果有多个存储，需要先缓存到内存或使用其他方案
	if len(m.storages) == 1 {
		url, err := m.storages[0].UploadFromReader(ctx, reader, remotePath, size)
		if err != nil {
			results[m.storages[0].Name()] = fmt.Sprintf("error: %v", err)
		} else {
			results[m.storages[0].Name()] = url
		}
	} else {
		// 多个存储时，需要先读取到内存
		// TODO: 对于大文件，可以考虑使用临时文件
		data, err := io.ReadAll(reader)
		if err != nil {
			for _, s := range m.storages {
				results[s.Name()] = fmt.Sprintf("error: failed to read data: %v", err)
			}
			return results
		}

		for _, s := range m.storages {
			// 为每个存储创建新的 reader
			r := bytes.NewReader(data)
			url, err := s.UploadFromReader(ctx, r, remotePath, int64(len(data)))
			if err != nil {
				results[s.Name()] = fmt.Sprintf("error: %v", err)
			} else {
				results[s.Name()] = url
			}
		}
	}

	return results
}

// HasStorages 是否有启用的存储
func (m *Manager) HasStorages() bool {
	return len(m.storages) > 0
}

// DeleteFromAll 从所有启用的存储中删除文件
func (m *Manager) DeleteFromAll(ctx context.Context, remotePath string) map[string]error {
	results := make(map[string]error)
	for _, s := range m.storages {
		err := s.Delete(ctx, remotePath)
		if err != nil {
			results[s.Name()] = err
		}
	}
	return results
}

// GetStorageByName 根据名称获取存储
func (m *Manager) GetStorageByName(name string) Storage {
	for _, s := range m.storages {
		if s.Name() == name {
			return s
		}
	}
	return nil
}
