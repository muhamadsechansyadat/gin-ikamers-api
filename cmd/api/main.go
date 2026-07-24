package main

import (
	"gin-ikamers-api/config"
	"gin-ikamers-api/handlers"
	"gin-ikamers-api/internal/database"
	"gin-ikamers-api/internal/middleware"
	"gin-ikamers-api/internal/routes"
	"gin-ikamers-api/internal/server"
	"gin-ikamers-api/logger"
	"github.com/gin-gonic/gin"
	"os"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config" + err.Error())
	}

	log := logger.New(cfg.Log)

	log.Info("starting server", "name", cfg.App.Name, "env", cfg.App.Env)

	db, err := database.Connect(cfg.DB, log)
	if err != nil {
		log.Error(
			"Failed to connect PostgreSQL",
			"error", err,
		)
		os.Exit(1)
	}

	defer db.Close()

	h := handlers.New(db, log, cfg)
	switch cfg.App.Env {
	case "production":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()
	if err := router.SetTrustedProxies(cfg.App.TrustedProxies); err != nil {
		log.Error("set trusted proxies", "error", err)
		os.Exit(1)
	}

	router.Use(
		middleware.Recovery(log),
		middleware.CORS(),
	)

	router.HandleMethodNotAllowed = true
	routes.RegisterRouter(router, h)
	router.NoRoute(handlers.NotFoundHandler)
	router.NoMethod(handlers.MethodNotAllowedHandler)

	srv := server.New(
		router,
		log,
		cfg.App,
	)

	go func() {
		if err := srv.Start(); err != nil {
			log.Error(
				"Cannot start server",
				"error", err,
			)
			os.Exit(1)
		}
	}()

	srv.WaitForShutdown()
}
