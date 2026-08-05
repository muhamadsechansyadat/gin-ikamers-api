package profile

import (
	"gin-ikamers-api/internal/platform/ratelimit"
	"gin-ikamers-api/internal/shared/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, mw gin.HandlerFunc, limiter *ratelimit.Limiter) {
	profile := rg.Group("/profile")
	profile.Use(mw)
	{
		profile.GET("", h.Show)
		profile.POST("", h.Update)
		profile.PUT("", h.Update)
		profile.POST("/avatar", h.UploadAvatar)
		profile.DELETE("/avatar", h.DeleteAvatar)

		profile.PUT("/password",
			middleware.RateLimit(limiter, ratelimit.PerHour(10), middleware.KeyByUserID("password_change")),
			h.ChangePassword)
		profile.POST("/email",
			middleware.RateLimit(limiter, ratelimit.PerHour(3), middleware.KeyByUserID("email_change")),
			h.RequestEmailChange)
		profile.POST("/email/verify",
			middleware.RateLimit(limiter, ratelimit.PerHour(10), middleware.KeyByUserID("email_verify")),
			h.ConfirmEmailChange)
	}
}
