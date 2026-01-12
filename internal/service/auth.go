package service

import (
	"time"

	"easy-stream/internal/config"
	"easy-stream/internal/model"
	"easy-stream/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repository.UserRepository
	jwtCfg   config.JWTConfig
}

func NewAuthService(userRepo *repository.UserRepository, jwtCfg config.JWTConfig) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwtCfg:   jwtCfg,
	}
}

// Login 用户登录
func (s *AuthService) Login(username, password string) (*model.LoginResponse, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// 生成 JWT
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

// GetUserByID 获取用户信息
func (s *AuthService) GetUserByID(id int64) (*model.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *AuthService) generateToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * time.Duration(s.jwtCfg.ExpireHour)).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtCfg.Secret))
}
