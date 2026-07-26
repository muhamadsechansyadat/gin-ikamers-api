package app

import (
	"database/sql"
	"gin-ikamers-api/internal/auth"
	"gin-ikamers-api/internal/config"
	"gin-ikamers-api/internal/user"
	"log/slog"
)

// App holds all wired dependencies for the application.
type App struct {
	Config *config.Config
	Logger *slog.Logger
	DB     *sql.DB

	// Handlers
	AuthHandler *auth.Handler
	UserHandler *user.Handler

	// Services (exposed for middleware)
	AuthService *auth.Service
}

// New wires up all dependencies and returns a ready-to-use App.
func New(db *sql.DB, logger *slog.Logger, cfg *config.Config) *App {
	// ── Repositories ────────────────────────────────
	userRepo := user.NewPostgresRepository(db)
	authRepo := auth.NewPostgresRepository(db)

	// ── Services ────────────────────────────────────
	tokenService := auth.NewTokenService(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
		cfg.JWT.Issuer,
	)
	authService := auth.NewService(userRepo, authRepo, tokenService, cfg.OAuth.GoogleClientID, 8)
	userService := user.NewService(userRepo)

	// ── Handlers ────────────────────────────────────
	authHandler := auth.NewHandler(authService)
	userHandler := user.NewHandler(userService)

	return &App{
		Config:      cfg,
		Logger:      logger,
		DB:          db,
		AuthHandler: authHandler,
		UserHandler: userHandler,
		AuthService: authService,
	}
}
