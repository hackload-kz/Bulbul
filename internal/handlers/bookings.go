package handlers

import (
	"log/slog"
	"net/http"

	"bulbul/internal/models"

	"github.com/gin-gonic/gin"
)

// Bookings handlers

// CreateBooking - POST /api/bookings
// Создать бронирование
func (h *Handlers) CreateBooking(c *gin.Context) {
	var req models.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.services.Bookings.Create(c.Request.Context(), &req)
	if err != nil {
		slog.Error("Failed to create booking", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking"})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// ListBookings - GET /api/bookings
// Получить список бронирований
func (h *Handlers) ListBookings(c *gin.Context) {
	userID := int64(1) // Default dummy user ID
	if v, exists := c.Get("user_id"); exists {
		if id, ok := v.(int64); ok {
			userID = id
		}
	}

	response, err := h.services.Bookings.List(c.Request.Context(), userID)
	if err != nil {
		slog.Error("Failed to list bookings", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list bookings"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CancelBooking - PATCH /api/bookings/cancel
// Отменить бронирование
func (h *Handlers) CancelBooking(c *gin.Context) {
	var req models.CancelBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.services.Bookings.Cancel(c.Request.Context(), &req)
	if err != nil {
		slog.Error("Failed to cancel booking", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel booking"})
		return
	}

	// Согласно спецификации - возвращаем 200 без тела ответа
	c.Status(http.StatusOK)
}
