package auth

import (
	"gin-ikamers-api/internal/shared/response"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleBindError(c, err)
		return
	}

	defaultRoleID := int64(8)

	u, err := h.service.RegisterWithPassword(c.Request.Context(), req.Email, req.Password, defaultRoleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "registration successful",
		"user": gin.H{
			"id":    u.ID,
			"email": u.Email,
		},
	})
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleBindError(c, err)
		return
	}

	ua := c.Request.UserAgent()
	ip := c.ClientIP()

	tokens, err := h.service.LoginWithPassword(c.Request.Context(), req.Email, req.Password, ua, ip)
	if err != nil {
		switch err {
		case ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		case ErrUserInactive:
			c.JSON(http.StatusForbidden, gin.H{"error": "user account is inactive"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "login successfully",
		"data":    tokens,
	})
}

func (h *Handler) LoginGoogle(c *gin.Context) {
	var req GoogleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ua := c.Request.UserAgent()
	ip := c.ClientIP()

	tokens, err := h.service.LoginWithGoogle(c.Request.Context(), req.IdToken, ua, ip)
	if err != nil {
		switch err {
		case ErrInvalidGoogleToken:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid google id token"})
		case ErrUserInactive:
			c.JSON(http.StatusForbidden, gin.H{"error": "user account is inactive"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "login successful",
		"data":    tokens,
	})
}

func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := h.service.RefreshAccessToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "token refreshed",
		"data":    tokens,
	})
}

func (h *Handler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.service.Logout(c.Request.Context(), userID.(int64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "logout successful",
	})
}

func GetExtractToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if header == "" {
		return ""
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}
