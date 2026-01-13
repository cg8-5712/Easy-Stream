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
			record_enabled, record_files,
			streamer_name, streamer_contact, scheduled_start_time, scheduled_end_time,
			auto_kick_delay, created_by, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id
	`
	now := time.Now()
	recordFiles, _ := stream.RecordFiles.Value()
	return r.db.QueryRow(query,
		stream.StreamKey, stream.Name, stream.Description, stream.DeviceID,
		stream.Status, stream.Visibility, stream.Password,
		stream.RecordEnabled, recordFiles,
		stream.StreamerName, stream.StreamerContact,
		stream.ScheduledStartTime, stream.ScheduledEndTime,
		stream.AutoKickDelay, stream.CreatedBy, now, now,
	).Scan(&stream.ID)
}

// GetByKey 根据 stream_key 获取
func (r *StreamRepository) GetByKey(key string) (*model.Stream, error) {
	query := `
		SELECT id, stream_key, name, description, device_id, status, visibility, password,
			   record_enabled, record_files,
			   protocol, bitrate, fps, streamer_name, streamer_contact,
			   scheduled_start_time, scheduled_end_time, auto_kick_delay,
			   actual_start_time, actual_end_time, last_frame_at,
			   current_viewers, total_viewers, peak_viewers,
			   created_by, created_at, updated_at
		FROM streams WHERE stream_key = $1
	`

	stream := &model.Stream{}
	err := r.db.QueryRow(query, key).Scan(
		&stream.ID, &stream.StreamKey, &stream.Name, &stream.Description,
		&stream.DeviceID, &stream.Status, &stream.Visibility, &stream.Password,
		&stream.RecordEnabled, &stream.RecordFiles,
		&stream.Protocol, &stream.Bitrate, &stream.FPS,
		&stream.StreamerName, &stream.StreamerContact,
		&stream.ScheduledStartTime, &stream.ScheduledEndTime, &stream.AutoKickDelay,
		&stream.ActualStartTime, &stream.ActualEndTime, &stream.LastFrameAt,
		&stream.CurrentViewers, &stream.TotalViewers, &stream.PeakViewers,
		&stream.CreatedBy, &stream.CreatedAt, &stream.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return stream, err
}

// GetByID 根据 ID 获取
func (r *StreamRepository) GetByID(id int64) (*model.Stream, error) {
	query := `
		SELECT id, stream_key, name, description, device_id, status, visibility, password,
			   record_enabled, record_files,
			   protocol, bitrate, fps, streamer_name, streamer_contact,
			   scheduled_start_time, scheduled_end_time, auto_kick_delay,
			   actual_start_time, actual_end_time, last_frame_at,
			   current_viewers, total_viewers, peak_viewers,
			   created_by, created_at, updated_at
		FROM streams WHERE id = $1
	`

	stream := &model.Stream{}
	err := r.db.QueryRow(query, id).Scan(
		&stream.ID, &stream.StreamKey, &stream.Name, &stream.Description,
		&stream.DeviceID, &stream.Status, &stream.Visibility, &stream.Password,
		&stream.RecordEnabled, &stream.RecordFiles,
		&stream.Protocol, &stream.Bitrate, &stream.FPS,
		&stream.StreamerName, &stream.StreamerContact,
		&stream.ScheduledStartTime, &stream.ScheduledEndTime, &stream.AutoKickDelay,
		&stream.ActualStartTime, &stream.ActualEndTime, &stream.LastFrameAt,
		&stream.CurrentViewers, &stream.TotalViewers, &stream.PeakViewers,
		&stream.CreatedBy, &stream.CreatedAt, &stream.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return stream, err
}

// List 获取推流列表
func (r *StreamRepository) List(req *model.StreamListRequest, offset, limit int) ([]*model.Stream, int64, error) {
	// 构建查询条件
	var conditions []string
	var args []interface{}
	argIndex := 1

	if req.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, req.Status)
		argIndex++
	}

	if req.Visibility != "" {
		conditions = append(conditions, fmt.Sprintf("visibility = $%d", argIndex))
		args = append(args, req.Visibility)
		argIndex++
	}

	// 时间范围过滤
	if req.TimeRange != "" {
		switch req.TimeRange {
		case model.TimeRangePast:
			// 已结束的直播
			conditions = append(conditions, "actual_end_time IS NOT NULL")
		case model.TimeRangeCurrent:
			// 正在进行的直播
			conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
			args = append(args, model.StreamStatusPushing)
			argIndex++
		case model.TimeRangeFuture:
			// 未开始的直播
			conditions = append(conditions, fmt.Sprintf("status = $%d AND scheduled_start_time > $%d", argIndex, argIndex+1))
			args = append(args, model.StreamStatusIdle, time.Now())
			argIndex += 2
		}
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
			   record_enabled, record_files,
			   protocol, bitrate, fps, streamer_name, streamer_contact,
			   scheduled_start_time, scheduled_end_time, auto_kick_delay,
			   actual_start_time, actual_end_time, last_frame_at,
			   current_viewers, total_viewers, peak_viewers,
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
			&s.RecordEnabled, &s.RecordFiles,
			&s.Protocol, &s.Bitrate, &s.FPS,
			&s.StreamerName, &s.StreamerContact,
			&s.ScheduledStartTime, &s.ScheduledEndTime, &s.AutoKickDelay,
			&s.ActualStartTime, &s.ActualEndTime, &s.LastFrameAt,
			&s.CurrentViewers, &s.TotalViewers, &s.PeakViewers,
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
			record_enabled=$7, record_files=$8,
			protocol=$9, bitrate=$10, fps=$11,
			streamer_name=$12, streamer_contact=$13,
			scheduled_start_time=$14, scheduled_end_time=$15, auto_kick_delay=$16,
			actual_start_time=$17, actual_end_time=$18, last_frame_at=$19,
			current_viewers=$20, total_viewers=$21, peak_viewers=$22,
			updated_at=$23
		WHERE stream_key=$24
	`
	recordFiles, _ := stream.RecordFiles.Value()
	_, err := r.db.Exec(query,
		stream.Name, stream.Description, stream.DeviceID, stream.Status,
		stream.Visibility, stream.Password,
		stream.RecordEnabled, recordFiles,
		stream.Protocol, stream.Bitrate, stream.FPS,
		stream.StreamerName, stream.StreamerContact,
		stream.ScheduledStartTime, stream.ScheduledEndTime, stream.AutoKickDelay,
		stream.ActualStartTime, stream.ActualEndTime, stream.LastFrameAt,
		stream.CurrentViewers, stream.TotalViewers, stream.PeakViewers,
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

// AppendRecordFile 追加录制文件路径
func (r *StreamRepository) AppendRecordFile(key, filePath string) error {
	query := `
		UPDATE streams
		SET record_files = record_files || $1::jsonb, updated_at = $2
		WHERE stream_key = $3
	`
	fileJSON := fmt.Sprintf(`["%s"]`, filePath)
	_, err := r.db.Exec(query, fileJSON, time.Now(), key)
	return err
}

// UpdateRecordEnabled 更新录制状态
func (r *StreamRepository) UpdateRecordEnabled(key string, enabled bool) error {
	query := `UPDATE streams SET record_enabled=$1, updated_at=$2 WHERE stream_key=$3`
	_, err := r.db.Exec(query, enabled, time.Now(), key)
	return err
}

// GetPushingStreams 获取所有正在推流的直播
func (r *StreamRepository) GetPushingStreams() ([]*model.Stream, error) {
	query := `
		SELECT id, stream_key, name, description, device_id, status, visibility, password,
			   record_enabled, record_files,
			   protocol, bitrate, fps, streamer_name, streamer_contact,
			   scheduled_start_time, scheduled_end_time, auto_kick_delay,
			   actual_start_time, actual_end_time, last_frame_at,
			   current_viewers, total_viewers, peak_viewers,
			   created_by, created_at, updated_at
		FROM streams WHERE status = $1
	`
	rows, err := r.db.Query(query, model.StreamStatusPushing)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var streams []*model.Stream
	for rows.Next() {
		s := &model.Stream{}
		err := rows.Scan(
			&s.ID, &s.StreamKey, &s.Name, &s.Description,
			&s.DeviceID, &s.Status, &s.Visibility, &s.Password,
			&s.RecordEnabled, &s.RecordFiles,
			&s.Protocol, &s.Bitrate, &s.FPS,
			&s.StreamerName, &s.StreamerContact,
			&s.ScheduledStartTime, &s.ScheduledEndTime, &s.AutoKickDelay,
			&s.ActualStartTime, &s.ActualEndTime, &s.LastFrameAt,
			&s.CurrentViewers, &s.TotalViewers, &s.PeakViewers,
			&s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		streams = append(streams, s)
	}
	return streams, nil
}

// IncrementViewers 增加观看人数（有人进入观看）
func (r *StreamRepository) IncrementViewers(key string) error {
	query := `
		UPDATE streams SET
			current_viewers = current_viewers + 1,
			total_viewers = total_viewers + 1,
			peak_viewers = GREATEST(peak_viewers, current_viewers + 1),
			updated_at = $1
		WHERE stream_key = $2
	`
	_, err := r.db.Exec(query, time.Now(), key)
	return err
}

// DecrementViewers 减少观看人数（有人离开）
func (r *StreamRepository) DecrementViewers(key string) error {
	query := `
		UPDATE streams SET
			current_viewers = GREATEST(0, current_viewers - 1),
			updated_at = $1
		WHERE stream_key = $2
	`
	_, err := r.db.Exec(query, time.Now(), key)
	return err
}

// ResetCurrentViewers 重置当前观看人数（直播结束时调用）
func (r *StreamRepository) ResetCurrentViewers(key string) error {
	query := `UPDATE streams SET current_viewers = 0, updated_at = $1 WHERE stream_key = $2`
	_, err := r.db.Exec(query, time.Now(), key)
	return err
}

// Delete 删除推流
func (r *StreamRepository) Delete(key string) error {
	_, err := r.db.Exec("DELETE FROM streams WHERE stream_key = $1", key)
	return err
}
