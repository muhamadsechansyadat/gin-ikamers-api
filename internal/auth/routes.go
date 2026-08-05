package auth

import (
	"gin-ikamers-api/internal/platform/ratelimit"
	"gin-ikamers-api/internal/shared/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, limiter *ratelimit.Limiter) {
	authGroup := rg.Group("/auth")
	{
		authGroup.POST("/register", middleware.RateLimit(limiter, ratelimit.PerHour(5), middleware.KeyByIP("register")), h.Register)
		authGroup.POST("/login", middleware.RateLimit(limiter, ratelimit.PerMinute(20), middleware.KeyByIP("login")), h.Login)
		authGroup.POST("/google", middleware.RateLimit(limiter, ratelimit.PerMinute(20), middleware.KeyByIP("login_google")), h.LoginGoogle)
		authGroup.POST("/refresh", middleware.RateLimit(limiter, ratelimit.PerMinute(30), middleware.KeyByIP("refresh")), h.Refresh)
		// auth.POST("/logout", mw, h.Logout)  // kalau perlu middleware
	}
}
