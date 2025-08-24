package models

import (
	"fmt"
	"strings"
)

// FlexibleBool - гибкий boolean тип, поддерживающий строки и числа
type FlexibleBool bool

// UnmarshalJSON поддерживает парсинг boolean из строки, числа и boolean
func (fb *FlexibleBool) UnmarshalJSON(data []byte) error {
	// Убираем кавычки
	str := string(data)
	str = strings.Trim(str, `"`)

	switch strings.ToLower(str) {
	case "true", "1", "yes", "on":
		*fb = true
	case "false", "0", "no", "off":
		*fb = false
	default:
		return fmt.Errorf("invalid boolean value: %s", str)
	}
	return nil
}

// Bool возвращает bool значение
func (fb FlexibleBool) Bool() bool {
	return bool(fb)
}

// CreateEventRequest - модель для создания события
type CreateEventRequest struct {
	Title    string       `json:"title" binding:"required"`
	External FlexibleBool `json:"external,omitempty"`
}

// CreateEventResponse - модель ответа при создании события
type CreateEventResponse struct {
	ID int64 `json:"id"`
}

// ListEventsResponseItem - элемент списка событий
type ListEventsResponseItem struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

// ListEventsResponse - список событий
type ListEventsResponse []ListEventsResponseItem

// ListBookingsResponseItem - элемент списка бронирований
type ListBookingsResponseItem struct {
	ID      int64                   `json:"id"`
	EventID int64                   `json:"event_id"`
	Seats   []ListSeatsResponseItem `json:"seats,omitempty"`
}

// ListBookingsResponse - список бронирований
type ListBookingsResponse []ListBookingsResponseItem

// CreateBookingRequest - модель для создания бронирования
type CreateBookingRequest struct {
	EventID int64 `json:"event_id" binding:"required"`
}

// CreateBookingResponse - модель ответа при создании бронирования
type CreateBookingResponse struct {
	ID int64 `json:"id"`
}

// ListSeatsResponseItem - элемент списка мест
type ListSeatsResponseItem struct {
	ID     string `json:"id"`
	Row    int64  `json:"row"`
	Number int64  `json:"number"`
	Status string `json:"status"`
	Price  string `json:"price"`
}

// ListSeatsResponse - список мест
type ListSeatsResponse []ListSeatsResponseItem

// SelectSeatRequest - модель для выбора места
type SelectSeatRequest struct {
	BookingID int64  `json:"booking_id" binding:"required"`
	SeatID    string `json:"seat_id" binding:"required"`
}

// ReleaseSeatRequest - модель для освобождения места
type ReleaseSeatRequest struct {
	SeatID string `json:"seat_id" binding:"required"`
}

// InitiatePaymentRequest - модель для инициации платежа
type InitiatePaymentRequest struct {
	BookingID int64 `json:"booking_id" binding:"required"`
}

// CancelBookingRequest - модель для отмены бронирования
type CancelBookingRequest struct {
	BookingID int64 `json:"booking_id" binding:"required"`
}

// PaymentNotificationPayload - модель для webhook уведомлений от платежного шлюза
type PaymentNotificationPayload struct {
	PaymentID string                 `json:"paymentId"`
	Status    string                 `json:"status"`
	TeamSlug  string                 `json:"teamSlug"`
	Timestamp string                 `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// AnalyticsResponse - модель ответа аналитики для события
type AnalyticsResponse struct {
	EventID       int64  `json:"event_id"`
	TotalSeats    int32  `json:"total_seats"`
	SoldSeats     int32  `json:"sold_seats"`
	ReservedSeats int32  `json:"reserved_seats"`
	FreeSeats     int32  `json:"free_seats"`
	TotalRevenue  string `json:"total_revenue"`
	BookingsCount int32  `json:"bookings_count"`
}
