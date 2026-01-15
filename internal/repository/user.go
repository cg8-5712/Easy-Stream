package repository

import (
	"database/sql"
	"time"

	"easy-stream/internal/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByUsername 根据用户名获取用户
func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	query := `SELECT id, username, password_hash, email, phone, real_name,
			  avatar, last_login_at, created_at, updated_at
			  FROM users WHERE username = $1`

	user := &model.User{}
	err := r.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash,
		&user.Email, &user.Phone, &user.RealName, &user.Avatar,
		&user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

// GetByID 根据 ID 获取用户
func (r *UserRepository) GetByID(id int64) (*model.User, error) {
	query := `SELECT id, username, password_hash, email, phone, real_name,
			  avatar, last_login_at, created_at, updated_at
			  FROM users WHERE id = $1`

	user := &model.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.PasswordHash,
		&user.Email, &user.Phone, &user.RealName, &user.Avatar,
		&user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

// UpdateLastLogin 更新最后登录时间
func (r *UserRepository) UpdateLastLogin(id int64, loginTime time.Time) error {
	query := `UPDATE users SET last_login_at = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(query, loginTime, time.Now(), id)
	return err
}

