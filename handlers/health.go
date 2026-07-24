package handlers

import (
	"context"
	"gin-ikamers-api/internal/response"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func (h *Handler) Home(c *gin.Context) {
	response.Success(
		c,
		http.StatusOK,
		h.Cfg.App.Name+" is running",
		gin.H{
			"env":  h.Cfg.App.Env,
			"time": time.Now().UTC(),
		},
	)
}

func (h *Handler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	dbStatus := "up"
	var dbErr string
	start := time.Now()

	if err := h.DB.PingContext(ctx); err != nil {
		dbStatus = "down"
		dbErr = err.Error()
	}

	payload := gin.H{
		"status": "UP",
		"database": gin.H{
			"status":     dbStatus,
			"latency_ms": time.Since(start).Milliseconds(),
		},
		"time": time.Now().UTC(),
	}

	if dbStatus == "down" {
		payload["status"] = "DOWN"
		payload["database"].(gin.H)["error"] = dbErr
		response.Error(c, http.StatusServiceUnavailable, "Service Unhealthy", payload)
		return
	}

	response.Success(c, http.StatusOK, "Server is healty", payload)
}
