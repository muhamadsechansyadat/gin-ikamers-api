package auth

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, h *Handler /*, mw gin.HandlerFunc */) {
	auth := rg.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/google", h.LoginGoogle)
		auth.POST("/refresh", h.Refresh)
		// auth.POST("/logout", mw, h.Logout)  // kalau perlu middleware
	}
}
