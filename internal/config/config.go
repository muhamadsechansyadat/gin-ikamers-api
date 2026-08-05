package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"log/slog"
	"strings"
	"time"
)

type AppConfig struct {
	Name              string
	Env               string
	Host              string
	Port              string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	TrustedProxies    []string
}

type DatabaseConfig struct {
	Host            string
	Port            string
	Name            string
	User            string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration
	PingTimeout     time.Duration
}

type LogConfig struct {
	Level  string
	Format string
}

type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Issuer          string
}

type OAuthConfig struct {
	GoogleClientID string
}

type StorageConfig struct {
	SupabaseURL           string
	SupabaseServiceKey    string
	SupabaseStorageBucket string
}

type MailerConfig struct {
	Host      string
	Port      int
	Username  string
	Password  string
	FromName  string
	FromEmail string
}

type Config struct {
	App     AppConfig
	DB      DatabaseConfig
	Log     LogConfig
	JWT     JWTConfig
	OAuth   OAuthConfig
	Storage StorageConfig
	Mailer  MailerConfig
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Warn(".env not loaded, using system environment", "error", err)
	}

	cfg := &Config{
		App: AppConfig{
			Name:              getString("APP_NAME", "Gin IKamers API"),
			Env:               getString("APP_ENV", "development"),
			Host:              getString("APP_HOST", "localhost"),
			Port:              getString("APP_PORT", "8080"),
			ReadTimeout:       getDuration("APP_READ_TIMEOUT", 10*time.Second),
			ReadHeaderTimeout: getDuration("APP_READ_HEADER_TIMEOUT", 5*time.Second),
			WriteTimeout:      getDuration("APP_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:       getDuration("APP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout:   getDuration("APP_SHUTDOWN_TIMEOUT", 10*time.Second),
			TrustedProxies:    getStringSlice("APP_TRUSTED_PROXIES", nil),
		},
		DB: DatabaseConfig{
			Host:            getString("DB_HOST", ""),
			Port:            getString("DB_PORT", "5432"),
			Name:            getString("DB_NAME", ""),
			User:            getString("DB_USER", ""),
			Password:        getString("DB_PASSWORD", ""),
			SSLMode:         getString("DB_SSLMODE", "disable"),
			MaxOpenConns:    getInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxIdleTime: getDuration("DB_CONN_MAX_IDLE_TIME", 5*time.Second),
			ConnMaxLifetime: getDuration("DB_CONN_MAX_LIFE_TIME", 30*time.Second),
			PingTimeout:     getDuration("DB_PING_TIMEOUT", 5*time.Second),
		},
		Log: LogConfig{
			Level:  getString("LOG_LEVEL", "info"),
			Format: getString("LOG_FORMAT", "text"),
		},
		JWT: JWTConfig{
			Secret:          getString("JWT_SECRET", ""),
			AccessTokenTTL:  getDuration("JWT_ACCESS_TOKEN_TTL", 360*time.Minute),
			RefreshTokenTTL: getDuration("JWT_REFRESH_TOKEN_TTL", 720*time.Hour),
			Issuer:          getString("JWT_ISSUER", "ikamers-api"),
		},
		OAuth: OAuthConfig{
			GoogleClientID: getString("GOOGLE_CLIENT_ID", ""),
		},
		Storage: StorageConfig{
			SupabaseURL:           getString("SUPABASE_URL", ""),
			SupabaseServiceKey:    getString("SUPABASE_SERVICE_KEY", ""),
			SupabaseStorageBucket: getString("SUPABASE_STORAGE_BUCKET", "assets"),
		},
		Mailer: MailerConfig{
			Host:      getString("MAIL_HOST", "smtp.gmail.com"),
			Port:      getInt("MAIL_PORT", 587),
			Username:  getString("MAIL_USERNAME", ""),
			Password:  getString("MAIL_PASSWORD", ""),
			FromName:  getString("MAIL_FROM_NAME", "IKamers"),
			FromEmail: getString("MAIL_FROM_EMAIL", ""),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	var errs []string

	if c.DB.Host == "" {
		errs = append(errs, "DB_HOST is required")
	}
	if c.DB.Name == "" {
		errs = append(errs, "DB_NAME is required")
	}
	if c.DB.Port == "" {
		errs = append(errs, "DB_PORT is required")
	}
	if c.DB.User == "" {
		errs = append(errs, "DB_USER is required")
	}
	if len(errs) > 0 {
		return fmt.Errorf("invalid DB configuration: %s", strings.Join(errs, "; "))
	}
	return nil
}
