package handler

import (
	"net/http"
	"strconv"

	"easy-stream/internal/model"
	"easy-stream/internal/service"

	"github.com/gin-gonic/gin"
)

type StreamHandler struct {
	streamSvc *service.StreamService
}

func NewStreamHandler(streamSvc *service.StreamService) *StreamHandler {
	return &StreamHandler{streamSvc: streamSvc}
}

// List 获取推流列表（支持游客和管理员）
func (h *StreamHandler) List(c *gin.Context) {
	status := c.Query("status")
	visibility := c.Query("visibility")
	timeRange := c.Query("time_range")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	// 检查用户是否已登录
	_, isLoggedIn := c.Get("user_id")

	req := &model.StreamListRequest{
		Status:     status,
		Visibility: visibility,
		TimeRange:  timeRange,
		Page:       page,
		PageSize:   pageSize,
	}

	resp, err := h.streamSvc.List(req, isLoggedIn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Create 创建推流码（管理员）
func (h *StreamHandler) Create(c *gin.Context) {
	var req model.CreateStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt64("user_id")
	stream, err := h.streamSvc.Create(&req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, stream)
}

// Get 获取推流详情（支持游客和管理员）
func (h *StreamHandler) Get(c *gin.Context) {
	key := c.Param("key")
	accessToken := c.Query("access_token")

	// 检查用户是否已登录
	_, isLoggedIn := c.Get("user_id")

	stream, err := h.streamSvc.Get(key, isLoggedIn, accessToken)
	if err != nil {
		if err == service.ErrStreamNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
			return
		}
		if err == service.ErrPrivateStream {
			c.JSON(http.StatusForbidden, gin.H{"error": "private stream requires authentication"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stream)
}

// GetByID 通过 ID 获取推流详情（管理员）
func (h *StreamHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	stream, err := h.streamSvc.GetByID(id)
	if err != nil {
		if err == service.ErrStreamNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stream)
}

// VerifyShareCode 验证分享码（游客）
func (h *StreamHandler) VerifyShareCode(c *gin.Context) {
	var req model.VerifyShareCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.streamSvc.VerifyShareCode(req.ShareCode)
	if err != nil {
		switch err {
		case service.ErrInvalidShareCode:
			c.JSON(http.StatusNotFound, gin.H{"error": "invalid share code"})
		case service.ErrStreamEnded:
			c.JSON(http.StatusGone, gin.H{"error": "stream has ended"})
		case service.ErrShareCodeMaxUsesReached:
			c.JSON(http.StatusForbidden, gin.H{"error": "share code max uses reached"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, token)
}

// AddShareCode 为直播添加分享码（管理员）
func (h *StreamHandler) AddShareCode(c *gin.Context) {
	key := c.Param("key")
	var req model.RegenerateShareCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	maxUses := 0
	if req.MaxUses != nil {
		maxUses = *req.MaxUses
	}

	stream, err := h.streamSvc.AddShareCode(key, maxUses)
	if err != nil {
		switch err {
		case service.ErrStreamNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
		case service.ErrNotPrivateStream:
			c.JSON(http.StatusBadRequest, gin.H{"error": "only private streams support sharing"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, stream)
}

// RegenerateShareCode 重新生成分享码（管理员）
func (h *StreamHandler) RegenerateShareCode(c *gin.Context) {
	key := c.Param("key")
	var req model.RegenerateShareCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stream, err := h.streamSvc.RegenerateShareCode(key, &req)
	if err != nil {
		switch err {
		case service.ErrStreamNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
		case service.ErrNotPrivateStream:
			c.JSON(http.StatusBadRequest, gin.H{"error": "only private streams support sharing"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, stream)
}

// UpdateShareCodeMaxUses 更新分享码最大使用次数（管理员）
func (h *StreamHandler) UpdateShareCodeMaxUses(c *gin.Context) {
	key := c.Param("key")
	var req struct {
		MaxUses int `json:"max_uses"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stream, err := h.streamSvc.UpdateShareCodeMaxUses(key, req.MaxUses)
	if err != nil {
		switch err {
		case service.ErrStreamNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
		case service.ErrNoShareCode:
			c.JSON(http.StatusBadRequest, gin.H{"error": "stream has no share code"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, stream)
}

// DeleteShareCode 删除分享码（管理员）
func (h *StreamHandler) DeleteShareCode(c *gin.Context) {
	key := c.Param("key")
	if err := h.streamSvc.DeleteShareCode(key); err != nil {
		if err == service.ErrStreamNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "share code deleted"})
}

// Update 更新推流信息（管理员）
func (h *StreamHandler) Update(c *gin.Context) {
	key := c.Param("key")
	var req model.UpdateStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stream, err := h.streamSvc.Update(key, &req)
	if err != nil {
		if err == service.ErrStreamNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stream)
}

// Delete 删除推流码（管理员）
func (h *StreamHandler) Delete(c *gin.Context) {
	key := c.Param("key")
	if err := h.streamSvc.Delete(key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// Kick 强制断流（管理员）
func (h *StreamHandler) Kick(c *gin.Context) {
	key := c.Param("key")
	if err := h.streamSvc.Kick(key); err != nil {
		if err == service.ErrStreamNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "kicked"})
}
