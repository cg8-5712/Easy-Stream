package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"easy-stream/internal/config"
	"easy-stream/internal/constants"
	"easy-stream/internal/model"
	"easy-stream/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  *repository.UserRepository
	redisRepo *repository.RedisClient
	jwtCfg    config.JWTConfig
}

func NewAuthService(userRepo *repository.UserRepository, redisRepo *repository.RedisClient, jwtCfg config.JWTConfig) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		redisRepo: redisRepo,
		jwtCfg:    jwtCfg,
	}
}

// Login 用户登录
func (s *AuthService) Login(username, password string) (*model.LoginResponse, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLoginAt = &now
	s.userRepo.UpdateLastLogin(user.ID, now)

	// 生成 Access Token (短期，2小时)
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	// 生成 Refresh Token (长期，7天)
	refreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// 将 Refresh Token 存储到 Redis (7天过期)
	if err := s.redisRepo.SetRefreshToken(user.ID, refreshToken, constants.RefreshTokenExpiry); err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(constants.AccessTokenExpiry.Seconds()), // 转换为秒
		User:         user,
	}, nil
}

// RefreshToken 刷新访问令牌
func (s *AuthService) RefreshToken(refreshToken string) (*model.RefreshTokenResponse, error) {
	// 从 Redis 验证 Refresh Token
	userID, err := s.redisRepo.GetUserIDByRefreshToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// 获取用户信息
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}

	// 生成新的 Access Token
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	// 生成新的 Refresh Token
	newRefreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// 删除旧的 Refresh Token，存储新的
	s.redisRepo.DeleteRefreshToken(refreshToken)
	if err := s.redisRepo.SetRefreshToken(user.ID, newRefreshToken, constants.RefreshTokenExpiry); err != nil {
		return nil, err
	}

	return &model.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(constants.AccessTokenExpiry.Seconds()), // 转换为秒
	}, nil
}

// Logout 登出（撤销 Refresh Token）
func (s *AuthService) Logout(refreshToken string) error {
	return s.redisRepo.DeleteRefreshToken(refreshToken)
}

// GetUserByID 获取用户信息
func (s *AuthService) GetUserByID(id int64) (*model.User, error) {
	return s.userRepo.GetByID(id)
}

// generateAccessToken 生成访问令牌（短期）
func (s *AuthService) generateAccessToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"type":     "access",
		"exp":      time.Now().Add(constants.AccessTokenExpiry).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtCfg.Secret))
}

// generateRefreshToken 生成刷新令牌（长期）
func (s *AuthService) generateRefreshToken(userID int64) (string, error) {
	// 生成随机字符串
	b := make([]byte, constants.TokenByteLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%d_%s", userID, hex.EncodeToString(b)), nil
}

