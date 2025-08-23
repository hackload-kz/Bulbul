package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"bulbul/cmd/consumers/handlers"
	"bulbul/cmd/consumers/jobs"
	"bulbul/internal/config"
	"bulbul/internal/consumers"
	"bulbul/internal/database"
	"bulbul/internal/external"
	"bulbul/internal/logger"
	"bulbul/internal/messaging"
	"bulbul/internal/models"
	"bulbul/internal/repository"
)

func main() {
	logger.Init("consumers", "info")
	slog.Info("Starting consumers service...")

	// Load configuration
	cfg := config.Load()

	// Override NATS client ID for consumers
	cfg.NATS.ClientID = "biletter-consumers"

	// Connect to database
	slog.Info("Connecting to database")
	db, err := database.Connect(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}
	defer db.Close()

	// Connect to NATS
	slog.Info("Connecting to NATS")
	natsClient, err := messaging.NewNATSClient(cfg.NATS)
	if err != nil {
		logger.Fatal("Failed to connect to NATS", "error", err)
	}
	defer natsClient.Close()

	// Create external clients
	ticketingClient := external.NewTicketingClient(cfg.Ticketing)

	// Create repositories
	bookingRepo := repository.NewBookingRepository(db)
	seatRepo := repository.NewSeatRepository(db)

	// Create handlers
	externalSyncHandler := handlers.NewExternalSyncHandler(ticketingClient, bookingRepo, seatRepo)

	// Subscribe to payment completed events
	slog.Info("Subscribing to payment completed events")
	_, err = natsClient.Subscribe(models.EventPaymentCompleted, externalSyncHandler.HandlePaymentCompleted)
	if err != nil {
		logger.Fatal("Failed to subscribe to payment completed events", "error", err)
	}

	// Create and start booking expiration job (background job only runs in consumers)
	slog.Info("Starting booking expiration job")
	expirationJob := jobs.NewBookingExpirationJob(bookingRepo, seatRepo, natsClient)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	expirationJob.Start(ctx)
	defer expirationJob.Stop()

	// Also start the existing consumer service if it exists
	consumerService, err := consumers.NewConsumerService(cfg)
	if err != nil {
		slog.Warn("Failed to create existing consumer service", "error", err)
	} else {
		if err := consumerService.Start(); err != nil {
			slog.Error("Failed to start existing consumers", "error", err)
		} else {
			defer consumerService.Shutdown(context.Background())
		}
	}

	slog.Info("Consumers service started successfully")

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down consumers service...")

	// Graceful shutdown handled by deferred calls
	slog.Info("Consumers service stopped")
}