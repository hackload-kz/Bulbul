package jobs

import (
	"context"
	"log/slog"
	"time"

	"bulbul/internal/messaging"
	"bulbul/internal/models"
	"bulbul/internal/repository"
)

const BookingExpirationTimeout = 15 * time.Minute

// BookingExpirationJob handles the cleanup of expired bookings
type BookingExpirationJob struct {
	bookingRepo *repository.BookingRepository
	seatRepo    *repository.SeatRepository
	natsClient  *messaging.NATSClient
	ticker      *time.Ticker
	done        chan bool
}

// NewBookingExpirationJob creates a new booking expiration job
func NewBookingExpirationJob(bookingRepo *repository.BookingRepository, seatRepo *repository.SeatRepository, natsClient *messaging.NATSClient) *BookingExpirationJob {
	return &BookingExpirationJob{
		bookingRepo: bookingRepo,
		seatRepo:    seatRepo,
		natsClient:  natsClient,
		done:        make(chan bool),
	}
}

// Start begins the background job that checks for expired bookings every 30 seconds
func (j *BookingExpirationJob) Start(ctx context.Context) {
	slog.Info("Starting booking expiration job", "check_interval", "30s", "timeout", BookingExpirationTimeout)

	j.ticker = time.NewTicker(30 * time.Second)

	// Run initial check immediately
	go j.checkExpiredBookings(ctx)

	go func() {
		for {
			select {
			case <-j.ticker.C:
				go j.checkExpiredBookings(ctx)
			case <-j.done:
				slog.Info("Booking expiration job stopped")
				return
			}
		}
	}()
}

// Stop gracefully stops the background job
func (j *BookingExpirationJob) Stop() {
	if j.ticker != nil {
		j.ticker.Stop()
	}
	close(j.done)
}

// checkExpiredBookings finds and cancels bookings that have exceeded the 15-minute timeout
func (j *BookingExpirationJob) checkExpiredBookings(ctx context.Context) {
	// Find bookings created more than 15 minutes ago with status='CREATED' and payment_status='PENDING'
	expirationTime := time.Now().Add(-BookingExpirationTimeout)
	
	expiredBookings, err := j.bookingRepo.GetExpiredBookings(ctx, expirationTime)
	if err != nil {
		slog.Error("Failed to get expired bookings", "error", err)
		return
	}

	if len(expiredBookings) == 0 {
		slog.Debug("No expired bookings found")
		return
	}

	slog.Info("Found expired bookings to process", "count", len(expiredBookings))

	for _, booking := range expiredBookings {
		if err := j.expireBooking(ctx, &booking); err != nil {
			slog.Error("Failed to expire booking",
				"error", err,
				"booking_id", booking.ID,
				"event_id", booking.EventID,
				"created_at", booking.CreatedAt)
		} else {
			slog.Info("Successfully expired booking",
				"booking_id", booking.ID,
				"event_id", booking.EventID,
				"elapsed_time", time.Since(booking.CreatedAt).String())
		}
	}
}

// expireBooking cancels a specific booking and releases its seats
func (j *BookingExpirationJob) expireBooking(ctx context.Context, booking *models.Booking) error {
	slog.Info("Expiring booking", "booking_id", booking.ID)

	// Get seats for this booking
	seats, err := j.bookingRepo.GetSeats(ctx, booking.ID)
	if err != nil {
		return err
	}

	// Release all seats
	for _, seat := range seats {
		if err := j.seatRepo.ReleaseSeat(ctx, seat.ID); err != nil {
			slog.Error("Failed to release seat during expiration",
				"error", err,
				"seat_id", seat.ID,
				"booking_id", booking.ID)
			// Continue with other seats even if one fails
		}
	}

	// Update booking status to CANCELLED
	booking.Status = "CANCELLED"
	booking.PaymentStatus = "CANCELLED"
	if err := j.bookingRepo.Update(ctx, booking); err != nil {
		return err
	}

	// Publish booking expired event
	expirationEvent := models.BookingExpiredEvent{
		BookingID: booking.ID,
		EventID:   booking.EventID,
		Reason:    "15-minute timeout exceeded",
		UserID:    booking.UserID,
		Timestamp: time.Now(),
	}

	if err := j.natsClient.Publish(models.EventBookingExpired, expirationEvent); err != nil {
		slog.Error("Failed to publish booking expired event",
			"error", err,
			"booking_id", booking.ID,
			"event_type", "booking.expired")
		// Don't return error - expiration should still succeed
	}

	slog.Info("Booking expired successfully",
		"booking_id", booking.ID,
		"seats_released", len(seats))

	return nil
}