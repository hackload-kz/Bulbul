package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"time"

	"bulbul/internal/config"
	"bulbul/internal/database"
	"bulbul/internal/external"
	"bulbul/internal/logger"
	"bulbul/internal/models"
	"bulbul/internal/repository"
)

const (
	TargetEventID = 1 // Event ID for external service
	TotalRows     = 100
	SeatsPerRow   = 1000
	TotalSeats    = TotalRows * SeatsPerRow
)

func main() {
	var eventID int64
	flag.Int64Var(&eventID, "event-id", TargetEventID, "Event ID to sync seats for")
	flag.Parse()

	logger.Init("sync-seats", "info")
	slog.Info("Starting seat synchronization", "event_id", eventID)

	// Load configuration
	cfg := config.Load()

	// Connect to database
	slog.Info("Connecting to database")
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create repository
	seatRepo := repository.NewSeatRepository(db)

	// Create external ticketing client
	ticketingClient := external.NewTicketingClient(cfg.Ticketing)

	// Run synchronization
	if err := syncSeats(context.Background(), seatRepo, ticketingClient, eventID); err != nil {
		log.Fatalf("Seat synchronization failed: %v", err)
	}

	slog.Info("Seat synchronization completed successfully")
}

func syncSeats(ctx context.Context, seatRepo *repository.SeatRepository, ticketingClient *external.TicketingClient, eventID int64) error {
	start := time.Now()

	// Step 1: Clear existing seats for the event
	slog.Info("Clearing existing seats", "event_id", eventID)
	if err := seatRepo.DeleteSeatsByEventID(ctx, eventID); err != nil {
		return fmt.Errorf("failed to clear existing seats: %w", err)
	}
	slog.Info("Existing seats cleared")

	// Step 2: Fetch all places from external service
	slog.Info("Fetching places from external service")
	allPlaces, err := fetchAllPlaces(ticketingClient)
	if err != nil {
		return fmt.Errorf("failed to fetch places from external service: %w", err)
	}
	slog.Info("Fetched places from external service", "count", len(allPlaces))

	// Step 3: Convert external places to seats with pricing
	seats := make([]models.Seat, 0, len(allPlaces))
	for _, place := range allPlaces {
		price := getPriceForSeat(place.Row)
		status := "FREE"
		if !place.IsFree {
			status = "RESERVED"
		}

		seats = append(seats, models.Seat{
			ID:      place.ID,
			EventID: eventID,
			Row:     place.Row,
			Number:  place.Seat,
			Status:  status,
			Price:   &price,
		})
	}

	// Step 4: Bulk insert seats
	slog.Info("Inserting seats into database", "count", len(seats))
	if err := seatRepo.BulkCreateSeats(ctx, seats); err != nil {
		return fmt.Errorf("failed to bulk create seats: %w", err)
	}

	elapsed := time.Since(start)
	slog.Info("Seat synchronization completed",
		"event_id", eventID,
		"seats_processed", len(seats),
		"duration", elapsed.String(),
		"seats_per_second", float64(len(seats))/elapsed.Seconds())

	return nil
}

func fetchAllPlaces(ticketingClient *external.TicketingClient) ([]external.Place, error) {
	var allPlaces []external.Place
	pageSize := 1000
	page := 1

	for {
		slog.Info("Fetching places", "page", page, "page_size", pageSize)
		places, err := ticketingClient.GetPlaces(page, pageSize)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch places page %d: %w", page, err)
		}

		if len(places) == 0 {
			break
		}

		allPlaces = append(allPlaces, places...)

		// If we got less than pageSize, we're done
		if len(places) < pageSize {
			break
		}

		page++

		// Safety check to prevent infinite loops
		if len(allPlaces) >= TotalSeats*2 {
			slog.Warn("Fetched more places than expected, stopping",
				"fetched", len(allPlaces),
				"expected_max", TotalSeats)
			break
		}
	}

	return allPlaces, nil
}

// getPriceForSeat calculates the price based on the seat's row position
func getPriceForSeat(row int) int64 {
	switch {
	case row <= 10:
		return 40000 // Front rows - 40,000 tenge
	case row <= 25:
		return 80000 // Near front - 80,000 tenge
	case row <= 45:
		return 120000 // Middle - 120,000 tenge
	case row <= 70:
		return 160000 // Back middle - 160,000 tenge
	default:
		return 200000 // Far back - 200,000 tenge
	}
}
