package handlers

import (
	"database/sql"
	"gin-ikamers-api/internal/auth"
	"gin-ikamers-api/internal/config"
	"gin-ikamers-api/internal/user"
	"log/slog"
)

type Handler struct {
	DB     *sql.DB
	Logger *slog.Logger
	Cfg    *config.Config
	Auth   *auth.Handler
}

func New(db *sql.DB, logger *slog.Logger, cfg *config.Config) *Handler {
	userRepo := user.NewPostgresRepository(db)

	tokenService := auth.NewTokenService(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
		cfg.JWT.Issuer,
	)

	authService := auth.NewService(userRepo, tokenService, cfg.OAuth.GoogleClientID, 8)

	return &Handler{
		DB:     db,
		Logger: logger,
		Cfg:    cfg,
		Auth:   auth.NewHandler(authService),
	}
}
