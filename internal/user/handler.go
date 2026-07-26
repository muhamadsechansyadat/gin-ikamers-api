package user

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	u, err := h.service.GetByID(c.Request.Context(), userID.(int64))
	if err != nil || u == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                u.ID,
		"uuid":              u.UUID,
		"email":             u.Email,
		"role":              u.RoleName,
		"is_active":         u.IsActive,
		"is_email_verified": u.IsEmailVerified,
	})
}
