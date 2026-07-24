package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func NotFoundHandler(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"message": "Route not found",
		"path":    c.Request.URL.Path,
	})

	c.Abort()
}

func MethodNotAllowedHandler(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, gin.H{
		"success": false,
		"message": "Method not allowed",
		"method":  c.Request.Method,
	})

	c.Abort()
}
