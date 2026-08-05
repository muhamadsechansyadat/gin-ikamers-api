package app

import (
	"gin-ikamers-api/internal/auth"
	"gin-ikamers-api/internal/platform/health"
	"gin-ikamers-api/internal/profile"
	"gin-ikamers-api/internal/shared/middleware"
	"gin-ikamers-api/internal/shared/response"
	"gin-ikamers-api/internal/user"
	"github.com/gin-gonic/gin"
	"net/http"
)

// RegisterRoutes registers all HTTP routes for the application.
func (a *App) RegisterRoutes(router *gin.Engine) {
	authMW := middleware.AuthMiddleware(func(token string) (int64, string, error) {
		claims, err := a.AuthService.VerifyAccessToken(token)
		if err != nil {
			return 0, "", err
		}
		return claims.UserID, claims.Role, nil
	})

	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.GET("/", home)
			v1.GET("/health", a.healthHandler)

			auth.RegisterRoutes(v1, a.AuthHandler, a.Limiter)
			user.RegisterRoutes(v1, a.UserHandler, authMW)
			profile.RegisterRoutes(v1, a.ProfileHandler, authMW, a.Limiter)
		}
	}
}

func home(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "welcome to ikamers-api"})
}

func (a *App) healthHandler(c *gin.Context) {
	result := a.Health.Run(c.Request.Context())
	if result.Status == health.StatusFail {
		response.ServiceUnavailable(c, "One or more dependencies are unhealthy", result.Checks)
		return
	}
	response.OK(c, "All dependencies are healthy", result.Checks)
}
