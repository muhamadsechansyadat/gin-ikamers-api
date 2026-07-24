package server

import (
	"context"
	"errors"
	"gin-ikamers-api/config"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	httpServer      *http.Server
	logger          *slog.Logger
	shutdownTimeout time.Duration
}

func New(
	router *gin.Engine,
	logger *slog.Logger,
	cfg config.AppConfig,
) *Server {

	return &Server{
		httpServer: &http.Server{
			Addr:              cfg.Host + ":" + cfg.Port,
			Handler:           router,
			ReadTimeout:       cfg.ReadTimeout,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
			WriteTimeout:      cfg.WriteTimeout,
			IdleTimeout:       cfg.IdleTimeout,
		},
		logger:          logger,
		shutdownTimeout: cfg.ShutdownTimeout,
	}
}

func (s *Server) Start() error {
	s.logger.Info("HTTP server started", "address", s.httpServer.Addr)

	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) WaitForShutdown() {

	quit := make(chan os.Signal, 1)

	signal.Notify(
		quit,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-quit

	s.logger.Info("Shutdown signal received")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		s.shutdownTimeout,
	)

	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {

		s.logger.Error(
			"Graceful shutdown failed",
			"error", err,
		)

		return
	}

	s.logger.Info("HTTP server stopped gracefully")
}
