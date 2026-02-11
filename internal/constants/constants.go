package constants

import "time"

// 密码相关常量
const (
	// BcryptCost bcrypt 哈希成本（12-14 推荐用于生产环境）
	BcryptCost = 12
)

// JWT 相关常量
const (
	// AccessTokenExpiry 访问令牌有效期
	AccessTokenExpiry = 2 * time.Hour
	// RefreshTokenExpiry 刷新令牌有效期
	RefreshTokenExpiry = 7 * 24 * time.Hour
)

// 分页相关常量
const (
	// DefaultPageSize 默认分页大小
	DefaultPageSize = 20
	// MaxPageSize 最大分页大小
	MaxPageSize = 100
	// MinPageSize 最小分页大小
	MinPageSize = 1
)

// 分享链接相关常量
const (
	// MaxShareCodeMaxUses 分享码最大使用次数限制
	MaxShareCodeMaxUses = 10000
	// MinShareCodeMaxUses 分享码最小使用次数
	MinShareCodeMaxUses = 1
)

// 直播相关常量
const (
	// MaxAutoKickDelay 最大自动踢流延迟（分钟）
	MaxAutoKickDelay = 1440 // 24小时
	// MinAutoKickDelay 最小自动踢流延迟（分钟）
	MinAutoKickDelay = 0
	// DefaultAutoKickDelay 默认自动踢流延迟（分钟）
	DefaultAutoKickDelay = 30
)

// 超时相关常量
const (
	// RecordOperationTimeout 录制操作超时时间
	RecordOperationTimeout = 30 * time.Second
)

// Token 相关常量
const (
	// TokenByteLength Token 字节长度
	TokenByteLength = 32
)
