package service

import (
	"context"
	"errors"
	"fmt"

	"easy-stream/internal/model"
	"easy-stream/internal/repository"
	"easy-stream/internal/storage"
	"easy-stream/pkg/logger"

	"go.uber.org/zap"
)

type RecordService struct {
	recordRepo    *repository.RecordRepository
	streamRepo    *repository.StreamRepository
	recordFileMgr storage.RecordFileManager
	storageMgr    *storage.Manager
}

func NewRecordService(recordRepo *repository.RecordRepository, streamRepo *repository.StreamRepository, recordFileMgr storage.RecordFileManager, storageMgr *storage.Manager) *RecordService {
	return &RecordService{
		recordRepo:    recordRepo,
		streamRepo:    streamRepo,
		recordFileMgr: recordFileMgr,
		storageMgr:    storageMgr,
	}
}

// GetAllRecords 获取所有录制文件列表（分页）
func (s *RecordService) GetAllRecords(page, pageSize int) ([]*model.Stream, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	streams, total, err := s.recordRepo.GetAllRecords(offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 为每个录制文件生成访问URL
	for _, stream := range streams {
		for i := range stream.RecordFiles {
			if s.recordFileMgr != nil {
				// 保留数据库中已有的URLs（对象存储等），添加或更新download字段
				if stream.RecordFiles[i].URLs == nil {
					stream.RecordFiles[i].URLs = make(map[string]string)
				}
				stream.RecordFiles[i].URLs["download"] = s.recordFileMgr.GetFileURL(stream.RecordFiles[i].FilePath)
			}
		}
	}

	return streams, total, nil
}

// GetRecordsByStreamKey 根据stream_key获取录制文件
func (s *RecordService) GetRecordsByStreamKey(key string) (*model.Stream, error) {
	stream, err := s.recordRepo.GetRecordsByStreamKey(key)
	if err != nil {
		if errors.Is(err, repository.ErrStreamNotFound) {
			return nil, ErrStreamNotFound
		}
		return nil, err
	}

	// 为每个录制文件生成访问URL
	for i := range stream.RecordFiles {
		if s.recordFileMgr != nil {
			// 保留数据库中已有的URLs（对象存储等），添加或更新download字段
			if stream.RecordFiles[i].URLs == nil {
				stream.RecordFiles[i].URLs = make(map[string]string)
			}
			stream.RecordFiles[i].URLs["download"] = s.recordFileMgr.GetFileURL(stream.RecordFiles[i].FilePath)
		}
	}

	return stream, nil
}

// DeleteRecordFile 删除指定的录制文件
func (s *RecordService) DeleteRecordFile(key, fileName string) error {
	// 先获取stream，确保存在
	stream, err := s.streamRepo.GetByKey(key)
	if err != nil {
		if errors.Is(err, repository.ErrStreamNotFound) {
			return ErrStreamNotFound
		}
		return err
	}

	// 检查文件是否存在
	var targetFile *model.RecordFile
	for i := range stream.RecordFiles {
		if stream.RecordFiles[i].FileName == fileName {
			targetFile = &stream.RecordFiles[i]
			break
		}
	}

	if targetFile == nil {
		return fmt.Errorf("record file not found")
	}

	ctx := context.Background()

	// 1. 删除ZLM原始文件（如果配置了文件管理器）
	if s.recordFileMgr != nil {
		if err := s.recordFileMgr.DeleteFile(targetFile.FilePath); err != nil {
			logger.Error("failed to delete original file",
				zap.String("file_path", targetFile.FilePath),
				zap.Error(err))
			// 继续执行，不因为原始文件删除失败而中断
		}
	}

	// 2. 删除对象存储中的备份文件
	if s.storageMgr != nil && s.storageMgr.HasStorages() {
		// 从URLs中获取需要删除的存储
		if targetFile.URLs != nil {
			for storageName := range targetFile.URLs {
				// 跳过download（这是原始文件，已经在上面删除了）
				if storageName == "download" {
					continue
				}

				// 删除对象存储中的文件
				storage := s.storageMgr.GetStorageByName(storageName)
				if storage != nil {
					if err := storage.Delete(ctx, targetFile.FileName); err != nil {
						logger.Error("failed to delete file from storage",
							zap.String("storage", storageName),
							zap.String("file", targetFile.FileName),
							zap.Error(err))
						// 继续执行，不因为某个存储删除失败而中断
					} else {
						logger.Info("file deleted from storage",
							zap.String("storage", storageName),
							zap.String("file", targetFile.FileName))
					}
				}
			}
		}
	}

	// 3. 删除数据库中的记录
	return s.recordRepo.DeleteRecordFile(key, fileName)
}

// IsLocalMode 是否是本地模式
func (s *RecordService) IsLocalMode() bool {
	if s.recordFileMgr == nil {
		return false
	}
	return s.recordFileMgr.IsLocal()
}

// GetFullPath 获取文件的完整路径（仅本地模式）
func (s *RecordService) GetFullPath(filePath string) string {
	if s.recordFileMgr == nil {
		return ""
	}
	return s.recordFileMgr.GetFullPath(filePath)
}
