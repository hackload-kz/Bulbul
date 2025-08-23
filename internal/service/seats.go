package service

import (
	"context"
	"fmt"
	"time"

	"bulbul/internal/external"
	"bulbul/internal/logger"
	"bulbul/internal/messaging"
	"bulbul/internal/models"
	"bulbul/internal/repository"
)

type SeatService struct {
	seatRepo        *repository.SeatRepository
	eventRepo       *repository.EventElasticsearchRepository
	bookingRepo     *repository.BookingRepository
	ticketingClient *external.TicketingClient
	natsClient      *messaging.NATSClient
}

func NewSeatService(seatRepo *repository.SeatRepository, eventRepo *repository.EventElasticsearchRepository, bookingRepo *repository.BookingRepository, ticketingClient *external.TicketingClient, natsClient *messaging.NATSClient) *SeatService {
	return &SeatService{
		seatRepo:        seatRepo,
		eventRepo:       eventRepo,
		bookingRepo:     bookingRepo,
		ticketingClient: ticketingClient,
		natsClient:      natsClient,
	}
}

func (s *SeatService) List(ctx context.Context, eventID int64, page, pageSize int, row *int, status *string) ([]models.ListSeatsResponseItem, error) {
	// For regular events, use database
	seats, err := s.seatRepo.GetByEventID(ctx, eventID, page, pageSize, row, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get seats: %w", err)
	}

	result := make([]models.ListSeatsResponseItem, len(seats))
	for i, seat := range seats {
		price := "0"
		if seat.Price != nil {
			price = fmt.Sprintf("%d", *seat.Price)
		}

		result[i] = models.ListSeatsResponseItem{
			ID:     seat.ID,
			Row:    int64(seat.Row),
			Number: int64(seat.Number),
			Status: seat.Status,
			Price:  price,
		}
	}

	return result, nil
}

func (s *SeatService) listExternalSeats(page, pageSize int) ([]models.ListSeatsResponseItem, error) {
	places, err := s.ticketingClient.GetPlaces(page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get external places: %w", err)
	}

	result := make([]models.ListSeatsResponseItem, len(places))
	for i, place := range places {
		status := "FREE"
		if !place.IsFree {
			status = "RESERVED"
		}

		result[i] = models.ListSeatsResponseItem{
			ID:     place.ID,
			Row:    int64(place.Row),
			Number: int64(place.Seat),
			Status: status,
		}
	}

	return result, nil
}

func (s *SeatService) Select(ctx context.Context, req *models.SelectSeatRequest) error {
	// First, get the booking to determine the event
	booking, err := s.getBookingByID(ctx, req.BookingID)
	if err != nil {
		return fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return fmt.Errorf("booking not found")
	}

	// If event ID = 1 (external), use ticketing service directly
	if booking.EventID == 1 {
		return s.selectExternalSeat(ctx, req)
	}

	// For regular events, get seat to verify it exists
	seat, err := s.seatRepo.GetByID(ctx, req.SeatID)
	if err != nil {
		return fmt.Errorf("failed to get seat: %w", err)
	}
	if seat == nil {
		return fmt.Errorf("seat not found")
	}

	// Verify the seat belongs to the same event as the booking
	if seat.EventID != booking.EventID {
		return fmt.Errorf("seat does not belong to the same event as the booking")
	}

	// Reserve the seat in database
	err = s.seatRepo.ReserveSeat(ctx, req.SeatID, req.BookingID)
	if err != nil {
		return fmt.Errorf("failed to reserve seat: %w", err)
	}

	// Publish seat selected event
	event := models.SeatSelectedEvent{
		BookingID: req.BookingID,
		SeatID:    req.SeatID,
		EventID:   seat.EventID,
		Timestamp: time.Now(),
	}

	if err := s.natsClient.Publish(models.EventSeatSelected, event); err != nil {
		// Log error but don't fail the operation
		logger.WithContext(ctx).Error("Failed to publish seat selected event",
			"error", err,
			"seat_id", req.SeatID,
			"booking_id", req.BookingID,
			"event_type", "seat.selected")
	}

	return nil
}

func (s *SeatService) selectExternalSeat(ctx context.Context, req *models.SelectSeatRequest) error {
	// For external seats, we need to map seat ID to place ID
	// This is simplified - in real implementation we'd need proper mapping
	placeID := req.SeatID // SeatID is already a string

	// Get the booking to get the external order ID
	booking, err := s.getBookingByID(ctx, req.BookingID)
	if err != nil {
		return fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return fmt.Errorf("booking not found")
	}

	if booking.OrderID == nil {
		return fmt.Errorf("external booking must have order ID")
	}

	orderID := *booking.OrderID

	err = s.ticketingClient.SelectPlace(placeID, orderID)
	if err != nil {
		return fmt.Errorf("failed to select external place: %w", err)
	}

	// Publish seat selected event
	event := models.SeatSelectedEvent{
		BookingID: req.BookingID,
		SeatID:    req.SeatID,
		EventID:   1, // External event
		Timestamp: time.Now(),
	}

	if err := s.natsClient.Publish(models.EventSeatSelected, event); err != nil {
		// Log error but don't fail the operation
		logger.WithContext(ctx).Error("Failed to publish seat selected event for external seat",
			"error", err,
			"seat_id", req.SeatID,
			"booking_id", req.BookingID,
			"event_type", "seat.selected")
	}

	return nil
}

func (s *SeatService) Release(ctx context.Context, req *models.ReleaseSeatRequest) error {
	// Get seat to determine event
	seat, err := s.seatRepo.GetByID(ctx, req.SeatID)
	if err != nil {
		return fmt.Errorf("failed to get seat: %w", err)
	}
	if seat == nil {
		return fmt.Errorf("seat not found")
	}

	// If event ID = 1 (external), use ticketing service
	if seat.EventID == 1 {
		return s.releaseExternalSeat(ctx, req)
	}

	// For regular events, use database
	err = s.seatRepo.ReleaseSeat(ctx, req.SeatID)
	if err != nil {
		return fmt.Errorf("failed to release seat: %w", err)
	}

	// Publish seat released event
	event := models.SeatReleasedEvent{
		SeatID:    req.SeatID,
		EventID:   seat.EventID,
		Timestamp: time.Now(),
	}

	if err := s.natsClient.Publish(models.EventSeatReleased, event); err != nil {
		// Log error but don't fail the operation
		logger.WithContext(ctx).Error("Failed to publish seat released event",
			"error", err,
			"seat_id", req.SeatID,
			"event_id", seat.EventID,
			"event_type", "seat.released")
	}

	return nil
}

func (s *SeatService) releaseExternalSeat(ctx context.Context, req *models.ReleaseSeatRequest) error {
	// For external seats, we need to get the booking to get the order ID
	booking, err := s.seatRepo.GetBookingBySeatID(ctx, req.SeatID)
	if err != nil {
		return fmt.Errorf("failed to get booking for seat: %w", err)
	}
	if booking == nil {
		return fmt.Errorf("no booking found for seat")
	}

	if booking.OrderID == nil {
		return fmt.Errorf("external booking must have order ID")
	}

	// For external seats, we need to map seat ID to place ID
	placeID := req.SeatID // SeatID is already a string

	err = s.ticketingClient.ReleasePlace(placeID)
	if err != nil {
		return fmt.Errorf("failed to release external place: %w", err)
	}

	// Publish seat released event
	event := models.SeatReleasedEvent{
		SeatID:    req.SeatID,
		EventID:   1, // External event
		Timestamp: time.Now(),
	}

	if err := s.natsClient.Publish(models.EventSeatReleased, event); err != nil {
		// Log error but don't fail the operation
		logger.WithContext(ctx).Error("Failed to publish seat released event for external seat",
			"error", err,
			"seat_id", req.SeatID,
			"event_type", "seat.released")
	}

	return nil
}

func (s *SeatService) getBookingByID(ctx context.Context, bookingID int64) (*models.Booking, error) {
	return s.bookingRepo.GetByID(ctx, bookingID)
}
