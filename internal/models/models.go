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
	Title    string      `json:"title" binding:"required"`
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
	ID      int64 `json:"id"`
	EventID int64 `json:"event_id"`
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
	ID       int64 `json:"id"`
	Row      int64 `json:"row"`
	Number   int64 `json:"number"`
	Reserved bool  `json:"reserved"`
}

// ListSeatsResponse - список мест
type ListSeatsResponse []ListSeatsResponseItem

// SelectSeatRequest - модель для выбора места
type SelectSeatRequest struct {
	BookingID int64 `json:"booking_id" binding:"required"`
	SeatID    int64 `json:"seat_id" binding:"required"`
}

// ReleaseSeatRequest - модель для освобождения места
type ReleaseSeatRequest struct {
	SeatID int64 `json:"seat_id" binding:"required"`
}

// InitiatePaymentRequest - модель для инициации платежа
type InitiatePaymentRequest struct {
	BookingID int64 `json:"booking_id" binding:"required"`
}

// CancelBookingRequest - модель для отмены бронирования
type CancelBookingRequest struct {
	BookingID int64 `json:"booking_id" binding:"required"`
}
