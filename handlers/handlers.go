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

	Auth *auth.Handler
	User *user.Handler

	AuthService *auth.Service
}

func New(db *sql.DB, logger *slog.Logger, cfg *config.Config) *Handler {
	// Repositories
	userRepo := user.NewPostgresRepository(db)
	authRepo := auth.NewPostgresRepository(db)

	// Services
	tokenService := auth.NewTokenService(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
		cfg.JWT.Issuer,
	)
	authService := auth.NewService(userRepo, authRepo, tokenService, cfg.OAuth.GoogleClientID, 8)
	userService := user.NewService(userRepo)

	return &Handler{
		DB:          db,
		Logger:      logger,
		Cfg:         cfg,
		Auth:        auth.NewHandler(authService),
		User:        user.NewHandler(userService),
		AuthService: authService,
	}
}
