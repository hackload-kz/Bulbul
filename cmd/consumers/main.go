package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"bulbul/internal/config"
	"bulbul/internal/consumers"
	"bulbul/internal/logger"
)

func main() {
	slog.Info("Starting consumers service...")

	// Load configuration
	cfg := config.Load()

	// Override NATS client ID for consumers
	cfg.NATS.ClientID = "biletter-consumers"

	// Create and start consumers
	consumerService, err := consumers.NewConsumerService(cfg)
	if err != nil {
		logger.Fatal("Failed to create consumer service", "error", err)
	}

	// Start consuming messages
	if err := consumerService.Start(); err != nil {
		logger.Fatal("Failed to start consumers", "error", err)
	}

	slog.Info("Consumers service started successfully")

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down consumers service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30)
	defer cancel()

	if err := consumerService.Shutdown(ctx); err != nil {
		slog.Error("Error during shutdown", "error", err)
	}

	slog.Info("Consumers service stopped")
}