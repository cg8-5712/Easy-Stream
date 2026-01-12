package service

import (
	"easy-stream/internal/config"
	"easy-stream/internal/model"
	"easy-stream/internal/repository"
	"easy-stream/internal/zlm"
	"easy-stream/pkg/utils"
)

type StreamService struct {
	streamRepo *repository.StreamRepository
	zlmClient  *zlm.Client
}

func NewStreamService(streamRepo *repository.StreamRepository, zlmCfg config.ZLMediaKitConfig) *StreamService {
	return &StreamService{
		streamRepo: streamRepo,
		zlmClient:  zlm.NewClient(zlmCfg.Host, zlmCfg.Port, zlmCfg.Secret),
	}
}

// Create 创建推流码
func (s *StreamService) Create(req *model.CreateStreamRequest) (*model.Stream, error) {
	stream := &model.Stream{
		StreamKey: utils.GenerateStreamKey(),
		Name:      req.Name,
		DeviceID:  req.DeviceID,
		Status:    model.StreamStatusIdle,
	}

	if err := s.streamRepo.Create(stream); err != nil {
		return nil, err
	}
	return stream, nil
}

// Get 获取推流信息
func (s *StreamService) Get(key string) (*model.Stream, error) {
	return s.streamRepo.GetByKey(key)
}

// List 获取推流列表
func (s *StreamService) List(status string, page, pageSize int) (*model.StreamListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	streams, total, err := s.streamRepo.List(status, offset, pageSize)
	if err != nil {
		return nil, err
	}

	return &model.StreamListResponse{
		Total:   total,
		Streams: streams,
	}, nil
}

// Update 更新推流信息
func (s *StreamService) Update(key string, req *model.UpdateStreamRequest) (*model.Stream, error) {
	stream, err := s.streamRepo.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if stream == nil {
		return nil, ErrStreamNotFound
	}

	if req.Name != "" {
		stream.Name = req.Name
	}
	if req.DeviceID != "" {
		stream.DeviceID = req.DeviceID
	}

	if err := s.streamRepo.Update(stream); err != nil {
		return nil, err
	}
	return stream, nil
}

// Delete 删除推流码
func (s *StreamService) Delete(key string) error {
	return s.streamRepo.Delete(key)
}

// Kick 强制断流
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
	return s.streamRepo.UpdateStatus(key, model.StreamStatusIdle)
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

	stream.Status = model.StreamStatusPushing
	stream.Protocol = req.Schema
	return s.streamRepo.Update(stream)
}

// OnUnpublish 处理推流结束回调
func (s *StreamService) OnUnpublish(req *model.OnUnpublishRequest) error {
	return s.streamRepo.UpdateStatus(req.Stream, model.StreamStatusIdle)
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
