package handlers

import (
	"log/slog"
	"net/http"

	"bulbul/internal/models"

	"github.com/gin-gonic/gin"
)

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
