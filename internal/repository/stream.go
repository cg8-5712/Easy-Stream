package repository

import (
	"database/sql"
	"time"

	"easy-stream/internal/model"
)

type StreamRepository struct {
	db *sql.DB
}

func NewStreamRepository(db *sql.DB) *StreamRepository {
	return &StreamRepository{db: db}
}

// Create 创建推流
func (r *StreamRepository) Create(stream *model.Stream) error {
	query := `
		INSERT INTO streams (stream_key, name, device_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	now := time.Now()
	return r.db.QueryRow(query,
		stream.StreamKey, stream.Name, stream.DeviceID, stream.Status, now, now,
	).Scan(&stream.ID)
}

// GetByKey 根据 stream_key 获取
func (r *StreamRepository) GetByKey(key string) (*model.Stream, error) {
	query := `SELECT id, stream_key, name, device_id, status, protocol, bitrate, fps,
			  last_frame_at, created_at, updated_at FROM streams WHERE stream_key = $1`

	stream := &model.Stream{}
	err := r.db.QueryRow(query, key).Scan(
		&stream.ID, &stream.StreamKey, &stream.Name, &stream.DeviceID,
		&stream.Status, &stream.Protocol, &stream.Bitrate, &stream.FPS,
		&stream.LastFrameAt, &stream.CreatedAt, &stream.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return stream, err
}

// List 获取推流列表
func (r *StreamRepository) List(status string, offset, limit int) ([]*model.Stream, int64, error) {
	var total int64
	countQuery := "SELECT COUNT(*) FROM streams"
	if status != "" {
		countQuery += " WHERE status = $1"
		r.db.QueryRow(countQuery, status).Scan(&total)
	} else {
		r.db.QueryRow(countQuery).Scan(&total)
	}

	query := `SELECT id, stream_key, name, device_id, status, protocol, bitrate, fps,
			  last_frame_at, created_at, updated_at FROM streams`
	var rows *sql.Rows
	var err error

	if status != "" {
		query += " WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3"
		rows, err = r.db.Query(query, status, limit, offset)
	} else {
		query += " ORDER BY created_at DESC LIMIT $1 OFFSET $2"
		rows, err = r.db.Query(query, limit, offset)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var streams []*model.Stream
	for rows.Next() {
		s := &model.Stream{}
		rows.Scan(&s.ID, &s.StreamKey, &s.Name, &s.DeviceID, &s.Status,
			&s.Protocol, &s.Bitrate, &s.FPS, &s.LastFrameAt, &s.CreatedAt, &s.UpdatedAt)
		streams = append(streams, s)
	}
	return streams, total, nil
}

// Update 更新推流信息
func (r *StreamRepository) Update(stream *model.Stream) error {
	query := `UPDATE streams SET name=$1, device_id=$2, status=$3, protocol=$4,
			  bitrate=$5, fps=$6, last_frame_at=$7, updated_at=$8 WHERE stream_key=$9`
	_, err := r.db.Exec(query, stream.Name, stream.DeviceID, stream.Status,
		stream.Protocol, stream.Bitrate, stream.FPS, stream.LastFrameAt,
		time.Now(), stream.StreamKey)
	return err
}

// UpdateStatus 更新状态
func (r *StreamRepository) UpdateStatus(key, status string) error {
	query := `UPDATE streams SET status=$1, updated_at=$2 WHERE stream_key=$3`
	_, err := r.db.Exec(query, status, time.Now(), key)
	return err
}

// Delete 删除推流
func (r *StreamRepository) Delete(key string) error {
	_, err := r.db.Exec("DELETE FROM streams WHERE stream_key = $1", key)
	return err
}
