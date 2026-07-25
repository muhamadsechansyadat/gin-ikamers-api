package logger

import (
	"gin-ikamers-api/internal/config"
	"log/slog"
	"os"
	"strings"
)

func New(cfg config.LogConfig) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: parseLevel(strings.ToLower(cfg.Level)),
	}

	var handler slog.Handler
	if strings.EqualFold(cfg.Format, "json") {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":

		return slog.LevelDebug
	case "warn", "warning":

		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
