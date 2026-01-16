package service

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"easy-stream/internal/model"
	"easy-stream/internal/repository"
)

type ShareLinkService struct {
	shareLinkRepo *repository.ShareLinkRepository
	streamRepo    *repository.StreamRepository
	redisRepo     *repository.RedisClient
}

func NewShareLinkService(
	shareLinkRepo *repository.ShareLinkRepository,
	streamRepo *repository.StreamRepository,
	redisRepo *repository.RedisClient,
) *ShareLinkService {
	return &ShareLinkService{
		shareLinkRepo: shareLinkRepo,
		streamRepo:    streamRepo,
		redisRepo:     redisRepo,
	}
}

// Create 创建分享链接（管理员）
func (s *ShareLinkService) Create(streamKey string, req *model.CreateShareLinkRequest, userID int64) (*model.CreateShareLinkResponse, error) {
	stream, err := s.streamRepo.GetByKey(streamKey)
	if err != nil {
		return nil, err
	}
	if stream == nil {
		return nil, ErrStreamNotFound
	}

	// 只有私有直播才能创建分享链接
	if stream.Visibility != model.StreamVisibilityPrivate {
		return nil, ErrNotPrivateStream
	}

	// 生成 token
	token, err := s.generateToken()
	if err != nil {
		return nil, err
	}

	link := &model.ShareLink{
		StreamID:  stream.ID,
		Token:     token,
		MaxUses:   req.MaxUses,
		CreatedBy: userID,
	}

	if err := s.shareLinkRepo.Create(link); err != nil {
		return nil, err
	}

	return &model.CreateShareLinkResponse{
		ID:        link.ID,
		Token:     token,
		ShareURL:  "/share/" + token,
		MaxUses:   link.MaxUses,
		UsedCount: 0,
	}, nil
}

// List 获取直播的所有分享链接（管理员）
func (s *ShareLinkService) List(streamKey string) (*model.ShareLinkListResponse, error) {
	stream, err := s.streamRepo.GetByKey(streamKey)
	if err != nil {
		return nil, err
	}
	if stream == nil {
		return nil, ErrStreamNotFound
	}

	links, err := s.shareLinkRepo.ListByStreamID(stream.ID)
	if err != nil {
		return nil, err
	}

	return &model.ShareLinkListResponse{
		Total: int64(len(links)),
		Links: links,
	}, nil
}

// VerifyToken 验证分享链接 token（游客）
func (s *ShareLinkService) VerifyToken(token string) (*model.StreamAccessToken, error) {
	link, err := s.shareLinkRepo.GetByToken(token)
	if err != nil {
		return nil, err
	}
	if link == nil {
		return nil, ErrInvalidShareLink
	}

	// 获取关联的直播
	stream, err := s.streamRepo.GetByID(link.StreamID)
	if err != nil {
		return nil, err
	}
	if stream == nil {
		return nil, ErrStreamNotFound
	}

	// 检查直播是否已结束（分享链接的有效期）
	if stream.Status == model.StreamStatusEnded {
		return nil, ErrStreamEnded
	}

	// 检查使用次数限制
	if link.MaxUses > 0 && link.UsedCount >= link.MaxUses {
		return nil, ErrShareLinkMaxUsesReached
	}

	// 增加使用次数
	if err := s.shareLinkRepo.IncrementUsedCount(token); err != nil {
		return nil, err
	}

	// 生成访问令牌（有效期2小时）
	accessToken, err := s.generateToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(2 * time.Hour)
	if err := s.redisRepo.SetStreamAccessToken(stream.StreamKey, accessToken, 2*time.Hour); err != nil {
		return nil, err
	}

	return &model.StreamAccessToken{
		StreamKey: stream.StreamKey,
		Token:     accessToken,
		ExpiresAt: expiresAt,
	}, nil
}

// UpdateMaxUses 更新分享链接最大使用次数（管理员）
func (s *ShareLinkService) UpdateMaxUses(linkID int64, maxUses int) (*model.ShareLink, error) {
	link, err := s.shareLinkRepo.GetByID(linkID)
	if err != nil {
		return nil, err
	}
	if link == nil {
		return nil, ErrShareLinkNotFound
	}

	if err := s.shareLinkRepo.UpdateMaxUses(linkID, maxUses); err != nil {
		return nil, err
	}

	return s.shareLinkRepo.GetByID(linkID)
}

// Delete 删除分享链接（管理员）
func (s *ShareLinkService) Delete(linkID int64) error {
	link, err := s.shareLinkRepo.GetByID(linkID)
	if err != nil {
		return err
	}
	if link == nil {
		return ErrShareLinkNotFound
	}

	return s.shareLinkRepo.Delete(linkID)
}

// generateToken 生成 token
func (s *ShareLinkService) generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
