package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// StringArray 自定义类型用于 PostgreSQL JSONB 数组
type StringArray []string

// Scan 实现 sql.Scanner 接口
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = []string{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		*a = []string{}
		return nil
	}
	return json.Unmarshal(bytes, a)
}

// Value 实现 driver.Valuer 接口
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}
	return json.Marshal(a)
}

// RecordFile 录制文件信息
type RecordFile struct {
	FileName  string            `json:"file_name"`      // 文件名
	FilePath  string            `json:"file_path"`      // 文件路径
	FileSize  int64             `json:"file_size"`      // 文件大小（字节）
	Duration  float64           `json:"duration"`       // 录制时长（秒）
	StartTime int64             `json:"start_time"`     // 录制开始时间戳
	TimeLen   float64           `json:"time_len"`       // 录制时长（秒，ZLM 返回）
	CreatedAt time.Time         `json:"created_at"`     // 文件创建时间
	URLs      map[string]string `json:"urls,omitempty"` // 各存储的访问URL
}

// RecordFileArray 录制文件数组
type RecordFileArray []RecordFile

// Scan 实现 sql.Scanner 接口
func (a *RecordFileArray) Scan(value interface{}) error {
	if value == nil {
		*a = []RecordFile{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		*a = []RecordFile{}
		return nil
	}
	return json.Unmarshal(bytes, a)
}

// Value 实现 driver.Valuer 接口
func (a RecordFileArray) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}
	return json.Marshal(a)
}

// Stream 推流信息
type Stream struct {
	ID          int64   `json:"id" db:"id"`
	StreamKey   string  `json:"stream_key" db:"stream_key"`
	Name        string  `json:"name" db:"name"`
	Description *string `json:"description" db:"description"`
	DeviceID    *string `json:"device_id" db:"device_id"`
	Status      string  `json:"status" db:"status"`         // idle / pushing / ended
	Visibility  string  `json:"visibility" db:"visibility"` // public / private
	// 分享码相关字段
	ShareCode          *string         `json:"share_code,omitempty" db:"share_code"`             // 分享码（私有直播自动生成）
	ShareCodeMaxUses   int             `json:"share_code_max_uses" db:"share_code_max_uses"`     // 分享码最大使用次数（0表示无限制）
	ShareCodeUsedCount int             `json:"share_code_used_count" db:"share_code_used_count"` // 分享码已使用次数
	RecordEnabled      bool            `json:"record_enabled" db:"record_enabled"`               // 是否开启录制
	RecordStatus       string          `json:"record_status" db:"record_status"`                 // 录制状态: idle / recording / stopped / failed
	RecordFiles        RecordFileArray `json:"record_files" db:"record_files"`                   // 录制文件列表（包含元数据）
	Protocol           *string         `json:"protocol" db:"protocol"`
	Bitrate            *int            `json:"bitrate" db:"bitrate"`
	FPS                *int            `json:"fps" db:"fps"`
	StreamerName       *string         `json:"streamer_name" db:"streamer_name"`               // 直播人员姓名
	StreamerContact    *string         `json:"streamer_contact" db:"streamer_contact"`         // 直播人员联系方式
	ScheduledStartTime *time.Time      `json:"scheduled_start_time" db:"scheduled_start_time"` // 预计开始时间
	ScheduledEndTime   *time.Time      `json:"scheduled_end_time" db:"scheduled_end_time"`     // 预计结束时间
	AutoKickDelay      int             `json:"auto_kick_delay" db:"auto_kick_delay"`           // 超过预计结束时间后，无推流多久自动结束（分钟）
	ActualStartTime    *time.Time      `json:"actual_start_time" db:"actual_start_time"`       // 实际开始时间
	ActualEndTime      *time.Time      `json:"actual_end_time" db:"actual_end_time"`           // 实际结束时间
	LastUnpublishAt    *time.Time      `json:"last_unpublish_at" db:"last_unpublish_at"`       // 最后断流时间
	LastFrameAt        *time.Time      `json:"last_frame_at" db:"last_frame_at"`
	// 观看统计
	CurrentViewers int       `json:"current_viewers" db:"current_viewers"` // 当前观看人数
	TotalViewers   int       `json:"total_viewers" db:"total_viewers"`     // 累计观看人次
	PeakViewers    int       `json:"peak_viewers" db:"peak_viewers"`       // 峰值观看人数
	CreatedBy      int64     `json:"created_by" db:"created_by"`           // 创建者用户ID
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// StreamStatus 流状态常量
const (
	StreamStatusIdle    = "idle"
	StreamStatusPushing = "pushing"
	StreamStatusEnded   = "ended"
)

// StreamVisibility 流可见性常量
const (
	StreamVisibilityPublic  = "public"
	StreamVisibilityPrivate = "private"
)

// RecordStatus 录制状态常量
const (
	RecordStatusIdle      = "idle"      // 未录制
	RecordStatusRecording = "recording" // 录制中
	RecordStatusStopped   = "stopped"   // 已停止
	RecordStatusFailed    = "failed"    // 录制失败
)

// CreateStreamRequest 创建推流请求
type CreateStreamRequest struct {
	Name               string     `json:"name" binding:"required"`
	Description        string     `json:"description"`
	DeviceID           string     `json:"device_id"`
	Visibility         string     `json:"visibility" binding:"required,oneof=public private"`
	ShareCodeMaxUses   *int       `json:"share_code_max_uses"` // 分享码最大使用次数（仅私有直播有效，0或不传表示无限制）
	RecordEnabled      bool       `json:"record_enabled"`      // 是否开启录制
	StreamerName       string     `json:"streamer_name" binding:"required"`
	StreamerContact    string     `json:"streamer_contact"`
	ScheduledStartTime *time.Time `json:"scheduled_start_time" binding:"required"`
	ScheduledEndTime   *time.Time `json:"scheduled_end_time" binding:"required"`
	AutoKickDelay      int        `json:"auto_kick_delay"` // 超过预计结束时间后无推流多久自动结束，默认30分钟
}

// UpdateStreamRequest 更新推流请求
type UpdateStreamRequest struct {
	Name               string     `json:"name"`
	Description        string     `json:"description"`
	DeviceID           string     `json:"device_id"`
	Visibility         string     `json:"visibility" binding:"omitempty,oneof=public private"`
	RecordEnabled      *bool      `json:"record_enabled"` // 使用指针以区分未传和传 false
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

// StreamListRequest 推流列表查询参数
type StreamListRequest struct {
	Status     string `form:"status"`     // idle / pushing / ended
	Visibility string `form:"visibility"` // public / private
	TimeRange  string `form:"time_range"` // past / current / future
	Page       int    `form:"page"`
	PageSize   int    `form:"pageSize"`
}

// TimeRange 时间范围常量
const (
	TimeRangePast    = "past"    // 已结束的直播
	TimeRangeCurrent = "current" // 正在进行的直播
	TimeRangeFuture  = "future"  // 未开始的直播
)

// StreamAccessToken 直播访问令牌（用于私有直播）
type StreamAccessToken struct {
	StreamID  int64     `json:"stream_id"`
	Token     string    `json:"access_token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// StreamPublicView 游客可见的直播信息（不含 stream_key 和敏感信息）
// 游客不可见：DeviceID, RecordFiles, Protocol, Bitrate, FPS, AutoKickDelay, LastFrameAt, CreatedBy
type StreamPublicView struct {
	ID                 int64      `json:"id"`
	Name               string     `json:"name"`
	Description        *string    `json:"description"`
	Status             string     `json:"status"`
	Visibility         string     `json:"visibility"`
	RecordEnabled      bool       `json:"record_enabled"`
	RecordStatus       string     `json:"record_status"` // 录制状态（游客可见，用于显示"录制中"）
	StreamerName       *string    `json:"streamer_name"`
	StreamerContact    *string    `json:"streamer_contact"`
	ScheduledStartTime *time.Time `json:"scheduled_start_time"`
	ScheduledEndTime   *time.Time `json:"scheduled_end_time"`
	ActualStartTime    *time.Time `json:"actual_start_time"`
	ActualEndTime      *time.Time `json:"actual_end_time"`
	CurrentViewers     int        `json:"current_viewers"`
	TotalViewers       int        `json:"total_viewers"`
	PeakViewers        int        `json:"peak_viewers"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// ToPublicView 将 Stream 转换为 StreamPublicView（游客视图）
func (s *Stream) ToPublicView() *StreamPublicView {
	return &StreamPublicView{
		ID:                 s.ID,
		Name:               s.Name,
		Description:        s.Description,
		Status:             s.Status,
		Visibility:         s.Visibility,
		RecordEnabled:      s.RecordEnabled,
		RecordStatus:       s.RecordStatus,
		StreamerName:       s.StreamerName,
		StreamerContact:    s.StreamerContact,
		ScheduledStartTime: s.ScheduledStartTime,
		ScheduledEndTime:   s.ScheduledEndTime,
		ActualStartTime:    s.ActualStartTime,
		ActualEndTime:      s.ActualEndTime,
		CurrentViewers:     s.CurrentViewers,
		TotalViewers:       s.TotalViewers,
		PeakViewers:        s.PeakViewers,
		CreatedAt:          s.CreatedAt,
		UpdatedAt:          s.UpdatedAt,
	}
}

// StreamPublicListResponse 游客推流列表响应
type StreamPublicListResponse struct {
	Total   int64               `json:"total"`
	Streams []*StreamPublicView `json:"streams"`
}

// PlayURLsResponse 播放地址响应
type PlayURLsResponse struct {
	StreamID   int64             `json:"stream_id"`
	StreamKey  string            `json:"stream_key"`
	StreamName string            `json:"stream_name"`
	Status     string            `json:"status"`
	PlayURLs   map[string]string `json:"play_urls"`
	PushURLs   map[string]string `json:"push_urls,omitempty"` // 仅管理员可见
}

// WebRTCPlayResponse WebRTC 播放/推流响应
type WebRTCPlayResponse struct {
	SDP string `json:"sdp"`
}
