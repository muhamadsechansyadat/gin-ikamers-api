package main

import (
	"os"

	"gin-ikamers-api/internal/app"
	"gin-ikamers-api/internal/config"
	"gin-ikamers-api/internal/platform/database"
	"gin-ikamers-api/internal/server"
	"gin-ikamers-api/internal/shared/middleware"
	"gin-ikamers-api/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	log := logger.New(cfg.Log)
	log.Info("starting server", "name", cfg.App.Name, "env", cfg.App.Env)

	db, err := database.Connect(cfg.DB, log)
	if err != nil {
		log.Error("Failed to connect PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Wire semua dependencies
	application := app.New(db, log, cfg)

	// Setup Gin
	setGinMode(cfg.App.Env)
	router := gin.New()
	if err := router.SetTrustedProxies(cfg.App.TrustedProxies); err != nil {
		log.Error("set trusted proxies", "error", err)
		os.Exit(1)
	}
	router.Use(middleware.Recovery(log), middleware.CORS())
	router.HandleMethodNotAllowed = true

	// Register routes
	application.RegisterRoutes(router)

	// 404 & 405 handlers — pindahin ke response package (langkah selanjutnya)
	router.NoRoute(notFoundHandler)
	router.NoMethod(methodNotAllowedHandler)

	// Start server
	srv := server.New(router, log, cfg.App)
	go func() {
		if err := srv.Start(); err != nil {
			log.Error("Cannot start server", "error", err)
			os.Exit(1)
		}
	}()
	srv.WaitForShutdown()
}

func setGinMode(env string) {
	switch env {
	case "production":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}
}

func notFoundHandler(c *gin.Context) {
	c.JSON(404, gin.H{"error": "route not found"})
}

func methodNotAllowedHandler(c *gin.Context) {
	c.JSON(405, gin.H{"error": "method not allowed"})
}
