package middleware

import (
	"fmt"
	"gin-ikamers-api/internal/platform/ratelimit"
	"gin-ikamers-api/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
	"strconv"
)

func RateLimit(l *ratelimit.Limiter, limit redis_rate.Limit, keyFn func(ctx *gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFn(c)
		if key == "" {
			c.Next()
			return
		}

		res, err := l.Allow(c.Request.Context(), key, limit)
		if err != nil {
			fmt.Println("[RATELIMIT ERROR]", "key:", key, "err:", err)
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(res.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))
		c.Header("X-RateLimit-Retry-After", strconv.FormatFloat(res.RetryAfter.Seconds(), 'f', 1, 64))

		if !res.Allowed {
			response.TooManyRequests(c,
				"Too many requests. Please try again later.",
				gin.H{"retry_after_seconds": res.RetryAfter.Seconds()})
			return
		}
		c.Next()
	}
}

func KeyByIP(prefix string) func(*gin.Context) string {
	return func(c *gin.Context) string {
		return prefix + ":ip:" + c.ClientIP()
	}
}

func KeyByUserID(prefix string) func(*gin.Context) string {
	return func(c *gin.Context) string {
		uid, exists := c.Get("user_id")
		if !exists {
			return ""
		}
		return prefix + ":user:" + strconv.FormatInt(uid.(int64), 10)
	}
}
