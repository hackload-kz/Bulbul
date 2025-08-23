package service

import (
	"bulbul/internal/external"
	"bulbul/internal/messaging"
	"bulbul/internal/repository"
)

type Services struct {
	Events   *EventService
	Seats    *SeatService
	Bookings *BookingService
	Reset    *ResetService
}

func NewServices(repos *repository.Repositories, natsClient *messaging.NATSClient, ticketingClient *external.TicketingClient, paymentClient *external.PaymentClient) *Services {
	eventService := NewEventService(repos.Events, repos.Seats, natsClient)
	seatService := NewSeatService(repos.Seats, repos.Events, repos.Bookings, ticketingClient, natsClient)
	bookingService := NewBookingService(repos.Bookings, repos.Events, repos.Seats, paymentClient, ticketingClient, natsClient)
	resetService := NewResetService(repos.Bookings, repos.Seats)

	return &Services{
		Events:   eventService,
		Seats:    seatService,
		Bookings: bookingService,
		Reset:    resetService,
	}
}
