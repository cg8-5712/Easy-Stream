package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"easy-stream/internal/config"
	"easy-stream/internal/model"
	"easy-stream/internal/repository"
	"easy-stream/internal/zlm"
	"easy-stream/pkg/utils"

	"golang.org/x/crypto/bcrypt"
)

type StreamService struct {
	streamRepo *repository.StreamRepository
	redisRepo  *repository.RedisClient
	zlmClient  *zlm.Client
}

func NewStreamService(streamRepo *repository.StreamRepository, redisRepo *repository.RedisClient, zlmCfg config.ZLMediaKitConfig) *StreamService {
	return &StreamService{
		streamRepo: streamRepo,
		redisRepo:  redisRepo,
		zlmClient:  zlm.NewClient(zlmCfg.Host, zlmCfg.Port, zlmCfg.Secret),
	}
}

// Create 创建推流码（管理员）
func (s *StreamService) Create(req *model.CreateStreamRequest, userID int64) (*model.Stream, error) {
	// 验证时间
	if req.ScheduledEndTime.Before(*req.ScheduledStartTime) {
		return nil, fmt.Errorf("scheduled end time must be after start time")
	}

	// 如果是私有直播，必须设置密码
	if req.Visibility == model.StreamVisibilityPrivate && req.Password == "" {
		return nil, ErrPrivateStream
	}

	// 设置默认超时时间（30分钟）
	autoKickDelay := req.AutoKickDelay
	if autoKickDelay == 0 {
		autoKickDelay = 30
	}

	// 加密密码
	var passwordHash string
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		passwordHash = string(hash)
	}

	stream := &model.Stream{
		StreamKey:          utils.GenerateStreamKey(),
		Name:               req.Name,
		Description:        req.Description,
		DeviceID:           req.DeviceID,
		Status:             model.StreamStatusIdle,
		Visibility:         req.Visibility,
		Password:           passwordHash,
		RecordEnabled:      req.RecordEnabled,
		RecordFiles:        model.StringArray{},
		StreamerName:       req.StreamerName,
		StreamerContact:    req.StreamerContact,
		ScheduledStartTime: req.ScheduledStartTime,
		ScheduledEndTime:   req.ScheduledEndTime,
		AutoKickDelay:      autoKickDelay,
		CreatedBy:          userID,
	}

	if err := s.streamRepo.Create(stream); err != nil {
		return nil, err
	}
	return stream, nil
}

// Get 获取推流信息（支持游客和管理员）
func (s *StreamService) Get(key string, userRole string, accessToken string) (*model.Stream, error) {
	stream, err := s.streamRepo.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if stream == nil {
		return nil, ErrStreamNotFound
	}

	// 管理员可以查看所有直播
	if userRole == model.UserRoleAdmin || userRole == model.UserRoleOperator {
		return stream, nil
	}

	// 公开直播，游客可以查看
	if stream.Visibility == model.StreamVisibilityPublic {
		return stream, nil
	}

	// 私有直播，需要验证访问令牌
	if accessToken != "" {
		valid, err := s.redisRepo.VerifyStreamAccessToken(key, accessToken)
		if err == nil && valid {
			return stream, nil
		}
	}

	return nil, ErrPrivateStream
}

// List 获取推流列表（游客只能看公开且正在直播的，管理员能看所有）
func (s *StreamService) List(req *model.StreamListRequest, userRole string) (*model.StreamListResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	offset := (req.Page - 1) * req.PageSize

	// 游客只能看公开且正在直播的
	if userRole != model.UserRoleAdmin && userRole != model.UserRoleOperator {
		req.Visibility = model.StreamVisibilityPublic
		req.Status = model.StreamStatusPushing
		req.TimeRange = "" // 游客不能使用时间范围过滤
	}

	streams, total, err := s.streamRepo.List(req, offset, req.PageSize)
	if err != nil {
		return nil, err
	}

	return &model.StreamListResponse{
		Total:   total,
		Streams: streams,
	}, nil
}

// GetByID 通过 ID 获取推流信息（管理员）
func (s *StreamService) GetByID(id int64) (*model.Stream, error) {
	stream, err := s.streamRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if stream == nil {
		return nil, ErrStreamNotFound
	}
	return stream, nil
}

// Update 更新推流信息（管理员）
func (s *StreamService) Update(key string, req *model.UpdateStreamRequest) (*model.Stream, error) {
	stream, err := s.streamRepo.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if stream == nil {
		return nil, ErrStreamNotFound
	}

	// 更新字段
	if req.Name != "" {
		stream.Name = req.Name
	}
	if req.Description != "" {
		stream.Description = req.Description
	}
	if req.DeviceID != "" {
		stream.DeviceID = req.DeviceID
	}
	if req.Visibility != "" {
		stream.Visibility = req.Visibility
	}
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		stream.Password = string(hash)
	}
	if req.StreamerName != "" {
		stream.StreamerName = req.StreamerName
	}
	if req.StreamerContact != "" {
		stream.StreamerContact = req.StreamerContact
	}
	if req.ScheduledStartTime != nil {
		stream.ScheduledStartTime = req.ScheduledStartTime
	}
	if req.ScheduledEndTime != nil {
		stream.ScheduledEndTime = req.ScheduledEndTime
	}
	if req.AutoKickDelay != nil {
		stream.AutoKickDelay = *req.AutoKickDelay
	}

	// 处理动态录制开关
	if req.RecordEnabled != nil {
		oldRecordEnabled := stream.RecordEnabled
		newRecordEnabled := *req.RecordEnabled

		// 如果录制状态发生变化且正在推流
		if oldRecordEnabled != newRecordEnabled && stream.Status == model.StreamStatusPushing {
			if newRecordEnabled {
				// 开启录制
				if _, err := s.zlmClient.StartRecord("live", key, zlm.RecordTypeMP4, ""); err != nil {
					// 记录错误但不阻止更新
					fmt.Printf("failed to start record for stream %s: %v\n", key, err)
				}
			} else {
				// 关闭录制
				if _, err := s.zlmClient.StopRecord("live", key, zlm.RecordTypeMP4); err != nil {
					fmt.Printf("failed to stop record for stream %s: %v\n", key, err)
				}
			}
		}
		stream.RecordEnabled = newRecordEnabled
	}

	if err := s.streamRepo.Update(stream); err != nil {
		return nil, err
	}
	return stream, nil
}

// Delete 删除推流码（管理员）
func (s *StreamService) Delete(key string) error {
	return s.streamRepo.Delete(key)
}

// Kick 强制断流（管理员）
func (s *StreamService) Kick(key string) error {
	stream, err := s.streamRepo.GetByKey(key)
	if err != nil {
		return err
	}
	if stream == nil {
		return ErrStreamNotFound
	}

	// 调用 ZLMediaKit 踢流
	_, err = s.zlmClient.CloseStreams("live", key, true)
	if err != nil {
		return err
	}

	// 更新状态
	now := time.Now()
	stream.ActualEndTime = &now
	return s.streamRepo.UpdateStatus(key, model.StreamStatusIdle)
}

// VerifyPassword 验证私有直播密码（游客）
func (s *StreamService) VerifyPassword(key, password string) (*model.StreamAccessToken, error) {
	stream, err := s.streamRepo.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if stream == nil {
		return nil, ErrStreamNotFound
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(stream.Password), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}

	// 生成访问令牌（有效期2小时）
	token, err := s.generateAccessToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(2 * time.Hour)
	if err := s.redisRepo.SetStreamAccessToken(key, token, 2*time.Hour); err != nil {
		return nil, err
	}

	return &model.StreamAccessToken{
		StreamKey: key,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

// OnPublish 处理推流开始回调
func (s *StreamService) OnPublish(req *model.OnPublishRequest) error {
	stream, err := s.streamRepo.GetByKey(req.Stream)
	if err != nil {
		return err
	}
	if stream == nil {
		return ErrStreamNotFound
	}

	// 检查流状态，已结束的流不允许再次推流
	if stream.Status == model.StreamStatusEnded {
		return ErrStreamExpired
	}

	// 更新状态和实际开始时间
	now := time.Now()
	stream.Status = model.StreamStatusPushing
	stream.Protocol = req.Schema
	stream.ActualStartTime = &now

	if err := s.streamRepo.Update(stream); err != nil {
		return err
	}

	// 如果开启了录制，自动开始录制
	if stream.RecordEnabled {
		go func() {
			if _, err := s.zlmClient.StartRecord("live", req.Stream, zlm.RecordTypeMP4, ""); err != nil {
				fmt.Printf("failed to start record for stream %s: %v\n", req.Stream, err)
			}
		}()
	}

	return nil
}

// OnUnpublish 处理推流结束回调
func (s *StreamService) OnUnpublish(req *model.OnUnpublishRequest) error {
	stream, err := s.streamRepo.GetByKey(req.Stream)
	if err != nil {
		return err
	}
	if stream == nil {
		return nil
	}

	// 如果开启了录制，停止录制
	if stream.RecordEnabled {
		go func() {
			if _, err := s.zlmClient.StopRecord("live", req.Stream, zlm.RecordTypeMP4); err != nil {
				fmt.Printf("failed to stop record for stream %s: %v\n", req.Stream, err)
			}
		}()
	}

	// 更新实际结束时间
	now := time.Now()
	stream.ActualEndTime = &now
	stream.Status = model.StreamStatusEnded
	return s.streamRepo.Update(stream)
}

// OnFlowReport 处理流量统计回调
func (s *StreamService) OnFlowReport(req *model.OnFlowReportRequest) error {
	stream, err := s.streamRepo.GetByKey(req.Stream)
	if err != nil || stream == nil {
		return err
	}
	// 可以在这里更新码率等信息
	return nil
}

// CheckExpiredStreams 检查并断开超时的直播（定时任务）
func (s *StreamService) CheckExpiredStreams() error {
	// 获取所有正在推流的流
	streams, err := s.streamRepo.GetPushingStreams()
	if err != nil {
		return err
	}

	now := time.Now()
	for _, stream := range streams {
		if stream.ScheduledEndTime == nil {
			continue
		}

		// 计算超时时间：预计结束时间 + 延迟时间
		expireTime := stream.ScheduledEndTime.Add(time.Duration(stream.AutoKickDelay) * time.Minute)

		// 如果已超时，强制断流
		if now.After(expireTime) {
			s.Kick(stream.StreamKey)
		}
	}

	return nil
}

// AddRecordFile 添加录制文件路径
func (s *StreamService) AddRecordFile(streamKey, filePath string) error {
	return s.streamRepo.AppendRecordFile(streamKey, filePath)
}

// generateAccessToken 生成访问令牌
func (s *StreamService) generateAccessToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
