package models

import "time"

// NATS Event Types
const (
	EventBookingCreated    = "booking.created"
	EventPaymentInitiated  = "payment.initiated"
	EventPaymentCompleted  = "payment.completed"
	EventPaymentFailed     = "payment.failed"
	EventSeatSelected      = "seat.selected"
	EventSeatReleased      = "seat.released"
	EventBookingCancelled  = "booking.cancelled"
)

// BookingCreatedEvent represents a booking creation event
type BookingCreatedEvent struct {
	BookingID int64     `json:"booking_id"`
	EventID   int64     `json:"event_id"`
	UserID    *int64    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

// PaymentInitiatedEvent represents a payment initiation event
type PaymentInitiatedEvent struct {
	BookingID   int64     `json:"booking_id"`
	EventID     int64     `json:"event_id"`
	TotalAmount int64     `json:"total_amount"`
	PaymentID   string    `json:"payment_id"`
	Timestamp   time.Time `json:"timestamp"`
}

// PaymentCompletedEvent represents a successful payment event
type PaymentCompletedEvent struct {
	BookingID int64     `json:"booking_id"`
	PaymentID string    `json:"payment_id"`
	OrderID   string    `json:"order_id"`
	Timestamp time.Time `json:"timestamp"`
}

// PaymentFailedEvent represents a failed payment event
type PaymentFailedEvent struct {
	BookingID int64     `json:"booking_id"`
	PaymentID string    `json:"payment_id"`
	OrderID   string    `json:"order_id"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

// SeatSelectedEvent represents a seat selection event
type SeatSelectedEvent struct {
	BookingID int64     `json:"booking_id"`
	SeatID    string    `json:"seat_id"`
	EventID   int64     `json:"event_id"`
	Timestamp time.Time `json:"timestamp"`
}

// SeatReleasedEvent represents a seat release event
type SeatReleasedEvent struct {
	BookingID int64     `json:"booking_id"`
	SeatID    string    `json:"seat_id"`
	EventID   int64     `json:"event_id"`
	Timestamp time.Time `json:"timestamp"`
}

// BookingCancelledEvent represents a booking cancellation event
type BookingCancelledEvent struct {
	BookingID int64     `json:"booking_id"`
	EventID   int64     `json:"event_id"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}