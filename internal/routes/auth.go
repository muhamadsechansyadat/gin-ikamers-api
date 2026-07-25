package routes

import (
	"gin-ikamers-api/handlers"
	"github.com/gin-gonic/gin"
)

func registerAuthRoutes(rg *gin.RouterGroup, h *handlers.AuthHandler) {
	auth := rg.Group("/auth")
	{
		// Public
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/google", h.LoginGoogle)
		auth.POST("/refresh", h.Refresh)

		// Protected (butuh middleware — bisa dipasang di sini)
		// auth.POST("/logout", middleware.AuthMiddleware(authService), h.Logout)
	}
}
