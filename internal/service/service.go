package service

import (
	"bulbul/internal/repository"
	"bulbul/internal/messaging"
	"bulbul/internal/external"
)

type Services struct {
	Events   *EventService
	Seats    *SeatService
	Bookings *BookingService
}

func NewServices(repos *repository.Repositories, natsClient *messaging.NATSClient, ticketingClient *external.TicketingClient, paymentClient *external.PaymentClient) *Services {
	eventService := NewEventService(repos.Events, repos.Seats, natsClient)
	seatService := NewSeatService(repos.Seats, repos.Events, ticketingClient, natsClient)
	bookingService := NewBookingService(repos.Bookings, repos.Events, repos.Seats, paymentClient, natsClient)

	return &Services{
		Events:   eventService,
		Seats:    seatService,
		Bookings: bookingService,
	}
}