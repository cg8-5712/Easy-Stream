package handler

import (
	"net/http"

	"easy-stream/internal/service"

	"github.com/gin-gonic/gin"
)

type SystemHandler struct {
	systemSvc *service.SystemService
}

func NewSystemHandler(systemSvc *service.SystemService) *SystemHandler {
	return &SystemHandler{
		systemSvc: systemSvc,
	}
}

// Health 健康检查
// @Summary 健康检查
// @Description 返回系统健康状态，包括数据库、Redis、ZLMediaKit 和网络连接状态
// @Tags 系统
// @Produce json
// @Success 200 {object} service.HealthStatus
// @Router /system/health [get]
func (h *SystemHandler) Health(c *gin.Context) {
	health := h.systemSvc.CheckHealth()

	// 根据健康状态返回不同的 HTTP 状态码
	statusCode := http.StatusOK
	if health.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, health)
}

// Stats 系统统计
func (h *SystemHandler) Stats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"online_streams": 0,
		"total_streams":  0,
	})
}
