package repository

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"

	"bulbul/internal/database"
	"bulbul/internal/models"
)

type SeatRepository struct {
	db *database.DB
}

func NewSeatRepository(db *database.DB) *SeatRepository {
	return &SeatRepository{db: db}
}

func (r *SeatRepository) CreateSeatsForEvent(ctx context.Context, eventID int64, rows, seatsPerRow int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Generate random prices between 1000 and 10000 (in kopecks)
	for row := 1; row <= rows; row++ {
		for seat := 1; seat <= seatsPerRow; seat++ {
			price := 1000 + rand.Intn(9000) // Random price between 1000-10000 kopecks

			query := `
				INSERT INTO seats (event_id, row_number, seat_number, status, price)
				VALUES ($1, $2, $3, 'FREE', $4)`

			_, err := tx.ExecContext(ctx, query, eventID, row, seat, price)
			if err != nil {
				return err
			}
		}
	}

	// Update total seats count
	updateQuery := `UPDATE events_archive SET total_seats = $1 WHERE id = $2`
	_, err = tx.ExecContext(ctx, updateQuery, rows*seatsPerRow, eventID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *SeatRepository) GetByEventID(ctx context.Context, eventID int64, page, pageSize int, row *int, status *string) ([]models.Seat, error) {
	var seats []models.Seat
	var args []interface{}
	argIndex := 1

	query := `
		SELECT id, event_id, row_number, seat_number, status, price, created_at, updated_at
		FROM seats
		WHERE event_id = $1`
	args = append(args, eventID)
	argIndex++

	if row != nil {
		query += fmt.Sprintf(" AND row_number = $%d", argIndex)
		args = append(args, *row)
		argIndex++
	}

	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *status)
		argIndex++
	}

	query += " ORDER BY row_number, seat_number"

	// Add pagination
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, pageSize, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
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

func (r *SeatRepository) GetByID(ctx context.Context, id string) (*models.Seat, error) {
	seat := &models.Seat{}
	query := `
		SELECT id, event_id, row_number, seat_number, status, price, created_at, updated_at
		FROM seats
		WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&seat.ID,
		&seat.EventID,
		&seat.Row,
		&seat.Number,
		&seat.Status,
		&seat.Price,
		&seat.CreatedAt,
		&seat.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return seat, err
}

func (r *SeatRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE seats SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *SeatRepository) ReserveSeat(ctx context.Context, seatID string, bookingID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if seat is available
	var currentStatus string
	checkQuery := `SELECT status FROM seats WHERE id = $1 FOR UPDATE`
	err = tx.QueryRowContext(ctx, checkQuery, seatID).Scan(&currentStatus)
	if err != nil {
		return err
	}

	if currentStatus != "FREE" {
		return fmt.Errorf("seat is not available")
	}

	// Reserve the seat
	updateQuery := `UPDATE seats SET status = 'RESERVED', updated_at = NOW() WHERE id = $1`
	_, err = tx.ExecContext(ctx, updateQuery, seatID)
	if err != nil {
		return err
	}

	// Add to booking_seats
	insertQuery := `INSERT INTO booking_seats (booking_id, seat_id) VALUES ($1, $2)`
	_, err = tx.ExecContext(ctx, insertQuery, bookingID, seatID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *SeatRepository) ReleaseSeat(ctx context.Context, seatID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update seat status to FREE
	updateQuery := `UPDATE seats SET status = 'FREE', updated_at = NOW() WHERE id = $1`
	_, err = tx.ExecContext(ctx, updateQuery, seatID)
	if err != nil {
		return err
	}

	// Remove from booking_seats
	deleteQuery := `DELETE FROM booking_seats WHERE seat_id = $1`
	_, err = tx.ExecContext(ctx, deleteQuery, seatID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
