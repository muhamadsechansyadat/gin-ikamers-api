package app

import (
	"gin-ikamers-api/internal/auth"
	"gin-ikamers-api/internal/profile"
	"gin-ikamers-api/internal/shared/middleware"
	"gin-ikamers-api/internal/user"
	"github.com/gin-gonic/gin"
	"net/http"
)

// RegisterRoutes registers all HTTP routes for the application.
func (a *App) RegisterRoutes(router *gin.Engine) {
	authMW := middleware.AuthMiddleware(a.AuthService)

	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.GET("/", home)
			v1.GET("/health", health)

			auth.RegisterRoutes(v1, a.AuthHandler)
			user.RegisterRoutes(v1, a.UserHandler, authMW)
			profile.RegisterRoutes(v1, a.ProfileHandler, authMW)
		}
	}
}

func home(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "welcome to ikamers-api"})
}

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
