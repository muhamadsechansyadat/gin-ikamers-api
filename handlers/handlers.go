package handlers

import (
	"database/sql"
	"gin-ikamers-api/config"
	"log/slog"
)

type Handler struct {
	DB     *sql.DB
	Logger *slog.Logger
	Cfg    *config.Config
}

func New(db *sql.DB, logger *slog.Logger, cfg *config.Config) *Handler {
	return &Handler{
		DB:     db,
		Logger: logger,
		Cfg:    cfg,
	}
}
