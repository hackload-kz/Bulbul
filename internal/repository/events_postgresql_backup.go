package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"bulbul/internal/database"
	"bulbul/internal/models"
)

type EventRepository struct {
	db *database.DB
}

func NewEventRepository(db *database.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(ctx context.Context, event *models.Event) error {
	query := `
		INSERT INTO events_archive (title, description, type, datetime_start, provider, external, total_seats)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		event.Title,
		event.Description,
		event.Type,
		event.DatetimeStart,
		event.Provider,
		event.External,
		event.TotalSeats,
	).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)

	return err
}

func (r *EventRepository) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	event := &models.Event{}
	query := `
		SELECT id, title, description, type, datetime_start, provider, external, total_seats, created_at, updated_at
		FROM events_archive
		WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&event.ID,
		&event.Title,
		&event.Description,
		&event.Type,
		&event.DatetimeStart,
		&event.Provider,
		&event.External,
		&event.TotalSeats,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return event, err
}

func (r *EventRepository) List(ctx context.Context, query string, date string, page, pageSize int) ([]models.Event, error) {
	var events []models.Event
	var args []interface{}
	argIndex := 1
	var searchQueryArgIndex int

	sqlQuery := `
		SELECT id, title, description, type, datetime_start, provider, external, total_seats, created_at, updated_at
		FROM events_archive
		WHERE 1=1`

	// Add search filter with full-text search
	if query != "" {
		// Use PostgreSQL full-text search with Russian language support
		searchQueryArgIndex = argIndex
		sqlQuery += fmt.Sprintf(" AND search_vector @@ to_tsquery('russian', $%d)", argIndex)
		
		// Prepare search query - handle multiple words and special characters
		searchQuery := prepareSearchQuery(query)
		
		args = append(args, searchQuery)
		argIndex++
	}

	// Add date filter
	if date != "" {
		sqlQuery += fmt.Sprintf(" AND DATE(datetime_start) = $%d", argIndex)
		args = append(args, date)
		argIndex++
	}

	// Add ordering - prioritize search relevance if searching, otherwise by ID
	if query != "" {
		sqlQuery += " ORDER BY ts_rank(search_vector, to_tsquery('russian', $" + fmt.Sprintf("%d", searchQueryArgIndex) + ")) DESC, id ASC"
	} else {
		sqlQuery += " ORDER BY id ASC"
	}

	// Add pagination
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		sqlQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, pageSize, offset)
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var event models.Event
		err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&event.Type,
			&event.DatetimeStart,
			&event.Provider,
			&event.External,
			&event.TotalSeats,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, rows.Err()
}

func (r *EventRepository) Update(ctx context.Context, event *models.Event) error {
	query := `
		UPDATE events_archive 
		SET title = $1, description = $2, type = $3, datetime_start = $4, 
		    provider = $5, external = $6, total_seats = $7, updated_at = $8
		WHERE id = $9`

	event.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		event.Title,
		event.Description,
		event.Type,
		event.DatetimeStart,
		event.Provider,
		event.External,
		event.TotalSeats,
		event.UpdatedAt,
		event.ID,
	)

	return err
}

// prepareSearchQuery formats a search query for PostgreSQL full-text search
func prepareSearchQuery(query string) string {
	// If query contains operators, return as-is
	if containsSearchOperators(query) {
		return query
	}
	
	// Split by spaces and handle each word
	words := strings.Fields(strings.TrimSpace(query))
	if len(words) == 0 {
		return ""
	}
	
	// Add prefix matching to each word and join with AND operator
	var formattedWords []string
	for _, word := range words {
		if word != "" {
			formattedWords = append(formattedWords, word+":*")
		}
	}
	
	return strings.Join(formattedWords, " & ")
}

// containsSearchOperators checks if the search query contains PostgreSQL search operators
func containsSearchOperators(query string) bool {
	operators := []string{"&", "|", "!", "(", ")", ":", "*"}
	for _, op := range operators {
		if strings.Contains(query, op) {
			return true
		}
	}
	return false
}
