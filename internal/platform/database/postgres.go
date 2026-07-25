package database

import (
	"context"
	"database/sql"
	"fmt"
	"gin-ikamers-api/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log/slog"
)

func Connect(cfg config.DatabaseConfig, logger *slog.Logger) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.PingTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	logger.Info(
		"Connected to PostgreSQL",
		"host", cfg.Host,
		"database", cfg.Name,
	)

	return db, nil
}
