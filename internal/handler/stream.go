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
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	// 获取用户角色（如果已登录）
	userRole, exists := c.Get("role")
	if !exists {
		userRole = "" // 游客
	}

	resp, err := h.streamSvc.List(status, visibility, userRole.(string), page, pageSize)
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
		if err == service.ErrPrivateStream {
			c.JSON(http.StatusBadRequest, gin.H{"error": "private stream requires password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, stream)
}

// Get 获取推流详情（支持游客和管理员）
func (h *StreamHandler) Get(c *gin.Context) {
	key := c.Param("key")
	accessToken := c.Query("access_token")

	// 获取用户角色
	userRole, exists := c.Get("role")
	if !exists {
		userRole = "" // 游客
	}

	stream, err := h.streamSvc.Get(key, userRole.(string), accessToken)
	if err != nil {
		if err == service.ErrStreamNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
			return
		}
		if err == service.ErrPrivateStream {
			c.JSON(http.StatusForbidden, gin.H{"error": "private stream requires password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stream)
}

// VerifyPassword 验证私有直播密码（游客）
func (h *StreamHandler) VerifyPassword(c *gin.Context) {
	key := c.Param("key")
	var req model.VerifyStreamPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.streamSvc.VerifyPassword(key, req.Password)
	if err != nil {
		if err == service.ErrStreamNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
			return
		}
		if err == service.ErrInvalidPassword {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, token)
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
