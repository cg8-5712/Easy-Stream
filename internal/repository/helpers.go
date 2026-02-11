package repository

import (
	"errors"
	"time"

	"easy-stream/internal/model"

	"gorm.io/gorm"
)

// GetSingleRecord fetches a single record by field and handles not found errors
// This helper reduces duplication in GetByID, GetByKey, GetByUsername patterns
func GetSingleRecord(db *gorm.DB, dest interface{}, field string, value interface{}, notFoundErr error) error {
	err := db.Where(field+" = ?", value).First(dest).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return notFoundErr
	}
	return err
}

// UpdateStreamByKeyWithTimestamp updates a stream by stream_key and sets updated_at
// This helper reduces duplication in stream repository update methods
func UpdateStreamByKeyWithTimestamp(db *gorm.DB, streamKey string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return db.Model(&model.Stream{}).Where("stream_key = ?", streamKey).Updates(updates).Error
}

// HandleNotFoundError converts gorm.ErrRecordNotFound to a specific error
// This helper reduces duplication in error handling across repositories
func HandleNotFoundError(err error, notFoundErr error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return notFoundErr
	}
	return err
}

