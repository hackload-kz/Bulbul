package repository

import (
	"context"
	"database/sql"
	"time"

	"bulbul/internal/database"
	"bulbul/internal/models"
)

type BookingRepository struct {
	db *database.DB
}

func NewBookingRepository(db *database.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) Create(ctx context.Context, booking *models.Booking) error {
	query := `
		INSERT INTO bookings (event_id, order_id, user_id, status, payment_status, total_amount)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		booking.EventID,
		booking.OrderID,
		booking.UserID,
		booking.Status,
		booking.PaymentStatus,
		booking.TotalAmount,
	).Scan(&booking.ID, &booking.CreatedAt, &booking.UpdatedAt)

	return err
}

func (r *BookingRepository) GetByID(ctx context.Context, id int64) (*models.Booking, error) {
	booking := &models.Booking{}
	query := `
		SELECT id, event_id, user_id, status, payment_status, total_amount, 
		       payment_id, order_id, created_at, updated_at
		FROM bookings
		WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&booking.ID,
		&booking.EventID,
		&booking.UserID,
		&booking.Status,
		&booking.PaymentStatus,
		&booking.TotalAmount,
		&booking.PaymentID,
		&booking.OrderID,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return booking, err
}

func (r *BookingRepository) GetByUserID(ctx context.Context, userID int64) ([]models.Booking, error) {
	var bookings []models.Booking
	query := `
		SELECT id, event_id, user_id, status, payment_status, total_amount,
		       payment_id, order_id, created_at, updated_at
		FROM bookings
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var booking models.Booking
		err := rows.Scan(
			&booking.ID,
			&booking.EventID,
			&booking.UserID,
			&booking.Status,
			&booking.PaymentStatus,
			&booking.TotalAmount,
			&booking.PaymentID,
			&booking.OrderID,
			&booking.CreatedAt,
			&booking.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}

	return bookings, rows.Err()
}

func (r *BookingRepository) Update(ctx context.Context, booking *models.Booking) error {
	query := `
		UPDATE bookings 
		SET status = $1, payment_status = $2, total_amount = $3, 
		    payment_id = $4, order_id = $5, updated_at = NOW()
		WHERE id = $6`

	_, err := r.db.ExecContext(ctx, query,
		booking.Status,
		booking.PaymentStatus,
		booking.TotalAmount,
		booking.PaymentID,
		booking.OrderID,
		booking.ID,
	)

	return err
}

func (r *BookingRepository) AddSeat(ctx context.Context, bookingID int64, seatID string) error {
	query := `INSERT INTO booking_seats (booking_id, seat_id) VALUES ($1, $2)`
	_, err := r.db.ExecContext(ctx, query, bookingID, seatID)
	return err
}

func (r *BookingRepository) GetSeats(ctx context.Context, bookingID int64) ([]models.Seat, error) {
	var seats []models.Seat
	query := `
		SELECT s.id, s.event_id, s.row_number, s.seat_number, s.status, s.price, s.created_at, s.updated_at
		FROM seats s
		JOIN booking_seats bs ON s.id = bs.seat_id
		WHERE bs.booking_id = $1
		ORDER BY s.row_number, s.seat_number`

	rows, err := r.db.QueryContext(ctx, query, bookingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var seat models.Seat
		err := rows.Scan(
			&seat.ID,
			&seat.EventID,
			&seat.Row,
			&seat.Number,
			&seat.Status,
			&seat.Price,
			&seat.CreatedAt,
			&seat.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		seats = append(seats, seat)
	}

	return seats, rows.Err()
}

func (r *BookingRepository) UpdatePaymentStatus(ctx context.Context, id int64, status string, paymentID string) error {
	query := `
		UPDATE bookings 
		SET payment_status = $1, payment_id = $2, updated_at = NOW()
		WHERE id = $3`

	_, err := r.db.ExecContext(ctx, query, status, paymentID, id)
	return err
}

// GetByPaymentID retrieves a booking by payment ID
func (r *BookingRepository) GetByPaymentID(ctx context.Context, paymentID string) (*models.Booking, error) {
	booking := &models.Booking{}
	query := `
		SELECT id, event_id, user_id, status, payment_status, total_amount, 
		       payment_id, order_id, created_at, updated_at
		FROM bookings
		WHERE payment_id = $1`

	err := r.db.QueryRowContext(ctx, query, paymentID).Scan(
		&booking.ID,
		&booking.EventID,
		&booking.UserID,
		&booking.Status,
		&booking.PaymentStatus,
		&booking.TotalAmount,
		&booking.PaymentID,
		&booking.OrderID,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return booking, err
}

// GetExpiredBookings retrieves bookings that have exceeded the expiration time
func (r *BookingRepository) GetExpiredBookings(ctx context.Context, expirationTime time.Time) ([]models.Booking, error) {
	var bookings []models.Booking
	query := `
		SELECT id, event_id, user_id, status, payment_status, total_amount, 
		       payment_id, order_id, created_at, updated_at
		FROM bookings
		WHERE status = 'CREATED' 
		  AND payment_status = 'PENDING'
		  AND created_at < $1
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, expirationTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var booking models.Booking
		err := rows.Scan(
			&booking.ID,
			&booking.EventID,
			&booking.UserID,
			&booking.Status,
			&booking.PaymentStatus,
			&booking.TotalAmount,
			&booking.PaymentID,
			&booking.OrderID,
			&booking.CreatedAt,
			&booking.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}

	return bookings, rows.Err()
}
