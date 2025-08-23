package repository

import (
	"context"
	"math"
	"time"

	"bulbul/internal/models"
	"bulbul/internal/search"
)

// EventElasticsearchRepository реализует репозиторий событий с использованием Elasticsearch
type EventElasticsearchRepository struct {
	es *search.ElasticsearchClient
}

// NewEventElasticsearchRepository создает новый репозиторий событий с Elasticsearch
func NewEventElasticsearchRepository(es *search.ElasticsearchClient) *EventElasticsearchRepository {
	return &EventElasticsearchRepository{es: es}
}

// Create создает новое событие
func (r *EventElasticsearchRepository) Create(ctx context.Context, event *models.Event) error {
	// Set timestamps
	now := time.Now()
	if event.CreatedAt.IsZero() {
		event.CreatedAt = now
	}
	if event.UpdatedAt.IsZero() {
		event.UpdatedAt = now
	}

	return r.es.IndexEvent(ctx, event)
}

// GetByID получает событие по ID
func (r *EventElasticsearchRepository) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	return r.es.GetByID(ctx, id)
}

// List возвращает список событий с поддержкой поиска, фильтрации и пагинации
func (r *EventElasticsearchRepository) List(ctx context.Context, query string, date string, page, pageSize int) ([]models.Event, error) {
	return r.es.Search(ctx, query, date, page, pageSize)
}

// Update обновляет событие
func (r *EventElasticsearchRepository) Update(ctx context.Context, event *models.Event) error {
	return r.es.UpdateEvent(ctx, event)
}

// Delete удаляет событие
func (r *EventElasticsearchRepository) Delete(ctx context.Context, id int64) error {
	return r.es.DeleteEvent(ctx, id)
}

// Count возвращает общее количество событий с учетом фильтров
func (r *EventElasticsearchRepository) Count(ctx context.Context, query string, date string) (int64, error) {
	return r.es.Count(ctx, query, date)
}

// GetTotalPages вычисляет общее количество страниц для пагинации
func (r *EventElasticsearchRepository) GetTotalPages(ctx context.Context, query string, date string, pageSize int) (int, error) {
	if pageSize <= 0 {
		return 0, nil
	}

	totalCount, err := r.Count(ctx, query, date)
	if err != nil {
		return 0, err
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))
	return totalPages, nil
}