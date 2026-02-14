package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"easy-stream/internal/model"
	"easy-stream/internal/service"
	"easy-stream/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RecordHandler struct {
	recordSvc *service.RecordService
}

func NewRecordHandler(recordSvc *service.RecordService) *RecordHandler {
	return &RecordHandler{recordSvc: recordSvc}
}

// ListRecords 获取所有录制文件列表（分页）
func (h *RecordHandler) ListRecords(c *gin.Context) {
	// 验证分页参数
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	streams, total, err := h.recordSvc.GetAllRecords(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":   total,
		"streams": streams,
	})
}

// GetRecordsByStreamKey 根据stream_key获取录制文件
func (h *RecordHandler) GetRecordsByStreamKey(c *gin.Context) {
	streamKey := c.Param("key")

	stream, err := h.recordSvc.GetRecordsByStreamKey(streamKey)
	if err != nil {
		if err == service.ErrStreamNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stream_id":     stream.ID,
		"stream_key":    stream.StreamKey,
		"stream_name":   stream.Name,
		"record_files":  stream.RecordFiles,
		"record_status": stream.RecordStatus,
	})
}

// DeleteRecordFile 删除指定的录制文件
func (h *RecordHandler) DeleteRecordFile(c *gin.Context) {
	streamKey := c.Param("key")
	filepath := c.Param("filepath")

	// 去掉前导的 /
	filename := strings.TrimPrefix(filepath, "/")

	if err := h.recordSvc.DeleteRecordFile(streamKey, filename); err != nil {
		if err == service.ErrStreamNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "record file deleted successfully"})
}

// DownloadFile 下载录制文件（通过Main Server代理）
func (h *RecordHandler) DownloadFile(c *gin.Context) {
	streamKey := c.Param("key")
	filepath := c.Param("filepath")

	// 去掉前导的 /
	filename := strings.TrimPrefix(filepath, "/")

	// 获取录制文件信息
	stream, err := h.recordSvc.GetRecordsByStreamKey(streamKey)
	if err != nil {
		if err == service.ErrStreamNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 查找指定的录制文件
	var targetFile *model.RecordFile
	for i := range stream.RecordFiles {
		if stream.RecordFiles[i].FileName == filename {
			targetFile = &stream.RecordFiles[i]
			break
		}
	}

	if targetFile == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	// 本地模式：直接返回文件
	if h.recordSvc.IsLocalMode() {
		fullPath := h.recordSvc.GetFullPath(targetFile.FilePath)
		if fullPath == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "file path not available"})
			return
		}

		// 设置下载响应头
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Header("Content-Type", "application/octet-stream")

		c.File(fullPath)
		return
	}

	// 远程模式：通过 Main Server 代理下载
	// 优先使用对象存储URL（如果有）
	var downloadURL string

	logger.Info("record download debug",
		zap.String("stream_key", streamKey),
		zap.String("filename", filename),
		zap.Any("target_file_urls", targetFile.URLs),
		zap.String("target_file_path", targetFile.FilePath))

	if targetFile.URLs != nil {
		// 优先级：对象存储 > 原始文件
		for storageName, url := range targetFile.URLs {
			if storageName != "download" && url != "" {
				downloadURL = url
				logger.Info("using object storage URL", zap.String("storage", storageName), zap.String("url", url))
				break
			}
		}

		// 如果没有对象存储URL，使用原始文件URL
		if downloadURL == "" {
			if url, ok := targetFile.URLs["download"]; ok && url != "" {
				downloadURL = url
				logger.Info("using download URL", zap.String("url", url))
			}
		}
	}

	logger.Info("final download URL", zap.String("url", downloadURL))

	if downloadURL == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not available"})
		return
	}

	// 代理下载：从远程服务器获取文件并流式传输给客户端
	logger.Info("fetching file from remote", zap.String("url", downloadURL))
	resp, err := http.Get(downloadURL)
	if err != nil {
		logger.Error("failed to fetch from remote", zap.Error(err), zap.String("url", downloadURL))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch file from remote server"})
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	logger.Info("remote response", zap.Int("status", resp.StatusCode), zap.Int64("content_length", resp.ContentLength))
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("remote server returned status %d", resp.StatusCode)})
		return
	}

	// 设置下载响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/octet-stream")
	if resp.ContentLength > 0 {
		c.Header("Content-Length", fmt.Sprintf("%d", resp.ContentLength))
	}

	// 流式传输文件内容
	c.DataFromReader(http.StatusOK, resp.ContentLength, "application/octet-stream", resp.Body, nil)
}

// PlayFile 播放录制文件（支持 HTTP Range 请求）
func (h *RecordHandler) PlayFile(c *gin.Context) {
	streamKey := c.Param("key")
	filepath := c.Param("filepath")

	// 去掉前导的 /
	filename := strings.TrimPrefix(filepath, "/")

	// 获取录制文件信息
	stream, err := h.recordSvc.GetRecordsByStreamKey(streamKey)
	if err != nil {
		if err == service.ErrStreamNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 查找指定的录制文件
	var targetFile *model.RecordFile
	for i := range stream.RecordFiles {
		if stream.RecordFiles[i].FileName == filename {
			targetFile = &stream.RecordFiles[i]
			break
		}
	}

	if targetFile == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	// 获取文件 URL
	var fileURL string
	if h.recordSvc.IsLocalMode() {
		// 本地模式：使用本地文件路径
		fullPath := h.recordSvc.GetFullPath(targetFile.FilePath)
		if fullPath == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "file path not available"})
			return
		}
		// 本地模式直接使用 ServeFile，它自动支持 Range 请求
		c.Header("Content-Type", "video/mp4")
		c.Header("Accept-Ranges", "bytes")
		http.ServeFile(c.Writer, c.Request, fullPath)
		return
	}

	// 远程模式：获取远程 URL
	if targetFile.URLs != nil {
		// 优先级：对象存储 > 原始文件
		for storageName, url := range targetFile.URLs {
			if storageName != "download" && url != "" {
				fileURL = url
				break
			}
		}

		// 如果没有对象存储URL，使用原始文件URL
		if fileURL == "" {
			if url, ok := targetFile.URLs["download"]; ok && url != "" {
				fileURL = url
			}
		}
	}

	if fileURL == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not available"})
		return
	}

	// 远程模式：代理 Range 请求
	h.proxyRangeRequest(c, fileURL, filename)
}

// proxyRangeRequest 代理 HTTP Range 请求
func (h *RecordHandler) proxyRangeRequest(c *gin.Context, remoteURL, filename string) {
	// 创建请求
	req, err := http.NewRequest("GET", remoteURL, nil)
	if err != nil {
		logger.Error("failed to create request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

	// 转发 Range header（如果有）
	rangeHeader := c.GetHeader("Range")
	if rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
		logger.Info("proxying range request", zap.String("range", rangeHeader), zap.String("url", remoteURL))
	}

	// 发起请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("failed to fetch from remote", zap.Error(err), zap.String("url", remoteURL))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch file from remote server"})
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		logger.Error("remote server error", zap.Int("status", resp.StatusCode), zap.String("url", remoteURL))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("remote server returned status %d", resp.StatusCode)})
		return
	}

	// 设置响应头
	c.Header("Content-Type", "video/mp4")
	c.Header("Accept-Ranges", "bytes")

	// 转发 Content-Range header（如果是 206 响应）
	if resp.StatusCode == http.StatusPartialContent {
		if contentRange := resp.Header.Get("Content-Range"); contentRange != "" {
			c.Header("Content-Range", contentRange)
		}
	}

	// 设置 Content-Length
	if resp.ContentLength > 0 {
		c.Header("Content-Length", strconv.FormatInt(resp.ContentLength, 10))
	}

	// 流式传输文件内容
	c.Status(resp.StatusCode)
	_, _ = io.Copy(c.Writer, resp.Body)
}
