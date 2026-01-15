package handler

import (
	"net/http"

	"easy-stream/internal/model"
	"easy-stream/internal/service"
	"easy-stream/internal/storage"

	"github.com/gin-gonic/gin"
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

	if err := h.streamSvc.OnPublish(&req); err != nil {
		// 根据错误类型返回不同的错误信息
		msg := "unknown error"
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

	h.streamSvc.OnUnpublish(&req)
	c.JSON(http.StatusOK, model.HookResponse{Code: 0, Msg: "success"})
}

// OnFlowReport 流量统计回调
func (h *HookHandler) OnFlowReport(c *gin.Context) {
	var req model.OnFlowReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.HookResponse{Code: 0, Msg: err.Error()})
		return
	}

	h.streamSvc.OnFlowReport(&req)
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

	h.streamSvc.OnPlay(&req)
	c.JSON(http.StatusOK, model.HookResponse{Code: 0, Msg: "success"})
}

// OnPlayerDisconnect 播放器断开回调
func (h *HookHandler) OnPlayerDisconnect(c *gin.Context) {
	var req model.OnPlayerDisconnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.HookResponse{Code: 0, Msg: err.Error()})
		return
	}

	h.streamSvc.OnPlayerDisconnect(&req)
	c.JSON(http.StatusOK, model.HookResponse{Code: 0, Msg: "success"})
}

// OnRecordMP4 录制完成回调
func (h *HookHandler) OnRecordMP4(c *gin.Context) {
	var req model.OnRecordMP4Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.HookResponse{Code: 0, Msg: err.Error()})
		return
	}

	// 记录录制文件到数据库
	if err := h.streamSvc.AddRecordFile(req.Stream, req.FilePath); err != nil {
		c.JSON(http.StatusOK, model.HookResponse{Code: 0, Msg: err.Error()})
		return
	}

	// 上传到所有启用的存储
	if h.storageManager != nil && h.storageManager.HasStorages() {
		go func() {
			remotePath := req.FileName
			h.storageManager.UploadToAll(c.Request.Context(), req.FilePath, remotePath)
		}()
	}

	c.JSON(http.StatusOK, model.HookResponse{Code: 0, Msg: "success"})
}
