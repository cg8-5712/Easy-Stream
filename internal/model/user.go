package model

import "time"

// User 用户信息
type User struct {
	ID           int64      `json:"id" db:"id" gorm:"column:id;primaryKey"`
	Username     string     `json:"username" db:"username" gorm:"column:username;uniqueIndex;not null"`
	PasswordHash string     `json:"-" db:"password_hash" gorm:"column:password_hash;not null"`
	Email        *string    `json:"email" db:"email" gorm:"column:email"`
	Phone        *string    `json:"phone" db:"phone" gorm:"column:phone"`
	RealName     *string    `json:"real_name" db:"real_name" gorm:"column:real_name"`
	Avatar       *string    `json:"avatar" db:"avatar" gorm:"column:avatar"`
	LastLoginAt  *time.Time `json:"last_login_at" db:"last_login_at" gorm:"column:last_login_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // 秒
	User         *User  `json:"user"`
}

// RefreshTokenRequest 刷新 Token 请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse 刷新 Token 响应
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// InitializeAdminRequest 初始化管理员请求
type InitializeAdminRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	RealName string `json:"real_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	// 数据库配置
	DatabaseType     string `json:"database_type" binding:"required,oneof=sqlite postgres mysql"`
	DatabaseFilePath string `json:"database_filepath"` // SQLite
	DatabaseHost     string `json:"database_host"`     // PostgreSQL/MySQL
	DatabasePort     string `json:"database_port"`     // PostgreSQL/MySQL
	DatabaseUser     string `json:"database_user"`     // PostgreSQL/MySQL
	DatabasePassword string `json:"database_password"` // PostgreSQL/MySQL
	DatabaseName     string `json:"database_name"`     // PostgreSQL/MySQL
	DatabaseSSLMode  string `json:"database_sslmode"`  // PostgreSQL
	// Redis 配置
	RedisHost     string `json:"redis_host" binding:"required"`
	RedisPort     string `json:"redis_port" binding:"required"`
	RedisPassword string `json:"redis_password"`
	RedisDB       int    `json:"redis_db"`
	// ZLMediaKit 配置
	ZLMHost        string `json:"zlm_host" binding:"required"`
	ZLMPort        string `json:"zlm_port" binding:"required"`
	ZLMSecret      string `json:"zlm_secret"`
	ZLMHookBaseURL string `json:"zlm_hook_base_url"`
	// JWT 配置
	JWTSecret string `json:"jwt_secret" binding:"omitempty,min=16"`
	// 服务器配置
	ServerHost string `json:"server_host"`
	ServerPort string `json:"server_port" binding:"required"`
}

// InitStatusResponse 初始化状态响应
type InitStatusResponse struct {
	Initialized bool `json:"initialized"`
}
