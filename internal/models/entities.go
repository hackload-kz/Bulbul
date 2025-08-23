package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	UserID        int64      `json:"user_id" db:"user_id"`
	Email         string     `json:"email" db:"email"`
	PasswordHash  string     `json:"-" db:"password_hash"`
	PasswordPlain *string    `json:"-" db:"password_plain"`
	FirstName     string     `json:"first_name" db:"first_name"`
	Surname       string     `json:"surname" db:"surname"`
	Birthday      *time.Time `json:"birthday" db:"birthday"`
	RegisteredAt  time.Time  `json:"registered_at" db:"registered_at"`
	IsActive      bool       `json:"is_active" db:"is_active"`
	LastLoggedIn  time.Time  `json:"last_logged_in" db:"last_logged_in"`
}

// Event represents an event in the system
type Event struct {
	ID            int64     `json:"id" db:"id"`
	Title         string    `json:"title" db:"title"`
	Description   *string   `json:"description" db:"description"`
	Type          string    `json:"type" db:"type"`
	DatetimeStart time.Time `json:"datetime_start" db:"datetime_start"`
	Provider      string    `json:"provider" db:"provider"`
	External      bool      `json:"external" db:"external"`
	TotalSeats    int       `json:"total_seats" db:"total_seats"`
}

// Seat represents a seat for an event
type Seat struct {
	ID        string    `json:"id" db:"id"`
	EventID   int64     `json:"event_id" db:"event_id"`
	Row       int       `json:"row" db:"row_number"`
	Number    int       `json:"number" db:"seat_number"`
	Status    string    `json:"status" db:"status"`
	Price     *int64    `json:"price" db:"price"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Booking represents a booking in the system
type Booking struct {
	ID            int64     `json:"id" db:"id"`
	EventID       int64     `json:"event_id" db:"event_id"`
	UserID        *int64    `json:"user_id" db:"user_id"`
	Status        string    `json:"status" db:"status"`
	PaymentStatus string    `json:"payment_status" db:"payment_status"`
	TotalAmount   *string   `json:"total_amount" db:"total_amount"`
	PaymentID     *string   `json:"payment_id" db:"payment_id"`
	OrderID       *string   `json:"order_id" db:"order_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	Seats         []Seat    `json:"seats,omitempty"` // Not from DB, filled separately
}

// BookingSeat represents the relationship between bookings and seats
type BookingSeat struct {
	ID         int64     `json:"id" db:"id"`
	BookingID  int64     `json:"booking_id" db:"booking_id"`
	SeatID     string    `json:"seat_id" db:"seat_id"`
	ReservedAt time.Time `json:"reserved_at" db:"reserved_at"`
}
