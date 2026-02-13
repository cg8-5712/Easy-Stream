package repository

import (
	"crypto/rand"
	"math/big"
	"time"

	"easy-stream/internal/config"
	"easy-stream/internal/model"
	"easy-stream/pkg/logger"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedData 在 debug 模式下插入测试数据
func SeedData(db *gorm.DB, cfg *config.Config) error {
	logger.Info("debug mode: starting to insert seed data")

	// 插入测试用户
	if err := seedUsers(db, cfg); err != nil {
		return err
	}

	// 插入测试直播流
	if err := seedStreams(db); err != nil {
		return err
	}

	logger.Info("debug mode: seed data insertion completed")
	return nil
}

// generateRandomPassword 生成随机密码（避免易混淆字符）
func generateRandomPassword(length int) (string, error) {
	// 避免易混淆的字符：0 O o, 1 l I i, 2 Z z
	const charset = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXY3456789"
	password := make([]byte, length)

	for i := range password {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		password[i] = charset[num.Int64()]
	}

	return string(password), nil
}

// seedUsers 插入测试用户
func seedUsers(db *gorm.DB, cfg *config.Config) error {
	// 检查 admin 用户是否已存在
	var count int64
	db.Model(&model.User{}).Where("username = ?", cfg.Admin.Username).Count(&count)
	if count > 0 {
		logger.Debug("admin user already exists, skipping seed")
		return nil
	}

	// 确定管理员密码
	adminPassword := cfg.Admin.Password
	if adminPassword == "" {
		// 生成随机密码
		var err error
		adminPassword, err = generateRandomPassword(12)
		if err != nil {
			return err
		}
		logger.Warn("admin password not configured, generated random password",
			zap.String("username", cfg.Admin.Username),
			zap.String("password", adminPassword))
	}

	// 生成密码哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 创建管理员用户
	realName := "Administrator"
	email := "admin@example.com"
	user := &model.User{
		Username:     cfg.Admin.Username,
		PasswordHash: string(hash),
		RealName:     &realName,
		Email:        &email,
	}

	if err := db.Create(user).Error; err != nil {
		return err
	}

	logger.Info("admin user created",
		zap.String("username", cfg.Admin.Username))

	return nil
}

// seedStreams 插入测试直播流
func seedStreams(db *gorm.DB) error {
	// 获取 admin 用户 ID
	var admin model.User
	if err := db.Where("username = ?", "admin").First(&admin).Error; err != nil {
		return err
	}

	streams := []struct {
		streamKey   string
		name        string
		description string
		visibility  string
		status      string
		streamer    string
	}{
		{"test-stream-001", "测试直播间1-正在直播", "这是一个正在直播的公开直播间", "public", "pushing", "测试主播A"},
		{"test-stream-002", "测试直播间2-未开始", "这是一个未开始的公开直播间", "public", "idle", "测试主播B"},
		{"test-stream-003", "测试直播间3-已结束", "这是一个已结束的公开直播间", "public", "ended", "测试主播C"},
		{"test-stream-004", "测试直播间4-私密直播", "这是一个私密的测试直播间", "private", "pushing", "测试主播D"},
	}

	now := time.Now()
	scheduledStart := now.Add(-1 * time.Hour)
	scheduledEnd := now.Add(24 * time.Hour)

	for _, s := range streams {
		// 检查直播流是否已存在
		var count int64
		db.Model(&model.Stream{}).Where("stream_key = ?", s.streamKey).Count(&count)
		if count > 0 {
			continue
		}

		// 创建直播流
		stream := &model.Stream{
			StreamKey:          s.streamKey,
			Name:               s.name,
			Description:        &s.description,
			Visibility:         s.visibility,
			Status:             s.status,
			StreamerName:       &s.streamer,
			ScheduledStartTime: &scheduledStart,
			ScheduledEndTime:   &scheduledEnd,
			AutoKickDelay:      30,
			RecordEnabled:      false,
			RecordStatus:       model.RecordStatusIdle,
			CreatedBy:          admin.ID,
		}

		if err := db.Create(stream).Error; err != nil {
			return err
		}

		logger.Debug("debug mode: created stream",
			zap.String("name", s.name),
			zap.String("stream_key", s.streamKey),
			zap.String("status", s.status))
	}

	return nil
}
