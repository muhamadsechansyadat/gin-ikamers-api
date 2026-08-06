package middleware

import (
	"gin-ikamers-api/internal/shared/response"
	"github.com/gin-gonic/gin"
	"strings"
)

type TokenVerifier func(token string) (userID int64, role string, err error)

func AuthMiddleware(verify TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token == "" {
			response.Unauthorized(c, "Missing authorization header", nil)
			return
		}

		userID, role, err := verify(token)
		if err != nil {
			response.Unauthorized(c, "Invalid token", nil)
			return
		}

		c.Set("user_id", userID)
		c.Set("user_role", role)

		c.Next()
	}
}

func extractBearerToken(c *gin.Context) string {
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
