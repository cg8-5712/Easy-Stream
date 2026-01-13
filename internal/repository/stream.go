package repository

import (
	"database/sql"
	"fmt"
	"strings"
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
		INSERT INTO streams (
			stream_key, name, description, device_id, status, visibility, password,
			streamer_name, streamer_contact, scheduled_start_time, scheduled_end_time,
			auto_kick_delay, created_by, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id
	`
	now := time.Now()
	return r.db.QueryRow(query,
		stream.StreamKey, stream.Name, stream.Description, stream.DeviceID,
		stream.Status, stream.Visibility, stream.Password,
		stream.StreamerName, stream.StreamerContact,
		stream.ScheduledStartTime, stream.ScheduledEndTime,
		stream.AutoKickDelay, stream.CreatedBy, now, now,
	).Scan(&stream.ID)
}

// GetByKey 根据 stream_key 获取
func (r *StreamRepository) GetByKey(key string) (*model.Stream, error) {
	query := `
		SELECT id, stream_key, name, description, device_id, status, visibility, password,
			   protocol, bitrate, fps, streamer_name, streamer_contact,
			   scheduled_start_time, scheduled_end_time, auto_kick_delay,
			   actual_start_time, actual_end_time, last_frame_at,
			   created_by, created_at, updated_at
		FROM streams WHERE stream_key = $1
	`

	stream := &model.Stream{}
	err := r.db.QueryRow(query, key).Scan(
		&stream.ID, &stream.StreamKey, &stream.Name, &stream.Description,
		&stream.DeviceID, &stream.Status, &stream.Visibility, &stream.Password,
		&stream.Protocol, &stream.Bitrate, &stream.FPS,
		&stream.StreamerName, &stream.StreamerContact,
		&stream.ScheduledStartTime, &stream.ScheduledEndTime, &stream.AutoKickDelay,
		&stream.ActualStartTime, &stream.ActualEndTime, &stream.LastFrameAt,
		&stream.CreatedBy, &stream.CreatedAt, &stream.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return stream, err
}

// List 获取推流列表
func (r *StreamRepository) List(status string, visibility string, offset, limit int) ([]*model.Stream, int64, error) {
	// 构建查询条件
	var conditions []string
	var args []interface{}
	argIndex := 1

	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	if visibility != "" {
		conditions = append(conditions, fmt.Sprintf("visibility = $%d", argIndex))
		args = append(args, visibility)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// 查询总数
	var total int64
	countQuery := "SELECT COUNT(*) FROM streams" + whereClause
	if len(args) > 0 {
		r.db.QueryRow(countQuery, args...).Scan(&total)
	} else {
		r.db.QueryRow(countQuery).Scan(&total)
	}

	// 查询列表
	query := `
		SELECT id, stream_key, name, description, device_id, status, visibility,
			   protocol, bitrate, fps, streamer_name, streamer_contact,
			   scheduled_start_time, scheduled_end_time, auto_kick_delay,
			   actual_start_time, actual_end_time, last_frame_at,
			   created_by, created_at, updated_at
		FROM streams` + whereClause + ` ORDER BY created_at DESC LIMIT $` +
		fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)

	args = append(args, limit, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var streams []*model.Stream
	for rows.Next() {
		s := &model.Stream{}
		err := rows.Scan(
			&s.ID, &s.StreamKey, &s.Name, &s.Description,
			&s.DeviceID, &s.Status, &s.Visibility,
			&s.Protocol, &s.Bitrate, &s.FPS,
			&s.StreamerName, &s.StreamerContact,
			&s.ScheduledStartTime, &s.ScheduledEndTime, &s.AutoKickDelay,
			&s.ActualStartTime, &s.ActualEndTime, &s.LastFrameAt,
			&s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		streams = append(streams, s)
	}
	return streams, total, nil
}

// Update 更新推流信息
func (r *StreamRepository) Update(stream *model.Stream) error {
	query := `
		UPDATE streams SET
			name=$1, description=$2, device_id=$3, status=$4, visibility=$5, password=$6,
			protocol=$7, bitrate=$8, fps=$9,
			streamer_name=$10, streamer_contact=$11,
			scheduled_start_time=$12, scheduled_end_time=$13, auto_kick_delay=$14,
			actual_start_time=$15, actual_end_time=$16, last_frame_at=$17,
			updated_at=$18
		WHERE stream_key=$19
	`
	_, err := r.db.Exec(query,
		stream.Name, stream.Description, stream.DeviceID, stream.Status,
		stream.Visibility, stream.Password,
		stream.Protocol, stream.Bitrate, stream.FPS,
		stream.StreamerName, stream.StreamerContact,
		stream.ScheduledStartTime, stream.ScheduledEndTime, stream.AutoKickDelay,
		stream.ActualStartTime, stream.ActualEndTime, stream.LastFrameAt,
		time.Now(), stream.StreamKey,
	)
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
