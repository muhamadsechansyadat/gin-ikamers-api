package middleware

import (
	"gin-ikamers-api/internal/shared/response"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"runtime/debug"
)

func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(
					"Panic, recovered",
					"error", err,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
					"stack", string(debug.Stack()),
				)

				response.Error(
					c,
					http.StatusInternalServerError,
					"Internal server error",
					nil,
				)

				c.Abort()
			}
		}()
		c.Next()
	}
}
