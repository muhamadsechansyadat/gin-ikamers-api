package app

import (
	"database/sql"
	"gin-ikamers-api/internal/auth"
	"gin-ikamers-api/internal/config"
	"gin-ikamers-api/internal/platform/mailer"
	"gin-ikamers-api/internal/platform/ratelimit"
	"gin-ikamers-api/internal/profile"
	"gin-ikamers-api/internal/shared/storage"
	"gin-ikamers-api/internal/user"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"log/slog"
)

// App holds all wired dependencies for the application.
type App struct {
	Config *config.Config
	Logger *slog.Logger
	DB     *sql.DB
	Gorm   *gorm.DB
	Redis  *redis.Client

	// Handlers
	AuthHandler    *auth.Handler
	UserHandler    *user.Handler
	ProfileHandler *profile.Handler

	// Services (exposed for middleware)
	AuthService *auth.Service
	Limiter     *ratelimit.Limiter
}

// New wires up all dependencies and returns a ready-to-use App.
func New(db *sql.DB, gdb *gorm.DB, rdb *redis.Client, logger *slog.Logger, cfg *config.Config) *App {
	supabaseStorage := storage.NewSupabaseClient(
		cfg.Storage.SupabaseURL,
		cfg.Storage.SupabaseServiceKey,
		cfg.Storage.SupabaseStorageBucket,
	)

	smtpMailer := mailer.NewSMTP(
		cfg.Mailer.Host, cfg.Mailer.Port,
		cfg.Mailer.Username, cfg.Mailer.Password,
		cfg.Mailer.FromName, cfg.Mailer.FromEmail,
	)
	asyncMailer := mailer.NewAsync(smtpMailer, logger)

	limiter := ratelimit.New(rdb)

	// ── Repositories ────────────────────────────────
	userRepo := user.NewPostgresRepository(db)
	authRepo := auth.NewPostgresRepository(db)
	profileRepo := profile.NewGormRepository(gdb)

	// ── Services ────────────────────────────────────
	tokenService := auth.NewTokenService(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
		cfg.JWT.Issuer,
	)
	authService := auth.NewService(userRepo, authRepo, tokenService, asyncMailer, cfg.OAuth.GoogleClientID, 8, limiter)
	userService := user.NewService(userRepo)
	profileService := profile.NewService(profileRepo, supabaseStorage)

	// ── Handlers ────────────────────────────────────
	authHandler := auth.NewHandler(authService)
	userHandler := user.NewHandler(userService)
	profileHandler := profile.NewHandler(profileService, userService, authService, supabaseStorage)

	return &App{
		Config: cfg,
		Logger: logger,
		DB:     db,
		Gorm:   gdb,
		Redis:  rdb,

		AuthHandler:    authHandler,
		UserHandler:    userHandler,
		AuthService:    authService,
		ProfileHandler: profileHandler,

		Limiter: limiter,
	}
}
