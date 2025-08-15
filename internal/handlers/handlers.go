package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"bulbul/internal/models"
)

// Events handlers

// CreateEvent - POST /api/events
// Создать событие
func CreateEvent(c *gin.Context) {
	var req models.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем полученные данные для отладки
	log.Printf("Received event request: title=%s, external=%v", req.Title, req.External.Bool())

	// Захардкоженный ответ согласно спецификации
	response := models.CreateEventResponse{
		ID: 1, // Захардкоженный ID
	}

	c.JSON(http.StatusCreated, response)
}

// ListEvents - GET /api/events
// Получить список событий
func ListEvents(c *gin.Context) {
	// Захардкоженный ответ согласно спецификации
	response := models.ListEventsResponse{
		{
			ID:    1,
			Title: "Концерт классической музыки",
		},
		{
			ID:    2,
			Title: "Театральная постановка",
		},
	}

	c.JSON(http.StatusOK, response)
}

// Bookings handlers

// CreateBooking - POST /api/bookings
// Создать бронирование
func CreateBooking(c *gin.Context) {
	var req models.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Захардкоженный ответ согласно спецификации
	response := models.CreateBookingResponse{
		ID: 1, // Захардкоженный ID
	}

	c.JSON(http.StatusCreated, response)
}

// ListBookings - GET /api/bookings
// Получить список бронирований
func ListBookings(c *gin.Context) {
	// Захардкоженный ответ согласно спецификации
	response := models.ListBookingsResponse{
		{
			ID:      1,
			EventID: 1,
		},
		{
			ID:      2,
			EventID: 2,
		},
	}

	c.JSON(http.StatusOK, response)
}

// InitiatePayment - PATCH /api/bookings/initiatePayment
// Инициировать платеж для бронирования
func InitiatePayment(c *gin.Context) {
	var req models.InitiatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Согласно спецификации - возвращаем 200 без тела ответа
	c.Status(http.StatusOK)
}

// CancelBooking - PATCH /api/bookings/cancel
// Отменить бронирование
func CancelBooking(c *gin.Context) {
	var req models.CancelBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Согласно спецификации - возвращаем 200 без тела ответа
	c.Status(http.StatusOK)
}

// Seats handlers

// ListSeats - GET /api/seats
// Получить список мест
func ListSeats(c *gin.Context) {
	// Получаем параметры согласно спецификации
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	eventID, _ := strconv.ParseInt(c.Query("event_id"), 10, 64)

	if eventID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "event_id is required"})
		return
	}

	// Валидация параметров согласно спецификации
	if page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page must be >= 1"})
		return
	}

	if pageSize < 1 || pageSize > 20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pageSize must be between 1 and 20"})
		return
	}

	// Захардкоженный ответ согласно спецификации
	response := models.ListSeatsResponse{
		{
			ID:       1,
			Row:      1,
			Number:   1,
			Reserved: false,
		},
		{
			ID:       2,
			Row:      1,
			Number:   2,
			Reserved: true,
		},
		{
			ID:       3,
			Row:      2,
			Number:   1,
			Reserved: false,
		},
	}

	c.JSON(http.StatusOK, response)
}

// SelectSeat - PATCH /api/seats/select
// Выбрать место для брони
func SelectSeat(c *gin.Context) {
	var req models.SelectSeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Захардкоженная логика - всегда успешно
	// Согласно спецификации - возвращаем 200 без тела ответа
	c.Status(http.StatusOK)
}

// ReleaseSeat - PATCH /api/seats/release
// Убрать место из брони
func ReleaseSeat(c *gin.Context) {
	var req models.ReleaseSeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Захардкоженная логика - всегда успешно
	// Согласно спецификации - возвращаем 200 без тела ответа
	c.Status(http.StatusOK)
}

// Payments handlers

// NotifyPaymentCompleted - GET /api/payments/success
// Уведомить сервис, что платеж успешно проведен
func NotifyPaymentCompleted(c *gin.Context) {
	orderID := c.Query("orderId")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "orderId is required"})
		return
	}

	// Согласно спецификации - возвращаем 200 без тела ответа
	c.Status(http.StatusOK)
}

// NotifyPaymentFailed - GET /api/payments/fail
// Уведомить сервис, что платеж неуспешно проведен
func NotifyPaymentFailed(c *gin.Context) {
	orderID := c.Query("orderId")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "orderId is required"})
		return
	}

	// Согласно спецификации - возвращаем 200 без тела ответа
	c.Status(http.StatusOK)
}
