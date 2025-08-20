package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"bulbul/internal/config"
	"bulbul/internal/database"
	"bulbul/internal/models"

	"github.com/google/uuid"
)

var (
	clearExisting = flag.Bool("clear", false, "Clear existing seats before generating new ones")
	eventID       = flag.Int("event", 0, "Generate seats only for specific event ID (0 = all events)")
	dryRun        = flag.Bool("dry-run", false, "Show what would be generated without making changes")
)

type SeatGenerator struct {
	db *database.DB
}

func main() {
	flag.Parse()

	slog.Info("Starting seat generator...")

	cfg := config.Load()
	db, err := database.Connect(cfg.Database)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.RunMigrations(); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	generator := &SeatGenerator{db: db}

	if err := generator.GenerateSeats(); err != nil {
		slog.Error("Failed to generate seats", "error", err)
		os.Exit(1)
	}

	slog.Info("Seat generation completed successfully!")
}

func (g *SeatGenerator) GenerateSeats() error {
	events, err := g.getEventsForSeatGeneration()
	if err != nil {
		return fmt.Errorf("failed to get events: %w", err)
	}

	if len(events) == 0 {
		slog.Info("No events found for seat generation")
		return nil
	}

	slog.Info("Found events for seat generation", "count", len(events))

	for _, event := range events {
		if err := g.generateSeatsForEvent(event); err != nil {
			slog.Error("Failed to generate seats for event", "event_id", event.ID, "title", event.Title, "error", err)
			continue
		}
		slog.Info("Generated seats for event", "event_id", event.ID, "title", event.Title)
	}

	return nil
}

func (g *SeatGenerator) getEventsForSeatGeneration() ([]models.Event, error) {
	query := `
		SELECT id, title, description, type, datetime_start, provider, external, total_seats, created_at, updated_at
		FROM events_archive 
		WHERE id != 1`

	args := []interface{}{}

	if *eventID > 0 {
		query += " AND id = $1"
		args = append(args, *eventID)
	}

	query += " ORDER BY id"

	rows, err := g.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
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

	return events, nil
}

func (g *SeatGenerator) generateSeatsForEvent(event models.Event) error {
	if !*clearExisting {
		existingCount, err := g.getExistingSeatCount(event.ID)
		if err != nil {
			return fmt.Errorf("failed to check existing seats: %w", err)
		}
		if existingCount > 0 {
			slog.Info("Event already has seats, skipping (use -clear to override)", "event_id", event.ID, "existing_count", existingCount)
			return nil
		}
	}

	if *dryRun {
		totalSeats := rand.Intn(901) + 100
		slog.Info("[DRY RUN] Would generate seats for event", "total_seats", totalSeats, "event_id", event.ID, "title", event.Title)
		return nil
	}

	tx, err := g.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if *clearExisting {
		if err := g.clearExistingSeats(tx, event.ID); err != nil {
			return fmt.Errorf("failed to clear existing seats: %w", err)
		}
	}

	seats := g.generateSeatLayout()
	totalSeats := len(seats)

	if err := g.insertSeats(tx, event.ID, seats); err != nil {
		return fmt.Errorf("failed to insert seats: %w", err)
	}

	if err := g.updateEventTotalSeats(tx, event.ID, totalSeats); err != nil {
		return fmt.Errorf("failed to update event total seats: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.Info("Generated seats for event", "total_seats", totalSeats, "event_id", event.ID)
	return nil
}

func (g *SeatGenerator) getExistingSeatCount(eventID int64) (int, error) {
	var count int
	err := g.db.QueryRow("SELECT COUNT(*) FROM seats WHERE event_id = $1", eventID).Scan(&count)
	return count, err
}

func (g *SeatGenerator) clearExistingSeats(tx *sql.Tx, eventID int64) error {
	_, err := tx.Exec("DELETE FROM seats WHERE event_id = $1", eventID)
	return err
}

func (g *SeatGenerator) generateSeatLayout() []SeatInfo {
	rand.Seed(time.Now().UnixNano())

	totalSeats := rand.Intn(901) + 100

	var seats []SeatInfo
	seatsGenerated := 0
	rowNumber := 1

	for seatsGenerated < totalSeats {
		seatsInRow := rand.Intn(11) + 10
		if seatsGenerated+seatsInRow > totalSeats {
			seatsInRow = totalSeats - seatsGenerated
		}

		for seatNum := 1; seatNum <= seatsInRow; seatNum++ {
			guid, _ := uuid.NewUUID()
			price := g.generateSeatPrice(rowNumber)
			seats = append(seats, SeatInfo{
				ID:     guid.String(),
				Row:    rowNumber,
				Number: seatNum,
				Price:  price,
			})
			seatsGenerated++
		}
		rowNumber++
	}

	return seats
}

func (g *SeatGenerator) generateSeatPrice(rowNumber int) int64 {
	basePrice := int64(2000)

	if rowNumber <= 3 {
		return basePrice + int64(rand.Intn(3000)+2000)
	} else if rowNumber <= 10 {
		return basePrice + int64(rand.Intn(2000)+1000)
	} else {
		return basePrice + int64(rand.Intn(1000))
	}
}

func (g *SeatGenerator) insertSeats(tx *sql.Tx, eventID int64, seats []SeatInfo) error {
	stmt := `
		INSERT INTO seats (id, event_id, row_number, seat_number, status, price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	now := time.Now()

	for _, seat := range seats {
		_, err := tx.Exec(stmt, seat.ID, eventID, seat.Row, seat.Number, "FREE", seat.Price, now, now)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *SeatGenerator) updateEventTotalSeats(tx *sql.Tx, eventID int64, totalSeats int) error {
	_, err := tx.Exec(
		"UPDATE events_archive SET total_seats = $1, updated_at = $2 WHERE id = $3",
		totalSeats, time.Now(), eventID,
	)
	return err
}

type SeatInfo struct {
	ID     string
	Row    int
	Number int
	Price  int64
}
