package service

import (
	"context"
	"log/slog"

	"bulbul/internal/repository"
)

type ResetService struct {
	bookingRepo *repository.BookingRepository
	seatRepo    *repository.SeatRepository
}

func NewResetService(bookingRepo *repository.BookingRepository, seatRepo *repository.SeatRepository) *ResetService {
	return &ResetService{
		bookingRepo: bookingRepo,
		seatRepo:    seatRepo,
	}
}

// ResetDatabase removes all bookings and marks all seats as free
func (s *ResetService) ResetDatabase(ctx context.Context) error {
	slog.Info("Starting database reset")

	// First delete all bookings and booking_seats
	if err := s.bookingRepo.DeleteAll(ctx); err != nil {
		slog.Error("Failed to delete all bookings", "error", err)
		return err
	}
	slog.Info("All bookings deleted successfully")

	// Then reset all seats to FREE status
	if err := s.seatRepo.ResetAllSeats(ctx); err != nil {
		slog.Error("Failed to reset all seats", "error", err)
		return err
	}
	slog.Info("All seats reset to FREE status")

	slog.Info("Database reset completed successfully")
	return nil
}