package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SystemHandler struct{}

func NewSystemHandler() *SystemHandler {
	return &SystemHandler{}
}

// Health 健康检查
func (h *SystemHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// Stats 系统统计
func (h *SystemHandler) Stats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"online_streams": 0,
		"total_streams":  0,
	})
}
