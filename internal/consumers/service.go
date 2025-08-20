package consumers

import (
	"context"
	"log/slog"

	"bulbul/internal/config"
	"bulbul/internal/database"
	"bulbul/internal/repository"
	"bulbul/internal/messaging"
	"bulbul/internal/external"
)

type ConsumerService struct {
	db       *database.DB
	nats     *messaging.NATSClient
	repos    *repository.Repositories
	handlers *Handlers
}

func NewConsumerService(cfg *config.Config) (*ConsumerService, error) {
	// Connect to database
	db, err := database.Connect(cfg.Database)
	if err != nil {
		return nil, err
	}

	// Connect to NATS
	natsClient, err := messaging.NewNATSClient(cfg.NATS)
	if err != nil {
		return nil, err
	}

	// Create repositories
	repos := repository.NewRepositories(db)

	// Create external clients
	ticketingClient := external.NewTicketingClient(cfg.Ticketing)
	paymentClient := external.NewPaymentClient(cfg.Payment)

	// Create handlers
	handlers := NewHandlers(repos, ticketingClient, paymentClient)

	return &ConsumerService{
		db:       db,
		nats:     natsClient,
		repos:    repos,
		handlers: handlers,
	}, nil
}

func (cs *ConsumerService) Start() error {
	slog.Info("Starting NATS consumers...")

	// Subscribe to booking events
	_, err := cs.nats.SubscribeQueue("booking.created", "consumers", cs.handlers.HandleBookingCreated)
	if err != nil {
		return err
	}

	// Subscribe to payment events
	_, err = cs.nats.SubscribeQueue("payment.initiated", "consumers", cs.handlers.HandlePaymentInitiated)
	if err != nil {
		return err
	}

	_, err = cs.nats.SubscribeQueue("payment.completed", "consumers", cs.handlers.HandlePaymentCompleted)
	if err != nil {
		return err
	}

	_, err = cs.nats.SubscribeQueue("payment.failed", "consumers", cs.handlers.HandlePaymentFailed)
	if err != nil {
		return err
	}

	// Subscribe to seat events
	_, err = cs.nats.SubscribeQueue("seat.selected", "consumers", cs.handlers.HandleSeatSelected)
	if err != nil {
		return err
	}

	_, err = cs.nats.SubscribeQueue("seat.released", "consumers", cs.handlers.HandleSeatReleased)
	if err != nil {
		return err
	}

	// Subscribe to booking cancellation events
	_, err = cs.nats.SubscribeQueue("booking.cancelled", "consumers", cs.handlers.HandleBookingCancelled)
	if err != nil {
		return err
	}

	slog.Info("All consumers started successfully")
	return nil
}

func (cs *ConsumerService) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down consumer service...")

	if cs.nats != nil {
		if err := cs.nats.Close(); err != nil {
			slog.Error("Error closing NATS connection", "error", err)
		}
	}

	if cs.db != nil {
		if err := cs.db.Close(); err != nil {
			slog.Error("Error closing database connection", "error", err)
			return err
		}
	}

	return nil
}