package profile

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, mw gin.HandlerFunc) {
	profile := rg.Group("/profile")
	profile.Use(mw)
	{
		profile.GET("", h.Show)
		profile.POST("", h.Update)
		profile.PUT("", h.Update)
		profile.POST("/avatar", h.UploadAvatar)
		profile.DELETE("/avatar", h.DeleteAvatar)

		profile.PUT("/password", h.ChangePassword)
		profile.POST("/email", h.RequestEmailChange)
		profile.POST("/email/verify", h.ConfirmEmailChange)
	}
}
