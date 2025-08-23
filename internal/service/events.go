package service

import (
	"context"
	"fmt"

	"bulbul/internal/messaging"
	"bulbul/internal/models"
	"bulbul/internal/repository"
)

type EventService struct {
	eventRepo  *repository.EventElasticsearchRepository
	seatRepo   *repository.SeatRepository
	natsClient *messaging.NATSClient
}

func NewEventService(eventRepo *repository.EventElasticsearchRepository, seatRepo *repository.SeatRepository, natsClient *messaging.NATSClient) *EventService {
	return &EventService{
		eventRepo:  eventRepo,
		seatRepo:   seatRepo,
		natsClient: natsClient,
	}
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
