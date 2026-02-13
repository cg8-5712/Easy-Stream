package repository

import (
	"time"

	"easy-stream/internal/model"

	"gorm.io/gorm"
)

type RecordRepository struct {
	db *gorm.DB
}

func NewRecordRepository(db *gorm.DB) *RecordRepository {
	return &RecordRepository{db: db}
}

// GetAllRecords 获取所有录制文件列表（分页）
func (r *RecordRepository) GetAllRecords(offset, limit int) ([]*model.Stream, int64, error) {
	query := r.db.Model(&model.Stream{}).Where("jsonb_array_length(record_files) > 0")

	// 查询总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询列表
	var streams []*model.Stream
	err := query.Order("updated_at DESC").Offset(offset).Limit(limit).Find(&streams).Error
	if err != nil {
		return nil, 0, err
	}

	return streams, total, nil
}

// GetRecordsByStreamKey 根据stream_key获取录制文件
func (r *RecordRepository) GetRecordsByStreamKey(key string) (*model.Stream, error) {
	var stream model.Stream
	return &stream, HandleNotFoundError(
		r.db.Where("stream_key = ?", key).First(&stream).Error,
		ErrStreamNotFound,
	)
}

// DeleteRecordFile 删除指定的录制文件
func (r *RecordRepository) DeleteRecordFile(key, fileName string) error {
	// 使用 PostgreSQL JSONB 操作删除指定文件
	return r.db.Exec(`
		UPDATE streams
		SET record_files = (
			SELECT jsonb_agg(elem)
			FROM jsonb_array_elements(record_files) elem
			WHERE elem->>'file_name' != ?
		),
		updated_at = ?
		WHERE stream_key = ?
	`, fileName, time.Now(), key).Error
}
