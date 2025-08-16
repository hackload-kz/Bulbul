package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bulbul/internal/api"
	"bulbul/internal/config"
	"bulbul/internal/validation"
)

func main() {
	// Проверяем, нужно ли запустить валидацию
	if len(os.Args) > 1 && os.Args[1] == "validate" {
		validation.RunValidation()
		return
	}

	// Загружаем конфигурацию
	cfg := config.Load()

	// Создаем и настраиваем сервер
	server := api.NewServer(cfg)

	// Создаем HTTP сервер
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: server.GetRouter(),
	}

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("Starting server on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Ждем сигнал для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Закрываем соединения
	if err := server.Cleanup(); err != nil {
		log.Printf("Error during cleanup: %v", err)
	}

	log.Println("Server stopped")
}
