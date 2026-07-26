package routes

import (
	"gin-ikamers-api/handlers"
	"gin-ikamers-api/internal/auth"
	"gin-ikamers-api/internal/shared/middleware"
	"gin-ikamers-api/internal/user"
	"github.com/gin-gonic/gin"
)

func RegisterRouter(router *gin.Engine, h *handlers.Handler) {
	authMW := middleware.AuthMiddleware(h.AuthService)

	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.GET("/", h.Home)
			v1.GET("/health", h.Health)

			auth.RegisterRoutes(v1, h.Auth)
			user.RegisterRoutes(v1, h.User, authMW)
		}
	}
}
