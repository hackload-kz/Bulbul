package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"bulbul/internal/external"
	"bulbul/internal/models"
	"bulbul/internal/repository"

	"github.com/nats-io/stan.go"
)

// ExternalSyncHandler handles payment completion events and syncs to external service
type ExternalSyncHandler struct {
	ticketingClient *external.TicketingClient
	bookingRepo     *repository.BookingRepository
	seatRepo        *repository.SeatRepository
}

// NewExternalSyncHandler creates a new external sync handler
func NewExternalSyncHandler(ticketingClient *external.TicketingClient, bookingRepo *repository.BookingRepository, seatRepo *repository.SeatRepository) *ExternalSyncHandler {
	return &ExternalSyncHandler{
		ticketingClient: ticketingClient,
		bookingRepo:     bookingRepo,
		seatRepo:        seatRepo,
	}
}

// HandlePaymentCompleted handles payment completion events
func (h *ExternalSyncHandler) HandlePaymentCompleted(msg *stan.Msg) {
	ctx := context.Background()

	var event models.PaymentCompletedEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		slog.Error("Failed to unmarshal payment completed event", "error", err)
		msg.Ack() // Acknowledge even on unmarshal error to avoid redelivery
		return
	}

	slog.Info("Processing payment completed event", "payment_id", event.PaymentID)

	// Get booking by payment ID
	booking, err := h.bookingRepo.GetByPaymentID(ctx, event.PaymentID)
	if err != nil {
		slog.Error("Failed to get booking by payment ID", "error", err, "payment_id", event.PaymentID)
		msg.Ack() // Acknowledge to avoid redelivery of database errors
		return
	}
	if booking == nil {
		slog.Warn("No booking found for payment ID", "payment_id", event.PaymentID)
		msg.Ack() // Acknowledge - no booking to process
		return
	}

	// Only sync for event_id=1 (external events)
	if booking.EventID != 1 {
		slog.Debug("Skipping external sync for non-external event", "event_id", booking.EventID, "booking_id", booking.ID)
		msg.Ack() // Acknowledge - no sync needed for non-external events
		return
	}

	slog.Info("Starting external sync for event_id=1", "booking_id", booking.ID)

	// Sync to external service
	if err := h.syncToExternalService(ctx, booking); err != nil {
		slog.Error("Failed to sync booking to external service", 
			"error", err, 
			"booking_id", booking.ID, 
			"event_id", booking.EventID)
		// TODO: Implement retry mechanism - for now, do not acknowledge to allow retry
		return
	}

	slog.Info("Successfully synced booking to external service", "booking_id", booking.ID)
	msg.Ack() // Acknowledge successful processing
}

// syncToExternalService syncs a booking to the external ticketing service
func (h *ExternalSyncHandler) syncToExternalService(ctx context.Context, booking *models.Booking) error {
	// Step 1: Start order with external service
	startOrderResp, err := h.ticketingClient.StartOrder()
	if err != nil {
		return fmt.Errorf("failed to start external order: %w", err)
	}

	slog.Info("Started external order", "order_id", startOrderResp.OrderID, "booking_id", booking.ID)

	// Step 2: Get seats for this booking
	seats, err := h.bookingRepo.GetSeats(ctx, booking.ID)
	if err != nil {
		return fmt.Errorf("failed to get seats for booking: %w", err)
	}

	if len(seats) == 0 {
		return fmt.Errorf("no seats found for booking %d", booking.ID)
	}

	// Step 3: Select places in external system
	for _, seat := range seats {
		// Use seat ID as place ID (external service expects string ID)
		placeID := seat.ID

		err := h.ticketingClient.SelectPlace(placeID, startOrderResp.OrderID)
		if err != nil {
			// If place selection fails, try to cancel the order
			_ = h.ticketingClient.CancelOrder(startOrderResp.OrderID)
			return fmt.Errorf("failed to select place %s: %w", placeID, err)
		}

		slog.Debug("Selected external place", "place_id", placeID, "order_id", startOrderResp.OrderID)
	}

	// Step 4: Submit order
	err = h.ticketingClient.SubmitOrder(startOrderResp.OrderID)
	if err != nil {
		// Try to cancel the order if submit fails
		_ = h.ticketingClient.CancelOrder(startOrderResp.OrderID)
		return fmt.Errorf("failed to submit external order: %w", err)
	}

	// Step 5: Confirm order
	err = h.ticketingClient.ConfirmOrder(startOrderResp.OrderID)
	if err != nil {
		// Try to cancel the order if confirm fails
		_ = h.ticketingClient.CancelOrder(startOrderResp.OrderID)
		return fmt.Errorf("failed to confirm external order: %w", err)
	}

	// Step 6: Update booking with external order ID
	booking.OrderID = &startOrderResp.OrderID
	err = h.bookingRepo.Update(ctx, booking)
	if err != nil {
		slog.Error("Failed to update booking with external order ID", 
			"error", err, 
			"booking_id", booking.ID, 
			"order_id", startOrderResp.OrderID)
		// Don't return error here as the external order is already confirmed
	}

	slog.Info("Successfully synced booking to external service", 
		"booking_id", booking.ID, 
		"external_order_id", startOrderResp.OrderID,
		"seats_count", len(seats))

	return nil
}