package repository

import (
	"easy-stream/internal/model"

	"gorm.io/gorm"
)

type ShareLinkRepository struct {
	db *gorm.DB
}

func NewShareLinkRepository(db *gorm.DB) *ShareLinkRepository {
	return &ShareLinkRepository{db: db}
}

// Create 创建分享链接
func (r *ShareLinkRepository) Create(link *model.ShareLink) error {
	return r.db.Create(link).Error
}

// GetByToken 根据 token 获取分享链接
func (r *ShareLinkRepository) GetByToken(token string) (*model.ShareLink, error) {
	var link model.ShareLink
	err := r.db.Where("token = ?", token).First(&link).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// GetByID 根据 ID 获取分享链接
func (r *ShareLinkRepository) GetByID(id int64) (*model.ShareLink, error) {
	var link model.ShareLink
	err := r.db.First(&link, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// ListByStreamKey 获取直播的所有分享链接
func (r *ShareLinkRepository) ListByStreamKey(streamKey string) ([]*model.ShareLink, error) {
	var links []*model.ShareLink
	err := r.db.Where("stream_key = ?", streamKey).Order("created_at DESC").Find(&links).Error
	return links, err
}

// IncrementUsedCount 增加使用次数
func (r *ShareLinkRepository) IncrementUsedCount(token string) error {
	return r.db.Model(&model.ShareLink{}).Where("token = ?", token).Update("used_count", gorm.Expr("used_count + 1")).Error
}

// IncrementUsedCountIfNotExceeded 原子地增加使用次数（仅当未超限时）
// 返回值：新的使用次数，如果返回 0 表示已达上限或 token 不存在
func (r *ShareLinkRepository) IncrementUsedCountIfNotExceeded(token string) (int, error) {
	var newCount int
	err := r.db.Raw(`
		UPDATE share_links
		SET used_count = used_count + 1
		WHERE token = ?
		  AND (max_uses = 0 OR used_count < max_uses)
		RETURNING used_count
	`, token).Scan(&newCount).Error

	if err == gorm.ErrRecordNotFound || r.db.RowsAffected == 0 {
		// 更新失败：要么 token 不存在，要么已达上限
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return newCount, nil
}

// UpdateMaxUses 更新最大使用次数
func (r *ShareLinkRepository) UpdateMaxUses(id int64, maxUses int) error {
	return r.db.Model(&model.ShareLink{}).Where("id = ?", id).Update("max_uses", maxUses).Error
}

// Delete 删除分享链接
func (r *ShareLinkRepository) Delete(id int64) error {
	return r.db.Delete(&model.ShareLink{}, id).Error
}

// DeleteByStreamKey 删除直播的所有分享链接
func (r *ShareLinkRepository) DeleteByStreamKey(streamKey string) error {
	return r.db.Where("stream_key = ?", streamKey).Delete(&model.ShareLink{}).Error
}
