package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"bulbul/internal/config"
	"bulbul/internal/consumers"
)

func main() {
	log.Println("Starting consumers service...")

	// Load configuration
	cfg := config.Load()

	// Override NATS client ID for consumers
	cfg.NATS.ClientID = "biletter-consumers"

	// Create and start consumers
	consumerService, err := consumers.NewConsumerService(cfg)
	if err != nil {
		log.Fatalf("Failed to create consumer service: %v", err)
	}

	// Start consuming messages
	if err := consumerService.Start(); err != nil {
		log.Fatalf("Failed to start consumers: %v", err)
	}

	log.Println("Consumers service started successfully")

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down consumers service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30)
	defer cancel()

	if err := consumerService.Shutdown(ctx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	log.Println("Consumers service stopped")
}