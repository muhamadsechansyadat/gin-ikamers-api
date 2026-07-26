package user

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, mw gin.HandlerFunc) {
	users := rg.Group("/users")
	users.Use(mw)
	{
		users.GET("/me", h.Me)
	}
}
