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

// List 获取推流列表
func (h *StreamHandler) List(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	resp, err := h.streamSvc.List(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Create 创建推流码
func (h *StreamHandler) Create(c *gin.Context) {
	var req model.CreateStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stream, err := h.streamSvc.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, stream)
}

// Get 获取推流详情
func (h *StreamHandler) Get(c *gin.Context) {
	key := c.Param("key")
	stream, err := h.streamSvc.Get(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if stream == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
		return
	}
	c.JSON(http.StatusOK, stream)
}

// Update 更新推流信息
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

// Delete 删除推流码
func (h *StreamHandler) Delete(c *gin.Context) {
	key := c.Param("key")
	if err := h.streamSvc.Delete(key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// Kick 强制断流
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
