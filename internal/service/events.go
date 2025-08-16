package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"bulbul/internal/messaging"
	"bulbul/internal/models"
	"bulbul/internal/repository"
)

type EventService struct {
	eventRepo  *repository.EventRepository
	seatRepo   *repository.SeatRepository
	natsClient *messaging.NATSClient
}

func NewEventService(eventRepo *repository.EventRepository, seatRepo *repository.SeatRepository, natsClient *messaging.NATSClient) *EventService {
	return &EventService{
		eventRepo:  eventRepo,
		seatRepo:   seatRepo,
		natsClient: natsClient,
	}
}

func (s *EventService) Create(ctx context.Context, req *models.CreateEventRequest) (*models.CreateEventResponse, error) {
	// For demonstration, we'll create a simple event
	// In real implementation, we'd need more data
	event := &models.Event{
		Title:         req.Title,
		Type:          "concert",                          // Default type
		DatetimeStart: time.Now().Add(7 * 24 * time.Hour), // One week from now
		Provider:      "Билеттер",
		External:      req.External.Bool(),
		TotalSeats:    0, // Will be set when seats are created
	}

	err := s.eventRepo.Create(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	// If not external event (ID != 1), generate random seats
	if !event.External && event.ID != 1 {
		rows := 10 + rand.Intn(21)        // 10-30 rows
		seatsPerRow := 15 + rand.Intn(16) // 15-30 seats per row

		err = s.seatRepo.CreateSeatsForEvent(ctx, event.ID, rows, seatsPerRow)
		if err != nil {
			return nil, fmt.Errorf("failed to create seats for event: %w", err)
		}
	}

	return &models.CreateEventResponse{ID: event.ID}, nil
}

func (s *EventService) List(ctx context.Context, query, date string, page, pageSize int) ([]models.ListEventsResponseItem, error) {
	events, err := s.eventRepo.List(ctx, query, date, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	result := make([]models.ListEventsResponseItem, len(events))
	for i, event := range events {
		result[i] = models.ListEventsResponseItem{
			ID:    event.ID,
			Title: event.Title,
		}
	}

	return result, nil
}
