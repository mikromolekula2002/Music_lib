package main

import (
	"context"
	"log"
	_ "mikromolekula2002/music_library_ver1.0/docs"
	"mikromolekula2002/music_library_ver1.0/internal/config"
	"mikromolekula2002/music_library_ver1.0/internal/repository"
	"mikromolekula2002/music_library_ver1.0/internal/router"
	"mikromolekula2002/music_library_ver1.0/internal/service"
	"mikromolekula2002/music_library_ver1.0/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate"
)

// @title Music Library API
// @version 1.0
// @description This is a RESTful API for managing a music library.

// @host localhost:8080
// @BasePath
func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal(err)
	}

	loger := logger.InitLogger(cfg.LoggerLevel, cfg.LoggerOut, cfg.LoggerFilePath)
	loger.Info("Starting the application...")

	loger.Debug("Connecting to the database...")
	songRepo, err := repository.NewRepository(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	if err != nil {
		loger.Fatal("Database connection failed: ", err)
	}
	loger.Debug("Connected to the database successfully.")

	loger.Debug("Applying database migrations...")
	if err := songRepo.ApplyMigrations("migration"); err != nil {
		if err == migrate.ErrNoChange {
			loger.Debug("Nothing to update schema")
		} else {
			loger.Warn("Migration failed: ", err)
		}
	}
	loger.Debug("Migrations applied successfully.")

	loger.Debug("Initializing services and router...")
	songService := service.NewSongService(songRepo, loger, cfg.MusicAPIHost, cfg.MusicBaseURL)
	songRouter := router.NewRouter(songService)
	songRouter.SetRoutes(cfg.EnvType)
	loger.Debug("Router initialized.")

	// Создаем сервер с тайм-аутами
	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: songRouter.Gin, // Используем Gin в качестве обработчика
	}

	// Канал для перехвата сигналов завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		loger.Infof("Starting server on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			loger.Fatalf("Server failed: %v", err)
		}
	}()

	// Ожидание сигнала завершения
	<-quit
	loger.Warn("Shutting down server...")

	// Контекст с тайм-аутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		loger.Errorf("Server forced to shutdown: %v", err)
	}

	loger.Info("Server stopped gracefully")
}
