package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"bulbul/internal/external"
	"bulbul/internal/messaging"
	"bulbul/internal/models"
	"bulbul/internal/repository"
)

type SeatService struct {
	seatRepo        *repository.SeatRepository
	eventRepo       *repository.EventRepository
	ticketingClient *external.TicketingClient
	natsClient      *messaging.NATSClient
}

func NewSeatService(seatRepo *repository.SeatRepository, eventRepo *repository.EventRepository, ticketingClient *external.TicketingClient, natsClient *messaging.NATSClient) *SeatService {
	return &SeatService{
		seatRepo:        seatRepo,
		eventRepo:       eventRepo,
		ticketingClient: ticketingClient,
		natsClient:      natsClient,
	}
}

func (s *SeatService) List(ctx context.Context, eventID int64, page, pageSize int, row *int, status *string) ([]models.ListSeatsResponseItem, error) {
	// If event ID = 1 (external), use ticketing service
	if eventID == 1 {
		return s.listExternalSeats(page, pageSize)
	}

	// Check if event exists and is external
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	if event == nil {
		return nil, fmt.Errorf("event not found")
	}

	// For regular events, use database
	seats, err := s.seatRepo.GetByEventID(ctx, eventID, page, pageSize, row, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get seats: %w", err)
	}

	result := make([]models.ListSeatsResponseItem, len(seats))
	for i, seat := range seats {
		price := "0.00"
		if seat.Price != nil {
			price = fmt.Sprintf("%.2f", float64(*seat.Price)/100.0)
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
			Price:  "50.00", // Default price for external seats
		}
	}

	return result, nil
}

func (s *SeatService) Select(ctx context.Context, req *models.SelectSeatRequest) error {
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
		return s.selectExternalSeat(ctx, req)
	}

	// For regular events, use database
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
		fmt.Printf("Failed to publish seat selected event: %v", err)
	}

	return nil
}

func (s *SeatService) selectExternalSeat(ctx context.Context, req *models.SelectSeatRequest) error {
	// For external seats, we need to map seat ID to place ID
	// This is simplified - in real implementation we'd need proper mapping
	placeID := req.SeatID // SeatID is already a string
	orderID := strconv.FormatInt(req.BookingID, 10)

	err := s.ticketingClient.SelectPlace(placeID, orderID)
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
		fmt.Printf("Failed to publish seat selected event: %v", err)
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
		fmt.Printf("Failed to publish seat released event: %v", err)
	}

	return nil
}

func (s *SeatService) releaseExternalSeat(ctx context.Context, req *models.ReleaseSeatRequest) error {
	// For external seats, we need to map seat ID to place ID
	placeID := req.SeatID // SeatID is already a string

	err := s.ticketingClient.ReleasePlace(placeID)
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
		fmt.Printf("Failed to publish seat released event: %v", err)
	}

	return nil
}
