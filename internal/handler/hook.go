package handler

import (
	"context"
	"net/http"
	"time"

	"easy-stream/internal/model"
	"easy-stream/internal/service"
	"easy-stream/internal/storage"
	"easy-stream/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HookHandler struct {
	streamSvc      *service.StreamService
	storageManager *storage.Manager
}

func NewHookHandler(streamSvc *service.StreamService, storageManager *storage.Manager) *HookHandler {
	return &HookHandler{
		streamSvc:      streamSvc,
		storageManager: storageManager,
	}
}

// OnPublish 推流开始回调
func (h *HookHandler) OnPublish(c *gin.Context) {
	var req model.OnPublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.HookResponse{Code: -1, Msg: err.Error()})
		return
	}

	logger.Info("OnPublish hook called",
		zap.String("app", req.App),
		zap.String("stream", req.Stream),
		zap.String("schema", req.Schema))

	if err := h.streamSvc.OnPublish(&req); err != nil {
		// 根据错误类型返回不同的错误信息
		var msg string
		switch err {
		case service.ErrStreamNotFound:
			msg = "stream not found"
		case service.ErrStreamExpired:
			msg = "stream expired"
		default:
			msg = err.Error()
		}
		// 返回 code=-1 会拒绝推流，ZLMediaKit 会断开连接
		c.JSON(http.StatusOK, model.HookResponse{Code: -1, Msg: msg})
		return
	}

	c.JSON(http.StatusOK, model.HookResponse{Code: 0, Msg: "success"})
}

// OnUnpublish 推流结束回调
func (h *HookHandler) OnUnpublish(c *gin.Context) {
	var req model.OnUnpublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.HookResponse{Code: 0, Msg: err.Error()})
		return
	}

	logger.Info("OnUnpublish hook called",
		zap.String("app", req.App),
		zap.String("stream", req.Stream))

	_ = h.streamSvc.OnUnpublish(&req)
	c.JSON(http.StatusOK, model.HookResponse{Code: 0, Msg: "success"})
}

// OnFlowReport 流量统计回调
func (h *HookHandler) OnFlowReport(c *gin.Context) {
	var req model.OnFlowReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.HookResponse{Code: 0, Msg: err.Error()})
		return
	}

	_ = h.streamSvc.OnFlowReport(&req)
	c.JSON(http.StatusOK, model.HookResponse{Code: 0, Msg: "success"})
}

// OnStreamNoneReader 无人观看回调
func (h *HookHandler) OnStreamNoneReader(c *gin.Context) {
	var req model.OnStreamNoneReaderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.HookResponse{Code: 0, Msg: err.Error()})
		return
	}

	// 返回 close: true 会关闭流
	c.JSON(http.StatusOK, gin.H{"code": 0, "close": false})
}

// OnPlay 播放开始回调
func (h *HookHandler) OnPlay(c *gin.Context) {
	var req model.OnPlayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.HookResponse{Code: 0, Msg: err.Error()})
		return
	}

	logger.Info("OnPlay hook called",
		zap.String("app", req.App),
		zap.String("stream", req.Stream),
		zap.String("id", req.ID))

	_ = h.streamSvc.OnPlay(&req)
	c.JSON(http.StatusOK, model.HookResponse{Code: 0, Msg: "success"})
}

// OnPlayerDisconnect 播放器断开回调
func (h *HookHandler) OnPlayerDisconnect(c *gin.Context) {
	var req model.OnPlayerDisconnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.HookResponse{Code: 0, Msg: err.Error()})
		return
	}

	logger.Info("OnPlayerDisconnect hook called",
		zap.String("app", req.App),
		zap.String("stream", req.Stream),
		zap.String("id", req.ID))

	_ = h.streamSvc.OnPlayerDisconnect(&req)
	c.JSON(http.StatusOK, model.HookResponse{Code: 0, Msg: "success"})
}

// OnStreamChanged 流注册/注销回调
func (h *HookHandler) OnStreamChanged(c *gin.Context) {
	var req model.OnStreamChangedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.HookResponse{Code: 0, Msg: err.Error()})
		return
	}

	logger.Info("OnStreamChanged hook called",
		zap.String("app", req.App),
		zap.String("stream", req.Stream),
		zap.Bool("regist", req.Regist))

	// regist=true 表示流注册（推流开始）
	if req.Regist {
		// 流已完全注册，此时开始录制
		_ = h.streamSvc.OnStreamRegistered(&req)
	} else {
		// regist=false 表示推流结束
		// 转换为 OnUnpublish 请求
		unpublishReq := &model.OnUnpublishRequest{
			App:        req.App,
			Stream:     req.Stream,
			Schema:     req.Schema,
			MediaSrvID: req.MediaSrvID,
		}
		_ = h.streamSvc.OnUnpublish(unpublishReq)
	}

	c.JSON(http.StatusOK, model.HookResponse{Code: 0, Msg: "success"})
}

// OnRecordMP4 录制完成回调
func (h *HookHandler) OnRecordMP4(c *gin.Context) {
	var req model.OnRecordMP4Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.HookResponse{Code: 0, Msg: err.Error()})
		return
	}

	// 处理文件路径：去掉 ZLM 的绝对路径前缀
	// ZLM 可能返回类似 /opt/media/bin/www/record/live/stream_xxx/... 的路径
	// 我们只需要保留 live/stream_xxx/... 部分
	filePath := h.stripRecordPathPrefix(req.FilePath)

	// 构建录制文件元数据
	recordFile := &model.RecordFile{
		FileName:  req.FileName,
		FilePath:  filePath,
		FileSize:  req.FileSize,
		Duration:  req.TimeLen,
		StartTime: req.StartTime,
		TimeLen:   req.TimeLen,
		CreatedAt: time.Now(),
		URLs:      make(map[string]string),
	}

	// 记录录制文件到数据库
	if err := h.streamSvc.AddRecordFile(req.Stream, recordFile); err != nil {
		c.JSON(http.StatusOK, model.HookResponse{Code: 0, Msg: err.Error()})
		return
	}

	// 上传到所有启用的存储
	if h.storageManager != nil && h.storageManager.HasStorages() {
		// 本地模式使用原始绝对路径，远程模式使用处理后的相对路径
		uploadPath := req.FilePath
		if !h.streamSvc.IsRecordFileLocal() {
			uploadPath = filePath
		}
		go h.uploadRecordFile(req.Stream, req.FileName, uploadPath, req.FileSize)
	}

	c.JSON(http.StatusOK, model.HookResponse{Code: 0, Msg: "success"})
}

// uploadRecordFile 上传录制文件到存储（支持本地和远程模式）
func (h *HookHandler) uploadRecordFile(streamKey, fileName, filePath string, fileSize int64) {
	ctx := context.Background()
	remotePath := fileName

	var urls map[string]string

	// 判断是本地模式还是远程模式
	// 本地模式：filePath 是本地路径，直接上传
	// 远程模式：filePath 是 ZLM 服务器路径，需要通过 HTTP 下载
	if h.streamSvc.IsRecordFileLocal() {
		// 本地模式：直接从本地文件上传
		urls = h.storageManager.UploadToAll(ctx, filePath, remotePath)
	} else {
		// 远程模式：通过 HTTP 流式下载并上传
		recordFileURL := h.streamSvc.GetRecordFileURL(filePath)
		if recordFileURL == "" {
			logger.Error("failed to get record file URL",
				zap.String("stream", streamKey),
				zap.String("file", fileName))
			return
		}

		// 发起 HTTP 请求下载文件
		resp, err := http.Get(recordFileURL)
		if err != nil {
			logger.Error("failed to download record file",
				zap.String("stream", streamKey),
				zap.String("file", fileName),
				zap.String("url", recordFileURL),
				zap.Error(err))
			return
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		if resp.StatusCode != http.StatusOK {
			logger.Error("failed to download record file: bad status",
				zap.String("stream", streamKey),
				zap.String("file", fileName),
				zap.Int("status", resp.StatusCode))
			return
		}

		// 流式上传到所有存储
		urls = h.storageManager.UploadFromReaderToAll(ctx, resp.Body, remotePath, fileSize)
	}

	// 更新录制文件的存储 URLs
	if err := h.streamSvc.UpdateRecordFileURLs(streamKey, fileName, urls); err != nil {
		logger.Error("failed to update record file URLs",
			zap.String("stream", streamKey),
			zap.String("file", fileName),
			zap.Error(err))
	} else {
		logger.Info("record file uploaded to storages",
			zap.String("stream", streamKey),
			zap.String("file", fileName),
			zap.Any("urls", urls))
	}
}

// stripRecordPathPrefix 去掉 ZLM 录制文件的绝对路径前缀
// 例如：/opt/media/bin/www/record/live/stream_xxx/... -> live/stream_xxx/...
func (h *HookHandler) stripRecordPathPrefix(filePath string) string {
	// 常见的 ZLM 录制路径前缀
	prefixes := []string{
		"/opt/media/bin/www/record/",
		"/opt/media/www/record/",
		"/www/record/",
		"record/",
	}

	for _, prefix := range prefixes {
		if len(filePath) > len(prefix) && filePath[:len(prefix)] == prefix {
			return filePath[len(prefix):]
		}
	}

	// 如果没有匹配的前缀，返回原路径
	return filePath
}
