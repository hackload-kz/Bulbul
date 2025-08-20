package consumers

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/nats-io/stan.go"
	"bulbul/internal/models"
	"bulbul/internal/repository"
	"bulbul/internal/external"
)

type Handlers struct {
	repos           *repository.Repositories
	ticketingClient *external.TicketingClient
	paymentClient   *external.PaymentClient
}

func NewHandlers(repos *repository.Repositories, ticketingClient *external.TicketingClient, paymentClient *external.PaymentClient) *Handlers {
	return &Handlers{
		repos:           repos,
		ticketingClient: ticketingClient,
		paymentClient:   paymentClient,
	}
}

func (h *Handlers) HandleBookingCreated(m *stan.Msg) {
	var event models.BookingCreatedEvent
	if err := json.Unmarshal(m.Data, &event); err != nil {
		slog.Error("Failed to unmarshal booking created event", "error", err)
		return
	}

	slog.Info("Processing booking created event", "event", event)

	// For now, just acknowledge the message
	// In a real implementation, we might:
	// - Send confirmation emails
	// - Update analytics
	// - Trigger other business processes

	m.Ack()
}

func (h *Handlers) HandlePaymentInitiated(m *stan.Msg) {
	var event models.PaymentInitiatedEvent
	if err := json.Unmarshal(m.Data, &event); err != nil {
		slog.Error("Failed to unmarshal payment initiated event", "error", err)
		return
	}

	slog.Info("Processing payment initiated event", "event", event)

	// Update booking status to reflect payment initiation
	ctx := context.Background()
	booking, err := h.repos.Bookings.GetByID(ctx, event.BookingID)
	if err != nil {
		slog.Error("Failed to get booking", "booking_id", event.BookingID, "error", err)
		return
	}

	if booking != nil {
		booking.PaymentStatus = "INITIATED"
		if err := h.repos.Bookings.Update(ctx, booking); err != nil {
			slog.Error("Failed to update booking", "booking_id", event.BookingID, "error", err)
			return
		}
	}

	m.Ack()
}

func (h *Handlers) HandlePaymentCompleted(m *stan.Msg) {
	var event models.PaymentCompletedEvent
	if err := json.Unmarshal(m.Data, &event); err != nil {
		slog.Error("Failed to unmarshal payment completed event", "error", err)
		return
	}

	slog.Info("Processing payment completed event", "event", event)

	ctx := context.Background()

	// Update booking status
	booking, err := h.repos.Bookings.GetByID(ctx, event.BookingID)
	if err != nil {
		slog.Error("Failed to get booking", "booking_id", event.BookingID, "error", err)
		return
	}

	if booking != nil {
		booking.Status = "CONFIRMED"
		booking.PaymentStatus = "COMPLETED"
		if err := h.repos.Bookings.Update(ctx, booking); err != nil {
			slog.Error("Failed to update booking", "booking_id", event.BookingID, "error", err)
			return
		}

		// Update seat statuses to SOLD
		seats, err := h.repos.Bookings.GetSeats(ctx, booking.ID)
		if err != nil {
			slog.Error("Failed to get booking seats", "booking_id", event.BookingID, "error", err)
			return
		}

		for _, seat := range seats {
			if err := h.repos.Seats.UpdateStatus(ctx, seat.ID, "SOLD"); err != nil {
				slog.Error("Failed to update seat status", "seat_id", seat.ID, "error", err)
			}
		}

		// For external events (ID=1), confirm with ticketing service
		if booking.EventID == 1 {
			// In a real implementation, we'd need proper order ID mapping
			orderID := event.OrderID
			if err := h.ticketingClient.ConfirmOrder(orderID); err != nil {
				slog.Error("Failed to confirm external order", "order_id", orderID, "error", err)
			}
		}
	}

	m.Ack()
}

func (h *Handlers) HandlePaymentFailed(m *stan.Msg) {
	var event models.PaymentFailedEvent
	if err := json.Unmarshal(m.Data, &event); err != nil {
		slog.Error("Failed to unmarshal payment failed event", "error", err)
		return
	}

	slog.Info("Processing payment failed event", "event", event)

	ctx := context.Background()

	// Update booking status
	booking, err := h.repos.Bookings.GetByID(ctx, event.BookingID)
	if err != nil {
		slog.Error("Failed to get booking", "booking_id", event.BookingID, "error", err)
		return
	}

	if booking != nil {
		booking.Status = "CANCELLED"
		booking.PaymentStatus = "FAILED"
		if err := h.repos.Bookings.Update(ctx, booking); err != nil {
			slog.Error("Failed to update booking", "booking_id", event.BookingID, "error", err)
			return
		}

		// Release all seats
		seats, err := h.repos.Bookings.GetSeats(ctx, booking.ID)
		if err != nil {
			slog.Error("Failed to get booking seats", "booking_id", event.BookingID, "error", err)
			return
		}

		for _, seat := range seats {
			if err := h.repos.Seats.ReleaseSeat(ctx, seat.ID); err != nil {
				slog.Error("Failed to release seat", "seat_id", seat.ID, "error", err)
			}
		}

		// For external events (ID=1), cancel with ticketing service
		if booking.EventID == 1 {
			// In a real implementation, we'd need proper order ID mapping
			orderID := event.OrderID
			if err := h.ticketingClient.CancelOrder(orderID); err != nil {
				slog.Error("Failed to cancel external order", "order_id", orderID, "error", err)
			}
		}
	}

	m.Ack()
}

func (h *Handlers) HandleSeatSelected(m *stan.Msg) {
	var event models.SeatSelectedEvent
	if err := json.Unmarshal(m.Data, &event); err != nil {
		slog.Error("Failed to unmarshal seat selected event", "error", err)
		return
	}

	slog.Info("Processing seat selected event", "event", event)

	// For now, just log the event
	// In a real implementation, we might:
	// - Update analytics
	// - Send notifications
	// - Update caches

	m.Ack()
}

func (h *Handlers) HandleSeatReleased(m *stan.Msg) {
	var event models.SeatReleasedEvent
	if err := json.Unmarshal(m.Data, &event); err != nil {
		slog.Error("Failed to unmarshal seat released event", "error", err)
		return
	}

	slog.Info("Processing seat released event", "event", event)

	// For now, just log the event
	// In a real implementation, we might:
	// - Update analytics
	// - Send notifications
	// - Update caches

	m.Ack()
}

func (h *Handlers) HandleBookingCancelled(m *stan.Msg) {
	var event models.BookingCancelledEvent
	if err := json.Unmarshal(m.Data, &event); err != nil {
		slog.Error("Failed to unmarshal booking cancelled event", "error", err)
		return
	}

	slog.Info("Processing booking cancelled event", "event", event)

	// For now, just log the event
	// In a real implementation, we might:
	// - Send cancellation emails
	// - Update analytics
	// - Process refunds

	m.Ack()
}