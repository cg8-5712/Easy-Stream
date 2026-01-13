package model

import "time"

// Stream 推流信息
type Stream struct {
	ID                 int64      `json:"id" db:"id"`
	StreamKey          string     `json:"stream_key" db:"stream_key"`
	Name               string     `json:"name" db:"name"`
	Description        string     `json:"description" db:"description"`
	DeviceID           string     `json:"device_id" db:"device_id"`
	Status             string     `json:"status" db:"status"` // idle / pushing / destroyed
	Visibility         string     `json:"visibility" db:"visibility"` // public / private
	Password           string     `json:"-" db:"password"` // 私有直播密码
	Protocol           string     `json:"protocol" db:"protocol"`
	Bitrate            int        `json:"bitrate" db:"bitrate"`
	FPS                int        `json:"fps" db:"fps"`
	StreamerName       string     `json:"streamer_name" db:"streamer_name"` // 直播人员姓名
	StreamerContact    string     `json:"streamer_contact" db:"streamer_contact"` // 直播人员联系方式
	ScheduledStartTime *time.Time `json:"scheduled_start_time" db:"scheduled_start_time"` // 预计开始时间
	ScheduledEndTime   *time.Time `json:"scheduled_end_time" db:"scheduled_end_time"` // 预计结束时间
	AutoKickDelay      int        `json:"auto_kick_delay" db:"auto_kick_delay"` // 超时自动断流延迟（分钟）
	ActualStartTime    *time.Time `json:"actual_start_time" db:"actual_start_time"` // 实际开始时间
	ActualEndTime      *time.Time `json:"actual_end_time" db:"actual_end_time"` // 实际结束时间
	LastFrameAt        *time.Time `json:"last_frame_at" db:"last_frame_at"`
	CreatedBy          int64      `json:"created_by" db:"created_by"` // 创建者用户ID
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

// StreamStatus 流状态常量
const (
	StreamStatusIdle      = "idle"
	StreamStatusPushing   = "pushing"
	StreamStatusDestroyed = "destroyed"
)

// StreamVisibility 流可见性常量
const (
	StreamVisibilityPublic  = "public"
	StreamVisibilityPrivate = "private"
)

// CreateStreamRequest 创建推流请求
type CreateStreamRequest struct {
	Name               string     `json:"name" binding:"required"`
	Description        string     `json:"description"`
	DeviceID           string     `json:"device_id"`
	Visibility         string     `json:"visibility" binding:"required,oneof=public private"`
	Password           string     `json:"password"` // 私有直播时必填
	StreamerName       string     `json:"streamer_name" binding:"required"`
	StreamerContact    string     `json:"streamer_contact"`
	ScheduledStartTime *time.Time `json:"scheduled_start_time" binding:"required"`
	ScheduledEndTime   *time.Time `json:"scheduled_end_time" binding:"required"`
	AutoKickDelay      int        `json:"auto_kick_delay"` // 默认30分钟
}

// UpdateStreamRequest 更新推流请求
type UpdateStreamRequest struct {
	Name               string     `json:"name"`
	Description        string     `json:"description"`
	DeviceID           string     `json:"device_id"`
	Visibility         string     `json:"visibility" binding:"omitempty,oneof=public private"`
	Password           string     `json:"password"`
	StreamerName       string     `json:"streamer_name"`
	StreamerContact    string     `json:"streamer_contact"`
	ScheduledStartTime *time.Time `json:"scheduled_start_time"`
	ScheduledEndTime   *time.Time `json:"scheduled_end_time"`
	AutoKickDelay      *int       `json:"auto_kick_delay"`
}

// StreamListResponse 推流列表响应
type StreamListResponse struct {
	Total   int64     `json:"total"`
	Streams []*Stream `json:"streams"`
}

// VerifyStreamPasswordRequest 验证私有直播密码请求
type VerifyStreamPasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

// StreamAccessToken 直播访问令牌（用于私有直播）
type StreamAccessToken struct {
	StreamKey string    `json:"stream_key"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}
