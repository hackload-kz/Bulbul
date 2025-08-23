package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"bulbul/internal/models"

	"github.com/gin-gonic/gin"
)

// Seats handlers

// ListSeats - GET /api/seats
// Получить список мест
func (h *Handlers) ListSeats(c *gin.Context) {
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

	// Опциональные параметры
	var row *int
	var status *string
	if rowParam := c.Query("row"); rowParam != "" {
		if r, err := strconv.Atoi(rowParam); err == nil {
			row = &r
		}
	}
	if statusParam := c.Query("status"); statusParam != "" {
		status = &statusParam
	}

	response, err := h.services.Seats.List(c.Request.Context(), eventID, page, pageSize, row, status)
	if err != nil {
		slog.Error("Failed to list seats", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list seats"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// SelectSeat - PATCH /api/seats/select
// Выбрать место для брони
func (h *Handlers) SelectSeat(c *gin.Context) {
	var req models.SelectSeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.services.Seats.Select(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to select seat")
		return
	}

	// Согласно спецификации - возвращаем 200 без тела ответа
	c.Status(http.StatusOK)
}

// ReleaseSeat - PATCH /api/seats/release
// Убрать место из брони
func (h *Handlers) ReleaseSeat(c *gin.Context) {
	var req models.ReleaseSeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.services.Seats.Release(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to release seat")
		return
	}

	// Согласно спецификации - возвращаем 200 без тела ответа
	c.Status(http.StatusOK)
}
