package repository

import (
	"bulbul/internal/database"
)

type Repositories struct {
	Events   *EventRepository
	Seats    *SeatRepository
	Bookings *BookingRepository
	Users    *UserRepository
}

func NewRepositories(db *database.DB) *Repositories {
	return &Repositories{
		Events:   NewEventRepository(db),
		Seats:    NewSeatRepository(db),
		Bookings: NewBookingRepository(db),
		Users:    NewUserRepository(db),
	}
}