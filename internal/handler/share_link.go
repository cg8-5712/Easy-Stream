package handler

import (
	"net/http"
	"strconv"

	"easy-stream/internal/model"
	"easy-stream/internal/service"

	"github.com/gin-gonic/gin"
)

type ShareLinkHandler struct {
	shareLinkSvc *service.ShareLinkService
}

func NewShareLinkHandler(shareLinkSvc *service.ShareLinkService) *ShareLinkHandler {
	return &ShareLinkHandler{shareLinkSvc: shareLinkSvc}
}

// Create 创建分享链接（管理员）
func (h *ShareLinkHandler) Create(c *gin.Context) {
	key := c.Param("key")
	var req model.CreateShareLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt64("user_id")
	resp, err := h.shareLinkSvc.Create(key, &req, userID)
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
	c.JSON(http.StatusCreated, resp)
}

// List 获取直播的所有分享链接（管理员）
func (h *ShareLinkHandler) List(c *gin.Context) {
	key := c.Param("key")
	resp, err := h.shareLinkSvc.List(key)
	if err != nil {
		if err == service.ErrStreamNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Verify 验证分享链接（游客）
func (h *ShareLinkHandler) Verify(c *gin.Context) {
	token := c.Param("token")
	resp, err := h.shareLinkSvc.VerifyToken(token)
	if err != nil {
		switch err {
		case service.ErrInvalidShareLink:
			c.JSON(http.StatusNotFound, gin.H{"error": "invalid share link"})
		case service.ErrStreamNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
		case service.ErrStreamEnded:
			c.JSON(http.StatusGone, gin.H{"error": "stream has ended"})
		case service.ErrShareLinkMaxUsesReached:
			c.JSON(http.StatusForbidden, gin.H{"error": "share link max uses reached"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateMaxUses 更新分享链接最大使用次数（管理员）
func (h *ShareLinkHandler) UpdateMaxUses(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		MaxUses int `json:"max_uses"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	link, err := h.shareLinkSvc.UpdateMaxUses(id, req.MaxUses)
	if err != nil {
		if err == service.ErrShareLinkNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "share link not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, link)
}

// Delete 删除分享链接（管理员）
func (h *ShareLinkHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.shareLinkSvc.Delete(id); err != nil {
		if err == service.ErrShareLinkNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "share link not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "share link deleted"})
}
