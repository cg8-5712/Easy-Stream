package repository

import (
	"encoding/json"
	"time"

	"easy-stream/internal/model"

	"gorm.io/gorm"
)

type StreamRepository struct {
	db *gorm.DB
}

func NewStreamRepository(db *gorm.DB) *StreamRepository {
	return &StreamRepository{db: db}
}

// Create 创建推流
func (r *StreamRepository) Create(stream *model.Stream) error {
	return r.db.Create(stream).Error
}

// GetByKey 根据 stream_key 获取
func (r *StreamRepository) GetByKey(key string) (*model.Stream, error) {
	var stream model.Stream
	return &stream, HandleNotFoundError(
		r.db.Where("stream_key = ?", key).First(&stream).Error,
		ErrStreamNotFound,
	)
}

// GetByID 根据 ID 获取
func (r *StreamRepository) GetByID(id int64) (*model.Stream, error) {
	var stream model.Stream
	return &stream, HandleNotFoundError(
		r.db.First(&stream, id).Error,
		ErrStreamNotFound,
	)
}

// List 获取推流列表
func (r *StreamRepository) List(req *model.StreamListRequest, offset, limit int) ([]*model.Stream, int64, error) {
	query := r.db.Model(&model.Stream{})

	// 构建查询条件
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if req.Visibility != "" {
		query = query.Where("visibility = ?", req.Visibility)
	}

	// 时间范围过滤
	if req.TimeRange != "" {
		switch req.TimeRange {
		case model.TimeRangePast:
			// 已结束的直播
			query = query.Where("actual_end_time IS NOT NULL")
		case model.TimeRangeCurrent:
			// 正在进行的直播
			query = query.Where("status = ?", model.StreamStatusPushing)
		case model.TimeRangeFuture:
			// 未开始的直播
			query = query.Where("status = ? AND scheduled_start_time > ?", model.StreamStatusIdle, time.Now())
		}
	}

	// 查询总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询列表
	var streams []*model.Stream
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&streams).Error
	if err != nil {
		return nil, 0, err
	}

	return streams, total, nil
}

// Update 更新推流信息
func (r *StreamRepository) Update(stream *model.Stream) error {
	return r.db.Save(stream).Error
}

// UpdateStatus 更新状态
func (r *StreamRepository) UpdateStatus(key, status string) error {
	return UpdateStreamByKeyWithTimestamp(r.db, key, map[string]interface{}{
		"status": status,
	})
}

// AppendRecordFile 追加录制文件（包含完整元数据）
// 使用 PostgreSQL JSONB 拼接操作
func (r *StreamRepository) AppendRecordFile(key string, recordFile *model.RecordFile) error {
	fileJSON, err := json.Marshal([]model.RecordFile{*recordFile})
	if err != nil {
		return err
	}
	return r.db.Exec(
		"UPDATE streams SET record_files = record_files || ?::jsonb, updated_at = ? WHERE stream_key = ?",
		fileJSON, time.Now(), key,
	).Error
}

// UpdateRecordEnabled 更新录制开关
func (r *StreamRepository) UpdateRecordEnabled(key string, enabled bool) error {
	return UpdateStreamByKeyWithTimestamp(r.db, key, map[string]interface{}{
		"record_enabled": enabled,
	})
}

// UpdateRecordStatus 更新录制状态
func (r *StreamRepository) UpdateRecordStatus(key, status string) error {
	return UpdateStreamByKeyWithTimestamp(r.db, key, map[string]interface{}{
		"record_status": status,
	})
}

// GetPushingStreams 获取所有正在推流的直播
func (r *StreamRepository) GetPushingStreams() ([]*model.Stream, error) {
	var streams []*model.Stream
	err := r.db.Where("status = ?", model.StreamStatusPushing).Find(&streams).Error
	return streams, err
}

// GetIdleStreams 获取所有空闲状态的直播（用于检查自动结束）
func (r *StreamRepository) GetIdleStreams() ([]*model.Stream, error) {
	var streams []*model.Stream
	err := r.db.Where("status = ?", model.StreamStatusIdle).Find(&streams).Error
	return streams, err
}

// IncrementViewers 增加观看人数（有人进入观看）
// 使用原生 SQL 保证原子性
// current_viewers: 当前在线人数 +1
// total_viewers: 累计观看人次 +1
// peak_viewers: 如果当前人数超过历史峰值则更新
func (r *StreamRepository) IncrementViewers(key string) error {
	return r.db.Exec(`
		UPDATE streams SET
			current_viewers = current_viewers + 1,
			total_viewers = total_viewers + 1,
			peak_viewers = GREATEST(peak_viewers, current_viewers + 1),
			updated_at = ?
		WHERE stream_key = ?
	`, time.Now(), key).Error
}

// DecrementViewers 减少观看人数（有人离开）
// 使用原生 SQL 保证原子性
func (r *StreamRepository) DecrementViewers(key string) error {
	return r.db.Exec(`
		UPDATE streams SET
			current_viewers = GREATEST(0, current_viewers - 1),
			updated_at = ?
		WHERE stream_key = ?
	`, time.Now(), key).Error
}

// ResetCurrentViewers 重置当前观看人数（直播结束时调用）
func (r *StreamRepository) ResetCurrentViewers(key string) error {
	return UpdateStreamByKeyWithTimestamp(r.db, key, map[string]interface{}{
		"current_viewers": 0,
	})
}

// Delete 删除推流
func (r *StreamRepository) Delete(key string) error {
	return r.db.Where("stream_key = ?", key).Delete(&model.Stream{}).Error
}

// GetByShareCode 根据分享码获取直播
func (r *StreamRepository) GetByShareCode(shareCode string) (*model.Stream, error) {
	var stream model.Stream
	return &stream, HandleNotFoundError(
		r.db.Where("share_code = ?", shareCode).First(&stream).Error,
		ErrStreamNotFound,
	)
}

// IncrementShareCodeUsedCount 增加分享码使用次数
func (r *StreamRepository) IncrementShareCodeUsedCount(streamKey string) error {
	return UpdateStreamByKeyWithTimestamp(r.db, streamKey, map[string]interface{}{
		"share_code_used_count": gorm.Expr("share_code_used_count + 1"),
	})
}

// UpdateShareCode 更新分享码（重新生成或添加分享码）
func (r *StreamRepository) UpdateShareCode(streamKey, shareCode string, maxUses int) error {
	return UpdateStreamByKeyWithTimestamp(r.db, streamKey, map[string]interface{}{
		"share_code":            shareCode,
		"share_code_max_uses":   maxUses,
		"share_code_used_count": 0,
	})
}

// UpdateShareCodeMaxUses 更新分享码最大使用次数（管理员调整）
func (r *StreamRepository) UpdateShareCodeMaxUses(streamKey string, maxUses int) error {
	return UpdateStreamByKeyWithTimestamp(r.db, streamKey, map[string]interface{}{
		"share_code_max_uses": maxUses,
	})
}

// DeleteShareCode 删除分享码
func (r *StreamRepository) DeleteShareCode(streamKey string) error {
	return UpdateStreamByKeyWithTimestamp(r.db, streamKey, map[string]interface{}{
		"share_code":            nil,
		"share_code_max_uses":   0,
		"share_code_used_count": 0,
	})
}
