//go:build embed_frontend

package web

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed all:dist
var frontendFS embed.FS

// ServeStatic serves the embedded frontend files
func ServeStatic(r *gin.Engine) error {
	// Get the embedded filesystem
	sub, err := fs.Sub(frontendFS, "dist")
	if err != nil {
		return err
	}

	// Serve static files with fallback to index.html for SPA
	r.NoRoute(func(c *gin.Context) {
		urlPath := c.Request.URL.Path

		// Skip API routes
		if strings.HasPrefix(urlPath, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		// Clean the path
		filePath := strings.TrimPrefix(urlPath, "/")
		if filePath == "" {
			filePath = "index.html"
		}

		// Try to open the file
		file, err := sub.Open(filePath)
		if err != nil {
			// File not found, serve index.html for SPA routing
			serveFile(c, sub, "index.html")
			return
		}
		file.Close()

		// File exists, serve it
		serveFile(c, sub, filePath)
	})

	return nil
}

func serveFile(c *gin.Context, fsys fs.FS, filePath string) {
	file, err := fsys.Open(filePath)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Set content type based on extension
	ext := path.Ext(filePath)
	contentType := getContentType(ext)
	c.Header("Content-Type", contentType)

	// Read and write the file
	data, err := io.ReadAll(file)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Data(http.StatusOK, contentType, data)
	_ = stat // avoid unused variable
}

func getContentType(ext string) string {
	switch ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	default:
		return "application/octet-stream"
	}
}
