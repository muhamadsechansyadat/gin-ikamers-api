package routes

import (
	"gin-ikamers-api/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRouter(router *gin.Engine, h *handlers.Handler) {
	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.GET("/", h.Home)
			v1.GET("/health", h.Health)

			registerAuthRoutes(v1, h.Auth)
		}
	}
}
