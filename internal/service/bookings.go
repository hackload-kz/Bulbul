package service

import (
	"context"
	"fmt"
	"time"

	"bulbul/internal/external"
	"bulbul/internal/logger"
	"bulbul/internal/messaging"
	"bulbul/internal/middleware"
	"bulbul/internal/models"
	"bulbul/internal/repository"

	"github.com/google/uuid"
)

type BookingService struct {
	bookingRepo     *repository.BookingRepository
	eventRepo       *repository.EventElasticsearchRepository
	seatRepo        *repository.SeatRepository
	paymentClient   *external.PaymentClient
	ticketingClient *external.TicketingClient
	natsClient      *messaging.NATSClient
}

func NewBookingService(bookingRepo *repository.BookingRepository, eventRepo *repository.EventElasticsearchRepository, seatRepo *repository.SeatRepository, paymentClient *external.PaymentClient, ticketingClient *external.TicketingClient, natsClient *messaging.NATSClient) *BookingService {
	return &BookingService{
		bookingRepo:     bookingRepo,
		eventRepo:       eventRepo,
		seatRepo:        seatRepo,
		paymentClient:   paymentClient,
		ticketingClient: ticketingClient,
		natsClient:      natsClient,
	}
}

func (s *BookingService) Create(ctx context.Context, req *models.CreateBookingRequest) (*models.CreateBookingResponse, error) {
	// Check if event exists
	event, err := s.eventRepo.GetByID(ctx, req.EventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	if event == nil {
		return nil, fmt.Errorf("event not found")
	}

	// Create booking
	booking := &models.Booking{
		EventID:       req.EventID,
		Status:        "CREATED",
		PaymentStatus: "PENDING",
		TotalAmount:   &[]string{"0"}[0], // Will be calculated when seats are added
	}

	// Set user_id from request context if present
	if id, ok := middleware.UserIDFromContext(ctx); ok {
		booking.UserID = &id
	}

	// For external events (event ID = 1), start order with ticketing service
	if req.EventID == 1 {
		startOrderResp, err := s.ticketingClient.StartOrder()
		if err != nil {
			return nil, fmt.Errorf("failed to start external order: %w", err)
		}

		// Set the external order ID
		booking.OrderID = &startOrderResp.OrderID
		booking.Status = "CONFIRMED"        // External bookings are confirmed immediately
		booking.PaymentStatus = "COMPLETED" // No payment needed for external events
	}

	err = s.bookingRepo.Create(ctx, booking)
	if err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	// Publish booking created event
	event_data := models.BookingCreatedEvent{
		BookingID: booking.ID,
		EventID:   booking.EventID,
		UserID:    booking.UserID,
		Timestamp: time.Now(),
	}

	if err := s.natsClient.Publish(models.EventBookingCreated, event_data); err != nil {
		// Log error but don't fail the operation
		logger.WithContext(ctx).Error("Failed to publish booking created event",
			"error", err,
			"booking_id", booking.ID,
			"event_type", "booking.created")
	}

	return &models.CreateBookingResponse{ID: booking.ID}, nil
}

func (s *BookingService) List(ctx context.Context, userID int64) ([]models.ListBookingsResponseItem, error) {
	bookings, err := s.bookingRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookings: %w", err)
	}

	result := make([]models.ListBookingsResponseItem, len(bookings))
	for i, booking := range bookings {
		result[i] = models.ListBookingsResponseItem{
			ID:      booking.ID,
			EventID: booking.EventID,
		}
	}

	return result, nil
}

func (s *BookingService) InitiatePayment(ctx context.Context, req *models.InitiatePaymentRequest) (string, error) {
	// Get booking
	booking, err := s.bookingRepo.GetByID(ctx, req.BookingID)
	if err != nil {
		return "", fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return "", fmt.Errorf("booking not found")
	}

	// Get booking seats to calculate total amount
	seats, err := s.bookingRepo.GetSeats(ctx, booking.ID)
	if err != nil {
		return "", fmt.Errorf("failed to get booking seats: %w", err)
	}

	if len(seats) == 0 {
		return "", fmt.Errorf("no seats in booking")
	}

	// Calculate total amount
	var totalAmount int64
	for _, seat := range seats {
		if seat.Price != nil {
			totalAmount += *seat.Price
		}
	}

	// Only use payment service for non-external events (ID != 1)
	if booking.EventID != 1 {
		// Generate unique order ID
		orderID := uuid.New().String()

		// Initialize payment
		paymentResp, err := s.paymentClient.InitPayment(totalAmount, orderID, "RUB", "Билет на мероприятие")
		if err != nil {
			return "", fmt.Errorf("failed to initialize payment: %w", err)
		}

		// Update booking with payment info
		booking.PaymentStatus = "INITIATED"
		booking.PaymentID = &paymentResp.PaymentID
		booking.OrderID = &orderID
		totalAmountStr := fmt.Sprintf("%d", totalAmount)
		booking.TotalAmount = &totalAmountStr

		err = s.bookingRepo.Update(ctx, booking)
		if err != nil {
			return "", fmt.Errorf("failed to update booking: %w", err)
		}

		// Publish payment initiated event
		event := models.PaymentInitiatedEvent{
			BookingID:   booking.ID,
			EventID:     booking.EventID,
			TotalAmount: totalAmount,
			PaymentID:   paymentResp.PaymentID,
			Timestamp:   time.Now(),
		}

		if err := s.natsClient.Publish(models.EventPaymentInitiated, event); err != nil {
			// Log error but don't fail the operation
			logger.WithContext(ctx).Error("Failed to publish payment initiated event",
				"error", err,
				"event_type", "payment.initiated")
		}

		// Return payment URL
		return paymentResp.PaymentURL, nil
	} else {
		// For external events (ID=1), just mark as payment not required
		booking.PaymentStatus = "COMPLETED"
		booking.Status = "CONFIRMED"
		totalAmountStr := fmt.Sprintf("%d", totalAmount)
		booking.TotalAmount = &totalAmountStr

		err = s.bookingRepo.Update(ctx, booking)
		if err != nil {
			return "", fmt.Errorf("failed to update booking: %w", err)
		}

		// For external events, return empty URL since no payment is needed
		return "", nil
	}
}

func (s *BookingService) Cancel(ctx context.Context, req *models.CancelBookingRequest) error {
	// Get booking
	booking, err := s.bookingRepo.GetByID(ctx, req.BookingID)
	if err != nil {
		return fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return fmt.Errorf("booking not found")
	}

	// Release all seats
	seats, err := s.bookingRepo.GetSeats(ctx, booking.ID)
	if err != nil {
		return fmt.Errorf("failed to get booking seats: %w", err)
	}

	for _, seat := range seats {
		if err := s.seatRepo.ReleaseSeat(ctx, seat.ID); err != nil {
			// Log error but continue
			logger.WithContext(ctx).Error("Failed to release seat during booking cancellation",
				"error", err,
				"seat_id", seat.ID)
		}
	}

	// Cancel payment if initiated
	if booking.PaymentID != nil && booking.PaymentStatus == "INITIATED" {
		if err := s.paymentClient.CancelPayment(*booking.PaymentID, "Booking cancelled by user"); err != nil {
			// Log error but continue
			logger.WithContext(ctx).Error("Failed to cancel payment during booking cancellation",
				"error", err,
				"payment_id", *booking.PaymentID)
		}
	}

	// Update booking status
	booking.Status = "CANCELLED"
	booking.PaymentStatus = "CANCELLED"

	err = s.bookingRepo.Update(ctx, booking)
	if err != nil {
		return fmt.Errorf("failed to update booking: %w", err)
	}

	// Publish booking cancelled event
	event := models.BookingCancelledEvent{
		BookingID: booking.ID,
		EventID:   booking.EventID,
		Reason:    "User cancellation",
		Timestamp: time.Now(),
	}

	if err := s.natsClient.Publish(models.EventBookingCancelled, event); err != nil {
		// Log error but don't fail the operation
		logger.WithContext(ctx).Error("Failed to publish booking cancelled event",
			"error", err,
			"event_type", "booking.cancelled")
	}

	return nil
}

func (s *BookingService) HandlePaymentNotification(ctx context.Context, notification *models.PaymentNotificationPayload) error {
	// For now, we'll skip detailed implementation as it requires more webhook handling logic
	// In a real system, we'd:
	// 1. Find booking by payment ID or order ID
	// 2. Update booking status based on notification status
	// 3. Confirm/cancel seats accordingly
	// 4. Publish appropriate events
	logger.WithContext(ctx).Info("Received payment notification",
		"payment_id", notification.PaymentID,
		"status", notification.Status)

	switch notification.Status {
	case "completed", "CONFIRMED":
		// Handle successful payment
		event := models.PaymentCompletedEvent{
			PaymentID: notification.PaymentID,
			Timestamp: time.Now(),
		}
		if err := s.natsClient.Publish(models.EventPaymentCompleted, event); err != nil {
			logger.WithContext(ctx).Error("Failed to publish payment completed event",
				"error", err,
				"payment_id", notification.PaymentID,
				"event_type", "payment.completed")
		}

	case "failed", "REJECTED", "CANCELLED":
		// Handle failed payment
		event := models.PaymentFailedEvent{
			PaymentID: notification.PaymentID,
			Reason:    notification.Status,
			Timestamp: time.Now(),
		}
		if err := s.natsClient.Publish(models.EventPaymentFailed, event); err != nil {
			logger.WithContext(ctx).Error("Failed to publish payment failed event",
				"error", err,
				"payment_id", notification.PaymentID,
				"event_type", "payment.failed")
		}
	}

	return nil
}
