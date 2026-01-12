package handler

import (
	"net/http"

	"easy-stream/internal/model"
	"easy-stream/internal/service"

	"github.com/gin-gonic/gin"
)

type HookHandler struct {
	streamSvc *service.StreamService
}

func NewHookHandler(streamSvc *service.StreamService) *HookHandler {
	return &HookHandler{streamSvc: streamSvc}
}

// OnPublish 推流开始回调
func (h *HookHandler) OnPublish(c *gin.Context) {
	var req model.OnPublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.HookResponse{Code: -1, Msg: err.Error()})
		return
	}

	if err := h.streamSvc.OnPublish(&req); err != nil {
		// 返回非 0 code 会拒绝推流
		c.JSON(http.StatusOK, model.HookResponse{Code: -1, Msg: err.Error()})
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
