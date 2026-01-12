package repository

import (
	"database/sql"

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
	query := `SELECT id, username, password_hash, role, created_at, updated_at
			  FROM users WHERE username = $1`

	user := &model.User{}
	err := r.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Role,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

// GetByID 根据 ID 获取用户
func (r *UserRepository) GetByID(id int64) (*model.User, error) {
	query := `SELECT id, username, password_hash, role, created_at, updated_at
			  FROM users WHERE id = $1`

	user := &model.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Role,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}
