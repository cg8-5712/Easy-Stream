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

	// 设置默认超时时间（30分钟）
	autoKickDelay := req.AutoKickDelay
	if autoKickDelay == 0 {
		autoKickDelay = 30
	}

	stream := &model.Stream{
		StreamKey:          utils.GenerateStreamKey(),
		Name:               req.Name,
		Description:        strPtr(req.Description),
		DeviceID:           strPtr(req.DeviceID),
		Status:             model.StreamStatusIdle,
		Visibility:         req.Visibility,
		RecordEnabled:      req.RecordEnabled,
		RecordFiles:        model.StringArray{},
		StreamerName:       strPtr(req.StreamerName),
		StreamerContact:    strPtr(req.StreamerContact),
		ScheduledStartTime: req.ScheduledStartTime,
		ScheduledEndTime:   req.ScheduledEndTime,
		AutoKickDelay:      autoKickDelay,
		CreatedBy:          userID,
	}

	// 如果是私有直播，自动生成分享码
	if req.Visibility == model.StreamVisibilityPrivate {
		shareCode := s.generateShareCode()
		stream.ShareCode = &shareCode

		// 设置分享码最大使用次数
		if req.ShareCodeMaxUses != nil {
			stream.ShareCodeMaxUses = *req.ShareCodeMaxUses
		}
	}

	if err := s.streamRepo.Create(stream); err != nil {
		return nil, err
	}
	return stream, nil
}

// Get 获取推流信息（支持游客和管理员）
func (s *StreamService) Get(key string, isLoggedIn bool, accessToken string) (*model.Stream, error) {
	stream, err := s.streamRepo.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if stream == nil {
		return nil, ErrStreamNotFound
	}

	// 登录用户可以查看所有直播
	if isLoggedIn {
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
func (s *StreamService) List(req *model.StreamListRequest, isLoggedIn bool) (*model.StreamListResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	offset := (req.Page - 1) * req.PageSize

	// 未登录用户只能看公开且正在直播的
	if !isLoggedIn {
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
		stream.Description = strPtr(req.Description)
	}
	if req.DeviceID != "" {
		stream.DeviceID = strPtr(req.DeviceID)
	}
	if req.Visibility != "" {
		// 如果从公开改为私有，自动生成分享码
		if req.Visibility == model.StreamVisibilityPrivate && stream.Visibility == model.StreamVisibilityPublic {
			shareCode := s.generateShareCode()
			stream.ShareCode = &shareCode
			stream.ShareCodeMaxUses = 0
			stream.ShareCodeUsedCount = 0
		}
		// 如果从私有改为公开，清除分享码
		if req.Visibility == model.StreamVisibilityPublic && stream.Visibility == model.StreamVisibilityPrivate {
			stream.ShareCode = nil
			stream.ShareCodeMaxUses = 0
			stream.ShareCodeUsedCount = 0
		}
		stream.Visibility = req.Visibility
	}
	if req.StreamerName != "" {
		stream.StreamerName = strPtr(req.StreamerName)
	}
	if req.StreamerContact != "" {
		stream.StreamerContact = strPtr(req.StreamerContact)
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

// VerifyShareCode 验证分享码（游客）
func (s *StreamService) VerifyShareCode(shareCode string) (*model.StreamAccessToken, error) {
	stream, err := s.streamRepo.GetByShareCode(shareCode)
	if err != nil {
		return nil, err
	}
	if stream == nil {
		return nil, ErrInvalidShareCode
	}

	// 检查直播是否已结束
	if stream.Status == model.StreamStatusEnded {
		return nil, ErrStreamEnded
	}

	// 检查使用次数限制
	if stream.ShareCodeMaxUses > 0 && stream.ShareCodeUsedCount >= stream.ShareCodeMaxUses {
		return nil, ErrShareCodeMaxUsesReached
	}

	// 增加使用次数
	if err := s.streamRepo.IncrementShareCodeUsedCount(stream.StreamKey); err != nil {
		return nil, err
	}

	// 生成访问令牌（有效期2小时）
	token, err := s.generateAccessToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(2 * time.Hour)
	if err := s.redisRepo.SetStreamAccessToken(stream.StreamKey, token, 2*time.Hour); err != nil {
		return nil, err
	}

	return &model.StreamAccessToken{
		StreamKey: stream.StreamKey,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

// AddShareCode 为直播添加分享码（管理员）
func (s *StreamService) AddShareCode(streamKey string, maxUses int) (*model.Stream, error) {
	stream, err := s.streamRepo.GetByKey(streamKey)
	if err != nil {
		return nil, err
	}
	if stream == nil {
		return nil, ErrStreamNotFound
	}

	// 只有私有直播才能添加分享码
	if stream.Visibility != model.StreamVisibilityPrivate {
		return nil, ErrNotPrivateStream
	}

	shareCode := s.generateShareCode()
	if err := s.streamRepo.UpdateShareCode(streamKey, shareCode, maxUses); err != nil {
		return nil, err
	}

	return s.streamRepo.GetByKey(streamKey)
}

// RegenerateShareCode 重新生成分享码（管理员）
func (s *StreamService) RegenerateShareCode(streamKey string, req *model.RegenerateShareCodeRequest) (*model.Stream, error) {
	stream, err := s.streamRepo.GetByKey(streamKey)
	if err != nil {
		return nil, err
	}
	if stream == nil {
		return nil, ErrStreamNotFound
	}

	// 只有私有直播才能生成分享码
	if stream.Visibility != model.StreamVisibilityPrivate {
		return nil, ErrNotPrivateStream
	}

	shareCode := s.generateShareCode()
	maxUses := 0
	if req.MaxUses != nil {
		maxUses = *req.MaxUses
	}

	if err := s.streamRepo.UpdateShareCode(streamKey, shareCode, maxUses); err != nil {
		return nil, err
	}

	return s.streamRepo.GetByKey(streamKey)
}

// UpdateShareCodeMaxUses 更新分享码最大使用次数（管理员）
func (s *StreamService) UpdateShareCodeMaxUses(streamKey string, maxUses int) (*model.Stream, error) {
	stream, err := s.streamRepo.GetByKey(streamKey)
	if err != nil {
		return nil, err
	}
	if stream == nil {
		return nil, ErrStreamNotFound
	}

	if stream.ShareCode == nil {
		return nil, ErrNoShareCode
	}

	if err := s.streamRepo.UpdateShareCodeMaxUses(streamKey, maxUses); err != nil {
		return nil, err
	}

	return s.streamRepo.GetByKey(streamKey)
}

// DeleteShareCode 删除分享码（管理员）
func (s *StreamService) DeleteShareCode(streamKey string) error {
	stream, err := s.streamRepo.GetByKey(streamKey)
	if err != nil {
		return err
	}
	if stream == nil {
		return ErrStreamNotFound
	}

	return s.streamRepo.DeleteShareCode(streamKey)
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
	stream.Protocol = strPtr(req.Schema)
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

	// 清理 Redis 中的访问令牌（分享码和分享链接生成的令牌）
	if err := s.redisRepo.DeleteStreamAccessTokens(req.Stream); err != nil {
		fmt.Printf("failed to delete access tokens for stream %s: %v\n", req.Stream, err)
	}

	// 重置当前观看人数
	s.streamRepo.ResetCurrentViewers(req.Stream)

	// 更新实际结束时间
	now := time.Now()
	stream.ActualEndTime = &now
	stream.Status = model.StreamStatusEnded
	stream.CurrentViewers = 0
	return s.streamRepo.Update(stream)
}

// OnPlay 处理播放开始回调
func (s *StreamService) OnPlay(req *model.OnPlayRequest) error {
	// 增加观看人数
	return s.streamRepo.IncrementViewers(req.Stream)
}

// OnPlayerDisconnect 处理播放器断开回调
func (s *StreamService) OnPlayerDisconnect(req *model.OnPlayerDisconnectRequest) error {
	// 减少观看人数
	return s.streamRepo.DecrementViewers(req.Stream)
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

// generateShareCode 生成6位分享码
func (s *StreamService) generateShareCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // 排除易混淆字符 I,O,0,1
	b := make([]byte, 6)
	rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

// generateAccessToken 生成访问令牌
func (s *StreamService) generateAccessToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// strPtr 将字符串转换为指针
func strPtr(s string) *string {
	return &s
}
