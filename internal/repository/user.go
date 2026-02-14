package repository

import (
	"time"

	"easy-stream/internal/model"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByUsername 根据用户名获取用户
func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	return &user, HandleNotFoundError(
		r.db.Where("username = ?", username).First(&user).Error,
		ErrUserNotFound,
	)
}

// GetByID 根据 ID 获取用户
func (r *UserRepository) GetByID(id int64) (*model.User, error) {
	var user model.User
	return &user, HandleNotFoundError(
		r.db.First(&user, id).Error,
		ErrUserNotFound,
	)
}

// UpdateLastLogin 更新最后登录时间
func (r *UserRepository) UpdateLastLogin(id int64, loginTime time.Time) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_login_at": loginTime,
		"updated_at":    time.Now(),
	}).Error
}

// HasAnyUser 检查是否存在任何用户
func (r *UserRepository) HasAnyUser() (bool, error) {
	var count int64
	if err := r.db.Model(&model.User{}).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// Create 创建用户
func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}
