package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	JWT        JWTConfig
	Admin      AdminConfig
	ZLMediaKit ZLMediaKitConfig
	Log        LogConfig
	Storage    StorageConfig
}

type ServerConfig struct {
	Host string
	Port string
	Mode string // debug / release
}

type DatabaseConfig struct {
	Type     string // postgres / mysql / sqlite
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	FilePath string // SQLite 数据库文件路径
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret string
}

type AdminConfig struct {
	Username string
	Password string
}

type ZLMediaKitConfig struct {
	Host         string
	Port         string
	Secret       string
	HookBaseURL  string // Hook 回调基础 URL，如 http://localhost:8080/api/v1/hooks
	HTTPPort     string // HLS/FLV 播放端口，默认 "80"
	HTTPSPort    string // HTTPS 端口，默认 "443"
	ExternalHost string // 外部访问地址（生成播放 URL）
	WebRTCPort   string // WebRTC 端口，默认 "8000"
	// 录制文件配置
	RecordMode      string // 录制文件访问模式: local（本地文件系统）/ remote（远程HTTP访问）
	RecordLocalPath string // 本地模式：录制文件本地路径（与ZLM的record_path对应）
	RecordBaseURL   string // 远程模式：录制文件访问基础URL（如 http://zlm-server:8080/record）
}

type LogConfig struct {
	Level string // debug / info / warn / error
}

// StorageConfig 存储配置
type StorageConfig struct {
	Targets []StorageTarget `mapstructure:"targets"` // 多个存储目标
}

// StorageTarget 存储目标配置
type StorageTarget struct {
	Name     string `mapstructure:"name"`     // 存储名称标识
	Type     string `mapstructure:"type"`     // 类型: local / s3 / cos / oss
	Enabled  bool   `mapstructure:"enabled"`  // 是否启用
	Default  bool   `mapstructure:"default"`  // 是否为默认存储
	LocalDir string `mapstructure:"localDir"` // 本地存储目录（type=local时使用）
	// S3 兼容存储配置（S3/COS/OSS 通用）
	Endpoint        string `mapstructure:"endpoint"`        // 端点地址
	Region          string `mapstructure:"region"`          // 区域
	Bucket          string `mapstructure:"bucket"`          // 存储桶名称
	AccessKeyID     string `mapstructure:"accessKeyId"`     // 访问密钥ID
	SecretAccessKey string `mapstructure:"secretAccessKey"` // 访问密钥
	PathPrefix      string `mapstructure:"pathPrefix"`      // 存储路径前缀
	CustomDomain    string `mapstructure:"customDomain"`    // 自定义域名（用于生成访问URL）
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// 设置默认值
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.filepath", "./easy_stream.db")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("admin.username", "admin")
	viper.SetDefault("admin.password", "")
	viper.SetDefault("zlmediakit.port", "80")
	viper.SetDefault("zlmediakit.httpPort", "80")
	viper.SetDefault("zlmediakit.httpsPort", "443")
	viper.SetDefault("zlmediakit.webrtcPort", "8000")
	viper.SetDefault("zlmediakit.recordMode", "local")
	viper.SetDefault("log.level", "info")

	// 支持环境变量
	viper.AutomaticEnv()
	// 设置环境变量键替换（将 . 替换为 _）
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 绑定关键环境变量
	viper.BindEnv("zlmediakit.hookbaseurl", "ZLMEDIAKIT_HOOKBASEURL")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
