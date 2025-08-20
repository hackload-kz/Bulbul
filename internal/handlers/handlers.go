package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"bulbul/internal/cache"
	"bulbul/internal/models"
	"bulbul/internal/service"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	services     *service.Services
	valkeyClient *cache.ValkeyClient
}

func NewHandlers(services *service.Services, valkeyClient *cache.ValkeyClient) *Handlers {
	return &Handlers{
		services:     services,
		valkeyClient: valkeyClient,
	}
}

// Events handlers

// CreateEvent - POST /api/events
// Создать событие
func (h *Handlers) CreateEvent(c *gin.Context) {
	var req models.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем полученные данные для отладки
	slog.Info("Received event request", "title", req.Title, "external", req.External.Bool())

	response, err := h.services.Events.Create(c.Request.Context(), &req)
	if err != nil {
		slog.Error("Failed to create event", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// ListEvents - GET /api/events
// Получить список событий
func (h *Handlers) ListEvents(c *gin.Context) {
	query := c.Query("query")
	date := c.Query("date")

	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	// Validate pagination parameters
	if page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page must be >= 1"})
		return
	}

	if pageSize < 1 || pageSize > 20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pageSize must be between 1 and 20"})
		return
	}

	// Check if we should use caching
	shouldCache := h.shouldCacheEventsRequest(query, date, pageSize)

	// Try to get from cache if conditions are met and cache client is available
	if shouldCache && h.valkeyClient != nil {
		// Use raw JSON to avoid unmarshaling/marshaling overhead
		// Wrap in recovery to prevent cache issues from crashing the handler
		rawJSON, err := h.valkeyClient.GetEventsListRaw(c.Request.Context(), page, pageSize)
		if err == nil {
			// Cache hit - return raw JSON data directly
			slog.Info("Cache hit for events list", "page", page, "pageSize", pageSize)
			c.Data(http.StatusOK, "application/json", rawJSON)
			return
		}
		// Cache miss or error - continue to fetch from database
		slog.Info("Cache miss for events list", "page", page, "pageSize", pageSize, "error", err)
	}

	// Fetch from database
	response, err := h.services.Events.List(c.Request.Context(), query, date, page, pageSize)
	if err != nil {
		slog.Error("Failed to list events", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list events"})
		return
	}

	// Store in cache if conditions are met and cache client is available
	if shouldCache && h.valkeyClient != nil {
		h.valkeyClient.SetEventsList(c.Request.Context(), page, pageSize, response)
	}

	c.JSON(http.StatusOK, response)
}

// shouldCacheEventsRequest determines if the request should be cached
func (h *Handlers) shouldCacheEventsRequest(query, date string, pageSize int) bool {
	// Don't cache if query or date parameters are provided
	if query != "" || date != "" {
		return false
	}

	// Only cache if pageSize is divisible by 5
	return pageSize%5 == 0
}

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

// InitiatePayment - PATCH /api/bookings/initiatePayment
// Инициировать платеж для бронирования
func (h *Handlers) InitiatePayment(c *gin.Context) {
	var req models.InitiatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	paymentURL, err := h.services.Bookings.InitiatePayment(c.Request.Context(), &req)
	if err != nil {
		slog.Error("Failed to initiate payment", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate payment"})
		return
	}

	// If payment URL is provided, redirect to payment gateway
	if paymentURL != "" {
		c.Header("Location", paymentURL)
		c.Status(http.StatusFound) // 302
	} else {
		// For external events, no payment needed
		c.Status(http.StatusOK)
	}
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
		slog.Error("Failed to select seat", "error", err)
		c.JSON(419, gin.H{"error": "Failed to select seat"})
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
		slog.Error("Failed to release seat", "error", err)
		c.JSON(419, gin.H{"error": "Failed to release seat"})
		return
	}

	// Согласно спецификации - возвращаем 200 без тела ответа
	c.Status(http.StatusOK)
}

// Payments handlers

// NotifyPaymentCompleted - GET /api/payments/success
// Уведомить сервис, что платеж успешно проведен
func (h *Handlers) NotifyPaymentCompleted(c *gin.Context) {
	orderID := c.Query("orderId")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "orderId is required"})
		return
	}

	// For now, just log the success
	slog.Info("Payment completed for order", "order_id", orderID)

	// Согласно спецификации - возвращаем 200 без тела ответа
	c.Status(http.StatusOK)
}

// NotifyPaymentFailed - GET /api/payments/fail
// Уведомить сервис, что платеж неуспешно проведен
func (h *Handlers) NotifyPaymentFailed(c *gin.Context) {
	orderID := c.Query("orderId")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "orderId is required"})
		return
	}

	// For now, just log the failure
	slog.Error("Payment failed for order", "order_id", orderID)

	// Согласно спецификации - возвращаем 200 без тела ответа
	c.Status(http.StatusOK)
}

// OnPaymentUpdates - POST /api/payments/notifications
// Принимать уведомления от платежного шлюза
func (h *Handlers) OnPaymentUpdates(c *gin.Context) {
	var notification models.PaymentNotificationPayload
	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.services.Bookings.HandlePaymentNotification(c.Request.Context(), &notification)
	if err != nil {
		slog.Error("Failed to handle payment notification", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to handle notification"})
		return
	}

	// Согласно спецификации - возвращаем 200 без тела ответа
	c.Status(http.StatusOK)
}
