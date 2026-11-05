package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/bladekkk/mcap/backend/internal/config"
	"github.com/bladekkk/mcap/backend/internal/database"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	cfg := config.LoadConfig()

	// Логирование
	logFile := filepath.Join(cfg.DataDir, "mcap.log")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		log.SetOutput(f)
		log.Println("MCAP started")
		defer f.Close()
	}

	log.Printf("Data dir: %s", cfg.DataDir)
	log.Printf("Cache dir: %s", cfg.CacheDir)

	// Инициализация БД
	dbPath := filepath.Join(cfg.DataDir, "mcap.db")
	db, err := database.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Database migrated successfully")

	app := fiber.New(fiber.Config{
		AppName: cfg.AppName,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	app.Use(logger.New())

	// API routes
	api := app.Group("/api/v1")

	// Servers
	api.Get("/servers", ListServers(db, cfg))
	api.Post("/servers", CreateServer(db, cfg))
	api.Get("/servers/:id", GetServer(db, cfg))
	api.Post("/servers/:id/start", StartServer(db, cfg))
	api.Post("/servers/:id/stop", StopServer(db, cfg))
	api.Post("/servers/:id/restart", RestartServer(db, cfg))
	api.Delete("/servers/:id", DeleteServer(db, cfg))

	// Java
	api.Get("/java", ListJDKs(db))
	api.Post("/java", InstallJDK(db, cfg))

	// Downloads
	api.Post("/downloads", DownloadJAR(db, cfg))

	// Metrics
	api.Get("/metrics/:id", GetMetrics(db))
	api.Get("/metrics/:id/history", GetMetricsHistory(db))

	// Plugins
	api.Get("/plugins", SearchPlugins(cfg))
	api.Post("/plugins/download", DownloadPlugin(db, cfg))

	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := app.Listen(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
