package storage

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"easy-stream/internal/config"
	"easy-stream/pkg/logger"

	"go.uber.org/zap"
)

// RecordFileManager 录制文件管理接口
type RecordFileManager interface {
	// GetFileURL 获取录制文件的访问URL
	GetFileURL(filePath string) string
	// DeleteFile 删除录制文件
	DeleteFile(filePath string) error
	// FileExists 检查文件是否存在
	FileExists(filePath string) (bool, error)
	// GetFileSize 获取文件大小
	GetFileSize(filePath string) (int64, error)
	// GetFullPath 获取文件的完整路径（用于本地文件访问）
	GetFullPath(filePath string) string
	// IsLocal 是否是本地模式
	IsLocal() bool
}

// LocalRecordFileManager 本地文件系统管理器（main server和ZLM在同一服务器）
type LocalRecordFileManager struct {
	basePath string // 录制文件的本地基础路径
	baseURL  string // 用于生成访问URL的基础URL
}

// NewLocalRecordFileManager 创建本地文件管理器
func NewLocalRecordFileManager(cfg config.ZLMediaKitConfig) *LocalRecordFileManager {
	return &LocalRecordFileManager{
		basePath: cfg.RecordLocalPath,
		baseURL:  cfg.RecordBaseURL,
	}
}

// GetFileURL 获取文件访问URL
func (m *LocalRecordFileManager) GetFileURL(filePath string) string {
	if m.baseURL == "" {
		return filePath
	}
	// 将文件路径转换为URL路径
	return fmt.Sprintf("%s/%s", m.baseURL, filepath.ToSlash(filePath))
}

// DeleteFile 删除本地文件
func (m *LocalRecordFileManager) DeleteFile(filePath string) error {
	fullPath := filepath.Join(m.basePath, filePath)
	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			logger.Warn("file not found when deleting", zap.String("path", fullPath))
			return nil // 文件不存在视为删除成功
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}
	logger.Info("record file deleted", zap.String("path", fullPath))
	return nil
}

// FileExists 检查文件是否存在
func (m *LocalRecordFileManager) FileExists(filePath string) (bool, error) {
	fullPath := filepath.Join(m.basePath, filePath)
	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetFileSize 获取文件大小
func (m *LocalRecordFileManager) GetFileSize(filePath string) (int64, error) {
	fullPath := filepath.Join(m.basePath, filePath)
	info, err := os.Stat(fullPath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// GetFullPath 获取文件的完整路径
func (m *LocalRecordFileManager) GetFullPath(filePath string) string {
	return filepath.Join(m.basePath, filePath)
}

// IsLocal 是否是本地模式
func (m *LocalRecordFileManager) IsLocal() bool {
	return true
}

// RemoteRecordFileManager 远程文件管理器（main server和ZLM在不同服务器）
type RemoteRecordFileManager struct {
	baseURL    string       // ZLM服务器的录制文件访问基础URL
	httpClient *http.Client // HTTP客户端
}

// NewRemoteRecordFileManager 创建远程文件管理器
func NewRemoteRecordFileManager(cfg config.ZLMediaKitConfig) *RemoteRecordFileManager {
	return &RemoteRecordFileManager{
		baseURL:    cfg.RecordBaseURL,
		httpClient: &http.Client{},
	}
}

// GetFileURL 获取文件访问URL
func (m *RemoteRecordFileManager) GetFileURL(filePath string) string {
	return fmt.Sprintf("%s/%s", m.baseURL, filepath.ToSlash(filePath))
}

// DeleteFile 远程模式下只删除数据库记录，不删除物理文件
// 物理文件需要在ZLM服务器上手动清理或通过定时任务清理
func (m *RemoteRecordFileManager) DeleteFile(filePath string) error {
	// 远程模式下，我们无法直接删除ZLM服务器上的文件
	// 因为ZLM没有提供删除录制文件的API
	// 这里只是记录日志，实际删除由数据库操作完成
	logger.Warn("remote mode: physical file not deleted, only database record removed",
		zap.String("file_path", filePath),
		zap.String("note", "please clean up files on ZLM server manually or via scheduled task"))
	return nil
}

// FileExists 检查远程文件是否存在
func (m *RemoteRecordFileManager) FileExists(filePath string) (bool, error) {
	url := fmt.Sprintf("%s/%s", m.baseURL, filepath.ToSlash(filePath))

	resp, err := m.httpClient.Head(url)
	if err != nil {
		return false, fmt.Errorf("failed to check remote file: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return true, nil
}

// GetFileSize 获取远程文件大小
func (m *RemoteRecordFileManager) GetFileSize(filePath string) (int64, error) {
	url := fmt.Sprintf("%s/%s", m.baseURL, filepath.ToSlash(filePath))

	resp, err := m.httpClient.Head(url)
	if err != nil {
		return 0, fmt.Errorf("failed to get remote file info: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp.ContentLength, nil
}

// GetFullPath 远程模式返回空字符串（不支持本地路径）
func (m *RemoteRecordFileManager) GetFullPath(filePath string) string {
	return ""
}

// IsLocal 是否是本地模式
func (m *RemoteRecordFileManager) IsLocal() bool {
	return false
}

// NewRecordFileManager 根据配置创建合适的文件管理器
func NewRecordFileManager(cfg config.ZLMediaKitConfig) (RecordFileManager, error) {
	switch cfg.RecordMode {
	case "local":
		if cfg.RecordLocalPath == "" {
			return nil, fmt.Errorf("recordLocalPath is required for local mode")
		}
		return NewLocalRecordFileManager(cfg), nil
	case "remote":
		if cfg.RecordBaseURL == "" {
			return nil, fmt.Errorf("recordBaseURL is required for remote mode")
		}
		return NewRemoteRecordFileManager(cfg), nil
	default:
		return nil, fmt.Errorf("unknown record mode: %s (must be 'local' or 'remote')", cfg.RecordMode)
	}
}
