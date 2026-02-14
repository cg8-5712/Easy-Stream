package repository

import (
	"fmt"
	"time"

	"easy-stream/internal/model"

	"gorm.io/gorm"
)

// DeleteShareCodeTx 在事务中删除分享码
func (r *StreamRepository) DeleteShareCodeTx(tx *gorm.DB, streamKey string) error {
	return tx.Model(&model.Stream{}).Where("stream_key = ?", streamKey).Updates(map[string]interface{}{
		"share_code":            nil,
		"share_code_max_uses":   0,
		"share_code_used_count": 0,
		"updated_at":            time.Now(),
	}).Error
}

// UpdateTx 在事务中更新推流信息
func (r *StreamRepository) UpdateTx(tx *gorm.DB, stream *model.Stream) error {
	return tx.Save(stream).Error
}

// EndStreamTx 在事务中执行结束直播的所有数据库操作
func (r *StreamRepository) EndStreamTx(stream *model.Stream, shareLinkRepo *ShareLinkRepository) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		streamKey := stream.StreamKey

		// 清理分享码
		if stream.ShareCode != nil {
			if err := r.DeleteShareCodeTx(tx, streamKey); err != nil {
				return fmt.Errorf("delete share code: %w", err)
			}
		}

		// 清理分享链接
		if err := shareLinkRepo.DeleteByStreamKeyTx(tx, streamKey); err != nil {
			return fmt.Errorf("delete share links: %w", err)
		}

		// 重置当前观看人数并更新状态
		now := time.Now()
		stream.ActualEndTime = &now
		stream.Status = model.StreamStatusEnded
		stream.CurrentViewers = 0
		stream.ShareCode = nil
		stream.ShareCodeMaxUses = 0
		stream.ShareCodeUsedCount = 0

		if err := r.UpdateTx(tx, stream); err != nil {
			return fmt.Errorf("update stream: %w", err)
		}

		return nil
	})
}
