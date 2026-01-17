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
