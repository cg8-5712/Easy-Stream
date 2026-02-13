package repository

import (
	"easy-stream/internal/model"
	"easy-stream/pkg/logger"

	"gorm.io/gorm"
)

// AutoMigrate 执行数据库自动迁移
func AutoMigrate(db *gorm.DB) error {
	logger.Info("starting database auto migration")

	// 自动迁移所有模型
	err := db.AutoMigrate(
		&model.User{},
		&model.Stream{},
		&model.ShareLink{},
	)

	if err != nil {
		return err
	}

	logger.Info("database auto migration completed")
	return nil
}
