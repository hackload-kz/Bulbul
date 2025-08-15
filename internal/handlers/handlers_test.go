package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"bulbul/internal/models"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	
	// API routes
	api := r.Group("/api")
	{
		events := api.Group("/events")
		{
			events.POST("", CreateEvent)
			events.GET("", ListEvents)
		}

		bookings := api.Group("/bookings")
		{
			bookings.POST("", CreateBooking)
			bookings.GET("", ListBookings)
			bookings.PATCH("/initiatePayment", InitiatePayment)
			bookings.PATCH("/cancel", CancelBooking)
		}

		seats := api.Group("/seats")
		{
			seats.GET("", ListSeats)
			seats.PATCH("/select", SelectSeat)
			seats.PATCH("/release", ReleaseSeat)
		}

		payments := api.Group("/payments")
		{
			payments.GET("/success", NotifyPaymentCompleted)
			payments.GET("/fail", NotifyPaymentFailed)
		}
	}
	
	return r
}

func TestCreateEvent(t *testing.T) {
	r := setupRouter()
	
	reqBody := models.CreateEventRequest{
		Title:    "Тестовое событие",
		External: false,
	}
	
	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/events", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response models.CreateEventResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), response.ID)
}

func TestListEvents(t *testing.T) {
	r := setupRouter()
	
	req, _ := http.NewRequest("GET", "/api/events", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.ListEventsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, "Концерт классической музыки", response[0].Title)
}

func TestCreateBooking(t *testing.T) {
	r := setupRouter()
	
	reqBody := models.CreateBookingRequest{
		EventID: 1,
	}
	
	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/bookings", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response models.CreateBookingResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), response.ID)
}

func TestListSeats(t *testing.T) {
	r := setupRouter()
	
	req, _ := http.NewRequest("GET", "/api/seats?event_id=1&page=1&pageSize=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.ListSeatsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 3)
}

func TestListSeatsValidation(t *testing.T) {
	r := setupRouter()
	
	// Тест без обязательного параметра event_id
	req, _ := http.NewRequest("GET", "/api/seats", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	// Тест с некорректным page
	req, _ = http.NewRequest("GET", "/api/seats?event_id=1&page=0", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	// Тест с некорректным pageSize
	req, _ = http.NewRequest("GET", "/api/seats?event_id=1&pageSize=25", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPaymentEndpoints(t *testing.T) {
	r := setupRouter()
	
	// Тест успешного платежа
	req, _ := http.NewRequest("GET", "/api/payments/success?orderId=123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Тест неуспешного платежа
	req, _ = http.NewRequest("GET", "/api/payments/fail?orderId=123", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Тест без orderId
	req, _ = http.NewRequest("GET", "/api/payments/success", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
