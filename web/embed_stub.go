//go:build !embed_frontend

package web

import (
	"github.com/gin-gonic/gin"
)

// ServeStatic does nothing when frontend is not embedded
func ServeStatic(r *gin.Engine) error {
	// Frontend not embedded, skip static file serving
	return nil
}
