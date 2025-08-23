package repository

import (
	"bulbul/internal/database"
	"bulbul/internal/search"
)

type Repositories struct {
	Events   *EventElasticsearchRepository
	Seats    *SeatRepository
	Bookings *BookingRepository
	Users    *UserRepository
}

func NewRepositories(db *database.DB) *Repositories {
	return &Repositories{
		Events:   nil, // Will be set when Elasticsearch client is available
		Seats:    NewSeatRepository(db),
		Bookings: NewBookingRepository(db),
		Users:    NewUserRepository(db),
	}
}

func NewRepositoriesWithElasticsearch(db *database.DB, es *search.ElasticsearchClient) *Repositories {
	return &Repositories{
		Events:   NewEventElasticsearchRepository(es),
		Seats:    NewSeatRepository(db),
		Bookings: NewBookingRepository(db),
		Users:    NewUserRepository(db),
	}
}