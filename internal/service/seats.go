package service

import (
	"context"
	"fmt"

	"bulbul/internal/external"
	"bulbul/internal/messaging"
	"bulbul/internal/middleware"
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
	// All events (including event_id=1) use local database
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

func (s *SeatService) Select(ctx context.Context, req *models.SelectSeatRequest) error {
	// First, get the booking to determine the event
	booking, err := s.getBookingByID(ctx, req.BookingID)
	if err != nil {
		return fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return fmt.Errorf("booking not found")
	}

	// Authorization: verify user owns this booking
	if userID, ok := middleware.UserIDFromContext(ctx); ok {
		if booking.UserID == nil || *booking.UserID != userID {
			return fmt.Errorf("unauthorized: booking does not belong to current user")
		}
	} else {
		return fmt.Errorf("unauthorized: user not authenticated")
	}

	// Get seat to verify it exists (all events use local database)
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

	// Get booking to verify ownership
	booking, err := s.seatRepo.GetBookingBySeatID(ctx, req.SeatID)
	if err != nil {
		return fmt.Errorf("failed to get booking for seat: %w", err)
	}
	if booking == nil {
		return fmt.Errorf("no booking found for seat")
	}

	// Authorization: verify user owns this booking
	if userID, ok := middleware.UserIDFromContext(ctx); ok {
		if booking.UserID == nil || *booking.UserID != userID {
			return fmt.Errorf("unauthorized: booking does not belong to current user")
		}
	} else {
		return fmt.Errorf("unauthorized: user not authenticated")
	}

	// Release seat from database (all events use local database)
	err = s.seatRepo.ReleaseSeat(ctx, req.SeatID)
	if err != nil {
		return fmt.Errorf("failed to release seat: %w", err)
	}

	return nil
}

func (s *SeatService) getBookingByID(ctx context.Context, bookingID int64) (*models.Booking, error) {
	return s.bookingRepo.GetByID(ctx, bookingID)
}
