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
	"easy-stream/internal/zlm"
	"easy-stream/pkg/logger"
	"easy-stream/pkg/utils"

	"go.uber.org/zap"
)

type StreamService struct {
	streamRepo    *repository.StreamRepository
	shareLinkRepo *repository.ShareLinkRepository
	redisRepo     *repository.RedisClient
	zlmClient     *zlm.Client
	zlmCfg        config.ZLMediaKitConfig
}

func NewStreamService(streamRepo *repository.StreamRepository, shareLinkRepo *repository.ShareLinkRepository, redisRepo *repository.RedisClient, zlmCfg config.ZLMediaKitConfig) *StreamService {
	return &StreamService{
		streamRepo:    streamRepo,
		shareLinkRepo: shareLinkRepo,
		redisRepo:     redisRepo,
		zlmClient:     zlm.NewClient(zlmCfg.Host, zlmCfg.Port, zlmCfg.Secret),
		zlmCfg:        zlmCfg,
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
		RecordStatus:       model.RecordStatusIdle,
		RecordFiles:        model.RecordFileArray{},
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
		if errors.Is(err, repository.ErrStreamNotFound) {
			return nil, ErrStreamNotFound
		}
		return nil, err
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
func (s *StreamService) List(req *model.StreamListRequest, isLoggedIn bool, accessToken string) (*model.StreamListResponse, error) {
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

	// 如果游客传入了 access_token，尝试获取对应的私有直播
	if !isLoggedIn && accessToken != "" {
		logger.Debug("processing access token for guest user",
			zap.Bool("has_access_token", true))
		streamKey, err := s.redisRepo.GetStreamKeyByAccessToken(accessToken)
		logger.Debug("retrieved stream key from access token",
			zap.String("stream_key", streamKey),
			zap.Error(err))
		if err == nil && streamKey != "" {
			// 获取对应的私有直播
			privateStream, err := s.streamRepo.GetByKey(streamKey)
			logger.Debug("retrieved private stream",
				zap.Bool("stream_found", privateStream != nil),
				zap.Error(err))
			if err == nil {
				logger.Debug("checking private stream status",
					zap.String("status", string(privateStream.Status)))
				// 只要不是已结束的直播就可以显示
				if privateStream.Status != model.StreamStatusEnded {
					// 检查是否已经在列表中
					found := false
					for _, stream := range streams {
						if stream.ID == privateStream.ID {
							found = true
							break
						}
					}
					if !found {
						// 将私有直播添加到列表开头
						streams = append([]*model.Stream{privateStream}, streams...)
						total++
						logger.Debug("added private stream to list",
							zap.Int64("stream_id", privateStream.ID))
					}
				}
			} else if !errors.Is(err, repository.ErrStreamNotFound) {
				logger.Error("failed to get private stream",
					zap.String("stream_key", streamKey),
					zap.Error(err))
			}
		}
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
		if errors.Is(err, repository.ErrStreamNotFound) {
			return nil, ErrStreamNotFound
		}
		return nil, err
	}
	return stream, nil
}

// VerifyAccessToken 验证访问令牌
func (s *StreamService) VerifyAccessToken(streamKey, accessToken string) (bool, error) {
	return s.redisRepo.VerifyStreamAccessToken(streamKey, accessToken)
}

// Update 更新推流信息（管理员）
func (s *StreamService) Update(key string, req *model.UpdateStreamRequest) (*model.Stream, error) {
	stream, err := s.streamRepo.GetByKey(key)
	if err != nil {
		if errors.Is(err, repository.ErrStreamNotFound) {
			return nil, ErrStreamNotFound
		}
		return nil, err
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
					logger.Error("failed to start record for stream",
						zap.String("stream_key", key),
						zap.Error(err))
				}
			} else {
				// 关闭录制
				if _, err := s.zlmClient.StopRecord("live", key, zlm.RecordTypeMP4); err != nil {
					logger.Error("failed to stop record for stream",
						zap.String("stream_key", key),
						zap.Error(err))
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

// Kick 强制断流（管理员）- 只断开推流，不结束直播
func (s *StreamService) Kick(key string) error {
	stream, err := s.streamRepo.GetByKey(key)
	if err != nil {
		if errors.Is(err, repository.ErrStreamNotFound) {
			return ErrStreamNotFound
		}
		return err
	}

	// 调用 ZLMediaKit 踢流
	_, err = s.zlmClient.CloseStreams("live", key, true)
	if err != nil {
		return err
	}

	// 状态改为 idle，记录断流时间（OnUnpublish 回调也会处理，这里是备份）
	now := time.Now()
	stream.LastUnpublishAt = &now
	stream.Status = model.StreamStatusIdle
	return s.streamRepo.Update(stream)
}

// End 手动结束直播（管理员）- 断流并标记为结束
func (s *StreamService) End(key string) error {
	stream, err := s.streamRepo.GetByKey(key)
	if err != nil {
		if errors.Is(err, repository.ErrStreamNotFound) {
			return ErrStreamNotFound
		}
		return err
	}

	// 如果正在推流，先断流
	if stream.Status == model.StreamStatusPushing {
		_, _ = s.zlmClient.CloseStreams("live", key, true)
	}

	// 执行结束流程
	return s.endStreamInternal(stream)
}

// endStreamInternal 内部方法：执行结束直播的所有清理工作
func (s *StreamService) endStreamInternal(stream *model.Stream) error {
	streamKey := stream.StreamKey

	// 在事务中执行数据库操作
	if err := s.streamRepo.EndStreamTx(stream, s.shareLinkRepo); err != nil {
		logger.Error("failed to end stream in transaction",
			zap.String("stream_key", streamKey),
			zap.Error(err))
		return err
	}

	// 事务成功后，清理 Redis（即使失败也不影响主流程）
	if err := s.redisRepo.DeleteStreamAccessTokens(streamKey); err != nil {
		logger.Warn("failed to delete access tokens from redis",
			zap.String("stream_key", streamKey),
			zap.Error(err))
	}

	return nil
}

// VerifyShareCode 验证分享码（游客）
func (s *StreamService) VerifyShareCode(shareCode string) (*model.StreamAccessToken, error) {
	stream, err := s.streamRepo.GetByShareCode(shareCode)
	if err != nil {
		if errors.Is(err, repository.ErrStreamNotFound) {
			return nil, ErrInvalidShareCode
		}
		return nil, err
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
		StreamID:  stream.ID,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

// AddShareCode 为直播添加分享码（管理员）
func (s *StreamService) AddShareCode(streamKey string, maxUses int) (*model.Stream, error) {
	stream, err := s.streamRepo.GetByKey(streamKey)
	if err != nil {
		if errors.Is(err, repository.ErrStreamNotFound) {
			return nil, ErrStreamNotFound
		}
		return nil, err
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
		if errors.Is(err, repository.ErrStreamNotFound) {
			return nil, ErrStreamNotFound
		}
		return nil, err
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
		if errors.Is(err, repository.ErrStreamNotFound) {
			return nil, ErrStreamNotFound
		}
		return nil, err
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
	_, err := s.streamRepo.GetByKey(streamKey)
	if err != nil {
		if errors.Is(err, repository.ErrStreamNotFound) {
			return ErrStreamNotFound
		}
		return err
	}

	return s.streamRepo.DeleteShareCode(streamKey)
}

// OnPublish 处理推流开始回调
func (s *StreamService) OnPublish(req *model.OnPublishRequest) error {
	stream, err := s.streamRepo.GetByKey(req.Stream)
	if err != nil {
		if errors.Is(err, repository.ErrStreamNotFound) {
			return ErrStreamNotFound
		}
		return err
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
		// 更新录制状态为 recording
		s.streamRepo.UpdateRecordStatus(req.Stream, model.RecordStatusRecording)

		go func() {
			// 1. 添加 panic 恢复，防止 goroutine 崩溃导致整个程序崩溃
			defer func() {
				if r := recover(); r != nil {
					logger.Error("panic in StartRecord goroutine",
						zap.Any("panic", r),
						zap.String("stream_key", req.Stream))
				}
			}()

			// 2. 使用 channel + select 实现超时控制
			done := make(chan error, 1)
			go func() {
				_, err := s.zlmClient.StartRecord("live", req.Stream, zlm.RecordTypeMP4, "")
				done <- err
			}()

			// 3. 等待结果或超时
			select {
			case err := <-done:
				if err != nil {
					logger.Error("failed to start record for stream",
						zap.String("stream_key", req.Stream),
						zap.Error(err))
					// 录制失败，更新状态为 failed
					s.streamRepo.UpdateRecordStatus(req.Stream, model.RecordStatusFailed)
				}
			case <-time.After(constants.RecordOperationTimeout):
				logger.Error("start record timeout",
					zap.String("stream_key", req.Stream))
				s.streamRepo.UpdateRecordStatus(req.Stream, model.RecordStatusFailed)
			}
		}()
	}

	return nil
}

// OnUnpublish 处理推流结束回调
func (s *StreamService) OnUnpublish(req *model.OnUnpublishRequest) error {
	stream, err := s.streamRepo.GetByKey(req.Stream)
	if err != nil {
		if errors.Is(err, repository.ErrStreamNotFound) {
			return nil
		}
		return err
	}

	// 如果开启了录制，停止录制
	if stream.RecordEnabled && stream.RecordStatus == model.RecordStatusRecording {
		// 更新录制状态为 stopped
		s.streamRepo.UpdateRecordStatus(req.Stream, model.RecordStatusStopped)

		go func() {
			// 1. 添加 panic 恢复，防止 goroutine 崩溃导致整个程序崩溃
			defer func() {
				if r := recover(); r != nil {
					logger.Error("panic in StopRecord goroutine",
						zap.Any("panic", r),
						zap.String("stream_key", req.Stream))
				}
			}()

			// 2. 使用 channel + select 实现超时控制
			done := make(chan error, 1)
			go func() {
				_, err := s.zlmClient.StopRecord("live", req.Stream, zlm.RecordTypeMP4)
				done <- err
			}()

			// 3. 等待结果或超时
			select {
			case err := <-done:
				if err != nil {
					logger.Error("failed to stop record for stream",
						zap.String("stream_key", req.Stream),
						zap.Error(err))
				}
			case <-time.After(constants.RecordOperationTimeout):
				logger.Error("stop record timeout",
					zap.String("stream_key", req.Stream))
			}
		}()
	}

	// 记录断流时间，状态改为 idle（等待自动结束或重新推流）
	now := time.Now()
	stream.LastUnpublishAt = &now
	stream.Status = model.StreamStatusIdle
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
// 当播放器或推流者断开连接时会触发此回调
func (s *StreamService) OnFlowReport(req *model.OnFlowReportRequest) error {
	_, err := s.streamRepo.GetByKey(req.Stream)
	if err != nil {
		if errors.Is(err, repository.ErrStreamNotFound) {
			return nil
		}
		return err
	}

	// Player=true 表示是播放者，断开时减少观看人数
	if req.Player {
		logger.Info("player disconnected via flow report",
			zap.String("stream", req.Stream),
			zap.Int64("totalBytes", req.TotalBytes))
		return s.streamRepo.DecrementViewers(req.Stream)
	}

	return nil
}

// CheckExpiredStreams 检查并处理超时的直播（定时任务）
func (s *StreamService) CheckExpiredStreams() error {
	now := time.Now()

	// 检查 idle 状态的流，超过预计结束时间 + auto_kick_delay 后自动结束
	idleStreams, err := s.streamRepo.GetIdleStreams()
	if err != nil {
		return err
	}

	for _, stream := range idleStreams {
		if stream.ScheduledEndTime == nil {
			continue
		}

		// 计算自动结束时间：预计结束时间 + AutoKickDelay
		autoEndTime := stream.ScheduledEndTime.Add(time.Duration(stream.AutoKickDelay) * time.Minute)

		// 如果已超时且没有在推流，自动结束直播
		if now.After(autoEndTime) {
			logger.Info("auto ending stream past scheduled end time",
				zap.String("stream_key", stream.StreamKey),
				zap.Int("auto_kick_delay_minutes", stream.AutoKickDelay))
			s.endStreamInternal(stream)
		}
	}

	return nil
}

// AddRecordFile 添加录制文件（包含完整元数据）
func (s *StreamService) AddRecordFile(streamKey string, recordFile *model.RecordFile) error {
	return s.streamRepo.AppendRecordFile(streamKey, recordFile)
}

// UpdateRecordStatus 更新录制状态
func (s *StreamService) UpdateRecordStatus(streamKey, status string) error {
	return s.streamRepo.UpdateRecordStatus(streamKey, status)
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
	b := make([]byte, constants.TokenByteLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// strPtr 将字符串转换为指针
func strPtr(s string) *string {
	return &s
}

// GetPlayURLs 获取播放地址（支持游客和管理员）
func (s *StreamService) GetPlayURLs(streamKey string, isAdmin bool) (*model.PlayURLsResponse, error) {
	stream, err := s.streamRepo.GetByKey(streamKey)
	if err != nil {
		if errors.Is(err, repository.ErrStreamNotFound) {
			return nil, ErrStreamNotFound
		}
		return nil, err
	}

	// 私有直播不允许游客访问
	if stream.Visibility == model.StreamVisibilityPrivate && !isAdmin {
		return nil, ErrStreamNotFound
	}

	// 构建播放地址
	host := s.zlmCfg.ExternalHost
	if host == "" {
		host = s.zlmCfg.Host
	}

	httpPort := s.zlmCfg.HTTPPort
	if httpPort == "" {
		httpPort = "80"
	}

	webrtcPort := s.zlmCfg.WebRTCPort
	if webrtcPort == "" {
		webrtcPort = "8000"
	}

	playURLs := map[string]string{
		"webrtc":   fmt.Sprintf("webrtc://%s:%s/live/%s", host, webrtcPort, streamKey),
		"hls":      fmt.Sprintf("http://%s:%s/live/%s/hls.m3u8", host, httpPort, streamKey),
		"http_flv": fmt.Sprintf("http://%s:%s/live/%s.live.flv", host, httpPort, streamKey),
		"ws_flv":   fmt.Sprintf("ws://%s:%s/live/%s.live.flv", host, httpPort, streamKey),
	}

	resp := &model.PlayURLsResponse{
		StreamID:   stream.ID,
		StreamKey:  streamKey,
		StreamName: stream.Name,
		Status:     stream.Status,
		PlayURLs:   playURLs,
	}

	// 管理员才返回推流地址
	if isAdmin {
		resp.PushURLs = map[string]string{
			"rtmp":    fmt.Sprintf("rtmp://%s:1935/live/%s", host, streamKey),
			"rtsp":    fmt.Sprintf("rtsp://%s:554/live/%s", host, streamKey),
			"srt":     fmt.Sprintf("srt://%s:9000?streamid=#!::r=live/%s,m=publish", host, streamKey),
			"http_ts": fmt.Sprintf("http://%s:%s/live/%s.live.ts", host, httpPort, streamKey),
		}
	}

	return resp, nil
}

// WebRTCPush 处理 WebRTC 推流
func (s *StreamService) WebRTCPush(streamKey, offerSDP string) (*model.WebRTCPlayResponse, error) {
	// 1. 验证 stream_key 是否存在
	stream, err := s.streamRepo.GetByKey(streamKey)
	if err != nil {
		if errors.Is(err, repository.ErrStreamNotFound) {
			return nil, ErrStreamNotFound
		}
		return nil, err
	}

	// 2. 检查直播状态（不能是已结束）
	if stream.Status == model.StreamStatusEnded {
		return nil, ErrStreamExpired
	}

	// 3. 调用 ZLM WebRTC Push API
	resp, err := s.zlmClient.WebRTCPush("live", streamKey, offerSDP)
	if err != nil {
		return nil, err
	}

	return &model.WebRTCPlayResponse{
		SDP: resp.SDP,
	}, nil
}
