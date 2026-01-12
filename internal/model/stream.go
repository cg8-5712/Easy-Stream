package model

import "time"

// Stream 推流信息
type Stream struct {
	ID          int64      `json:"id" db:"id"`
	StreamKey   string     `json:"stream_key" db:"stream_key"`
	Name        string     `json:"name" db:"name"`
	DeviceID    string     `json:"device_id" db:"device_id"`
	Status      string     `json:"status" db:"status"` // idle / pushing / destroyed
	Protocol    string     `json:"protocol" db:"protocol"`
	Bitrate     int        `json:"bitrate" db:"bitrate"`
	FPS         int        `json:"fps" db:"fps"`
	LastFrameAt *time.Time `json:"last_frame_at" db:"last_frame_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// StreamStatus 流状态常量
const (
	StreamStatusIdle      = "idle"
	StreamStatusPushing   = "pushing"
	StreamStatusDestroyed = "destroyed"
)

// CreateStreamRequest 创建推流请求
type CreateStreamRequest struct {
	Name     string `json:"name" binding:"required"`
	DeviceID string `json:"device_id"`
}

// UpdateStreamRequest 更新推流请求
type UpdateStreamRequest struct {
	Name     string `json:"name"`
	DeviceID string `json:"device_id"`
}

// StreamListResponse 推流列表响应
type StreamListResponse struct {
	Total   int64     `json:"total"`
	Streams []*Stream `json:"streams"`
}
