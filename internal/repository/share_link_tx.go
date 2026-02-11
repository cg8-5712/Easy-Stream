package repository

import (
	"easy-stream/internal/model"

	"gorm.io/gorm"
)

// DeleteByStreamKeyTx 在事务中删除直播的所有分享链接
func (r *ShareLinkRepository) DeleteByStreamKeyTx(tx *gorm.DB, streamKey string) error {
	return tx.Where("stream_key = ?", streamKey).Delete(&model.ShareLink{}).Error
}
