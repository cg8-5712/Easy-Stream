package handler

import (
	"io"
	"net/http"
	"strconv"

	"easy-stream/internal/model"
	"easy-stream/internal/service"

	"github.com/gin-gonic/gin"
)

// WebRTCPlayRequest WebRTC 播放请求
type WebRTCPlayRequest struct {
	SDP string `json:"sdp"`
}

// WebRTCPlayResponse WebRTC 播放响应
type WebRTCPlayResponse struct {
	Code int    `json:"code"`
	SDP  string `json:"sdp"`
}

// WebRTCPlay WebRTC 播放代理（游客和管理员都可以使用）
// 前端通过 streamId 请求，后端使用 streamKey 与 ZLM 交互
func (h *StreamHandler) WebRTCPlay(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// 读取 SDP offer
	var req WebRTCPlayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken := c.Query("access_token")

	// 检查用户是否已登录
	_, isLoggedIn := c.Get("user_id")

	// 获取流信息
	stream, err := h.streamSvc.GetByID(id)
	if err != nil {
		if err == service.ErrStreamNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 游客访问：公开直播可以直接看，私有直播需要 access_token
	if !isLoggedIn {
		if stream.Visibility == model.StreamVisibilityPrivate {
			if accessToken == "" {
				c.JSON(http.StatusForbidden, gin.H{"error": "private stream requires access token"})
				return
			}
			// 验证 access_token
			valid, err := h.streamSvc.VerifyAccessToken(stream.StreamKey, accessToken)
			if err != nil || !valid {
				c.JSON(http.StatusForbidden, gin.H{"error": "invalid access token"})
				return
			}
		}
	}

	// 调用 ZLM WebRTC 播放接口
	resp, err := h.streamSvc.WebRTCPlay(stream.StreamKey, req.SDP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetWebRTCSDP 获取 WebRTC 播放 SDP（简化版，直接返回 ZLM 的响应）
func (h *StreamHandler) GetWebRTCSDP(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	accessToken := c.Query("access_token")

	// 检查用户是否已登录
	_, isLoggedIn := c.Get("user_id")

	// 获取流信息
	stream, err := h.streamSvc.GetByID(id)
	if err != nil {
		if err == service.ErrStreamNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 游客访问：公开直播可以直接看，私有直播需要 access_token
	if !isLoggedIn {
		if stream.Visibility == model.StreamVisibilityPrivate {
			if accessToken == "" {
				c.JSON(http.StatusForbidden, gin.H{"error": "private stream requires access token"})
				return
			}
			// 验证 access_token
			valid, err := h.streamSvc.VerifyAccessToken(stream.StreamKey, accessToken)
			if err != nil || !valid {
				c.JSON(http.StatusForbidden, gin.H{"error": "invalid access token"})
				return
			}
		}
	}

	// 读取请求体中的 SDP offer
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	// 调用 ZLM WebRTC 播放接口
	resp, err := h.streamSvc.WebRTCPlay(stream.StreamKey, string(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 如果 ZLM 返回纯 SDP 文本，直接返回
	if resp.Code == 0 && resp.SDP != "" {
		c.Header("Content-Type", "application/sdp")
		c.String(http.StatusOK, resp.SDP)
		return
	}

	c.JSON(http.StatusOK, resp)
}
