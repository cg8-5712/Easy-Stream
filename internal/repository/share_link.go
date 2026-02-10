package repository

import (
	"database/sql"
	"time"

	"easy-stream/internal/model"
)

type ShareLinkRepository struct {
	db *sql.DB
}

func NewShareLinkRepository(db *sql.DB) *ShareLinkRepository {
	return &ShareLinkRepository{db: db}
}

// Create 创建分享链接
func (r *ShareLinkRepository) Create(link *model.ShareLink) error {
	query := `
		INSERT INTO share_links (stream_key, token, max_uses, used_count, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	return r.db.QueryRow(query,
		link.StreamKey, link.Token, link.MaxUses, 0, link.CreatedBy, time.Now(),
	).Scan(&link.ID)
}

// GetByToken 根据 token 获取分享链接
func (r *ShareLinkRepository) GetByToken(token string) (*model.ShareLink, error) {
	query := `
		SELECT id, stream_key, token, max_uses, used_count, created_by, created_at
		FROM share_links WHERE token = $1
	`
	link := &model.ShareLink{}
	err := r.db.QueryRow(query, token).Scan(
		&link.ID, &link.StreamKey, &link.Token,
		&link.MaxUses, &link.UsedCount, &link.CreatedBy, &link.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return link, err
}

// GetByID 根据 ID 获取分享链接
func (r *ShareLinkRepository) GetByID(id int64) (*model.ShareLink, error) {
	query := `
		SELECT id, stream_key, token, max_uses, used_count, created_by, created_at
		FROM share_links WHERE id = $1
	`
	link := &model.ShareLink{}
	err := r.db.QueryRow(query, id).Scan(
		&link.ID, &link.StreamKey, &link.Token,
		&link.MaxUses, &link.UsedCount, &link.CreatedBy, &link.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return link, err
}

// ListByStreamKey 获取直播的所有分享链接
func (r *ShareLinkRepository) ListByStreamKey(streamKey string) ([]*model.ShareLink, error) {
	query := `
		SELECT id, stream_key, token, max_uses, used_count, created_by, created_at
		FROM share_links WHERE stream_key = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, streamKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make([]*model.ShareLink, 0)
	for rows.Next() {
		link := &model.ShareLink{}
		err := rows.Scan(
			&link.ID, &link.StreamKey, &link.Token,
			&link.MaxUses, &link.UsedCount, &link.CreatedBy, &link.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, nil
}

// IncrementUsedCount 增加使用次数
func (r *ShareLinkRepository) IncrementUsedCount(token string) error {
	query := `UPDATE share_links SET used_count = used_count + 1 WHERE token = $1`
	_, err := r.db.Exec(query, token)
	return err
}

// IncrementUsedCountIfNotExceeded 原子地增加使用次数（仅当未超限时）
// 返回值：新的使用次数，如果返回 0 表示已达上限或 token 不存在
func (r *ShareLinkRepository) IncrementUsedCountIfNotExceeded(token string) (int, error) {
	query := `
		UPDATE share_links
		SET used_count = used_count + 1
		WHERE token = $1
		  AND (max_uses = 0 OR used_count < max_uses)
		RETURNING used_count
	`
	var newCount int
	err := r.db.QueryRow(query, token).Scan(&newCount)
	if err == sql.ErrNoRows {
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
	query := `UPDATE share_links SET max_uses = $1 WHERE id = $2`
	_, err := r.db.Exec(query, maxUses, id)
	return err
}

// Delete 删除分享链接
func (r *ShareLinkRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM share_links WHERE id = $1", id)
	return err
}

// DeleteByStreamKey 删除直播的所有分享链接
func (r *ShareLinkRepository) DeleteByStreamKey(streamKey string) error {
	_, err := r.db.Exec("DELETE FROM share_links WHERE stream_key = $1", streamKey)
	return err
}
