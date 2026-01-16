package model

import "time"

// ShareLink 分享链接
type ShareLink struct {
	ID        int64     `json:"id" db:"id"`
	StreamID  int64     `json:"stream_id" db:"stream_id"`
	Token     string    `json:"token" db:"token"`
	MaxUses   int       `json:"max_uses" db:"max_uses"`     // 最大使用次数（0表示无限制）
	UsedCount int       `json:"used_count" db:"used_count"` // 已使用次数
	CreatedBy int64     `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// CreateShareLinkRequest 创建分享链接请求
type CreateShareLinkRequest struct {
	MaxUses int `json:"max_uses"` // 最大使用次数，0表示无限制
}

// CreateShareLinkResponse 创建分享链接响应
type CreateShareLinkResponse struct {
	ID        int64  `json:"id"`
	Token     string `json:"token"`
	ShareURL  string `json:"share_url"` // 完整的分享链接
	MaxUses   int    `json:"max_uses"`
	UsedCount int    `json:"used_count"`
}

// ShareLinkListResponse 分享链接列表响应
type ShareLinkListResponse struct {
	Total int64        `json:"total"`
	Links []*ShareLink `json:"links"`
}

// VerifyShareCodeRequest 验证分享码请求
type VerifyShareCodeRequest struct {
	ShareCode string `json:"share_code" binding:"required"`
}

// RegenerateShareCodeRequest 重新生成分享码请求
type RegenerateShareCodeRequest struct {
	MaxUses *int `json:"max_uses"` // 最大使用次数，0或不传表示无限制
}
