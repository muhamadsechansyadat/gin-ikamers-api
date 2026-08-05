package user

import (
	"gin-ikamers-api/internal/shared/response"
	"github.com/gin-gonic/gin"
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
		response.Unauthorized(c, "Unauthorized", nil)
		return
	}

	u, err := h.service.GetByID(c.Request.Context(), userID.(int64))
	if err != nil || u == nil {
		response.NotFound(c, "User not found", nil)
		return
	}

	response.OK(c, "User profile fetched", ToResponse(u))
}
