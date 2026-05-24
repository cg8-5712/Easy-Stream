package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"os"

	"easy-stream/internal/config"
	"easy-stream/internal/model"
	"easy-stream/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authSvc.Login(req.Username, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// RefreshToken 刷新访问令牌
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authSvc.RefreshToken(req.RefreshToken)
	if err != nil {
		if err == service.ErrInvalidToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	var req model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "logged out"})
		return
	}

	_ = h.authSvc.Logout(req.RefreshToken)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// Profile 获取当前用户信息
func (h *AuthHandler) Profile(c *gin.Context) {
	userID := c.GetInt64("user_id")
	user, err := h.authSvc.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// CheckInitStatus 检查是否已初始化管理员账号
func (h *AuthHandler) CheckInitStatus(c *gin.Context) {
	initialized, err := h.authSvc.IsInitialized()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	configExists := config.ConfigFileExists()
	c.JSON(http.StatusOK, model.InitStatusResponse{
		Initialized:     initialized || configExists,
		ConfigExists:    configExists,
		NeedsAdminSetup: configExists && !initialized,
		CanWriteConfig:  canWriteConfigFile(),
	})
}

// InitializeAdmin 初始化管理员账号和配置
func (h *AuthHandler) InitializeAdmin(c *gin.Context) {
	var req model.InitializeAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	initialized, err := h.authSvc.IsInitialized()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if initialized {
		c.JSON(http.StatusConflict, gin.H{"error": "admin already initialized"})
		return
	}

	configExists := config.ConfigFileExists()
	if !configExists {
		jwtSecret := req.JWTSecret
		if jwtSecret == "" {
			b := make([]byte, 32)
			if _, err := rand.Read(b); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate JWT secret"})
				return
			}
			jwtSecret = hex.EncodeToString(b)
		}

		cfg := &config.Config{
			Server: config.ServerConfig{
				Host: req.ServerHost,
				Port: req.ServerPort,
				Mode: "release",
			},
			Database: config.DatabaseConfig{
				Type:     req.DatabaseType,
				FilePath: req.DatabaseFilePath,
				Host:     req.DatabaseHost,
				Port:     req.DatabasePort,
				User:     req.DatabaseUser,
				Password: req.DatabasePassword,
				DBName:   req.DatabaseName,
				SSLMode:  req.DatabaseSSLMode,
			},
			Redis: config.RedisConfig{
				Host:     req.RedisHost,
				Port:     req.RedisPort,
				Password: req.RedisPassword,
				DB:       req.RedisDB,
			},
			JWT: config.JWTConfig{
				Secret: jwtSecret,
			},
			Admin: config.AdminConfig{
				Username: req.Username,
				Password: "",
			},
			ZLMediaKit: config.ZLMediaKitConfig{
				Host:            req.ZLMHost,
				Port:            req.ZLMPort,
				Secret:          req.ZLMSecret,
				HookBaseURL:     req.ZLMHookBaseURL,
				HTTPPort:        "80",
				HTTPSPort:       "443",
				WebRTCPort:      "8000",
				RecordMode:      "local",
				RecordLocalPath: "./records",
			},
			Log: config.LogConfig{
				Level: "info",
			},
		}

		if cfg.Server.Host == "" {
			cfg.Server.Host = "0.0.0.0"
		}
		if cfg.Database.Type == "sqlite" && cfg.Database.FilePath == "" {
			cfg.Database.FilePath = "./easy_stream.db"
		}
		if cfg.Database.Type == "postgres" && cfg.Database.SSLMode == "" {
			cfg.Database.SSLMode = "disable"
		}

		if err := config.Validate(cfg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "configuration validation failed: " + err.Error()})
			return
		}

		if err := config.SaveConfig(cfg); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save configuration: " + err.Error()})
			return
		}
	}

	if err := h.authSvc.InitializeAdmin(req.Username, req.Password, req.RealName, req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create admin user: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "system initialized successfully"})
}

func canWriteConfigFile() bool {
	f, err := os.OpenFile("config.yaml", os.O_WRONLY|os.O_APPEND, 0)
	if err == nil {
		_ = f.Close()
		return true
	}
	return errors.Is(err, os.ErrNotExist)
}
