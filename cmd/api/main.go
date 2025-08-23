package main

import (
	"context"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bulbul/internal/api"
	"bulbul/internal/config"
	"bulbul/internal/logger"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.Load()

	// Инициализируем логгер
	logger.Init(cfg.LogLevel, cfg.LogFormat)

	slog.Info("Application starting",
		"log_level", cfg.LogLevel,
		"log_format", cfg.LogFormat,
		"gin_mode", cfg.GinMode)

	// Создаем и настраиваем сервер
	server := api.NewServer(cfg)

	// Запускаем pprof сервер если включен
	if cfg.PprofEnabled {
		go func() {
			slog.Info("Starting pprof server", "port", cfg.PprofPort)
			pprofServer := &http.Server{
				Addr:    ":" + cfg.PprofPort,
				Handler: http.DefaultServeMux, // pprof handlers are registered here
			}
			if err := pprofServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("pprof server failed", "error", err)
			}
		}()
	}

	// Создаем HTTP сервер
	srv := &http.Server{
		Addr:           ":" + cfg.Port,
		Handler:        server.GetRouter(),
		ReadTimeout:    cfg.RequestTimeout,
		WriteTimeout:   cfg.RequestTimeout,
		IdleTimeout:    2 * time.Minute,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Запускаем сервер в отдельной горутине
	go func() {
		slog.Info("Starting HTTP server", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", "error", err)
		}
	}()

	// Ждем сигнал для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	// Graceful shutdown с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	// Закрываем соединения
	if err := server.Cleanup(); err != nil {
		slog.Error("Error during cleanup", "error", err)
	}

	slog.Info("Server stopped gracefully")
}
