package validation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"bulbul/internal/models"
)

// SpecValidator - валидатор соответствия API спецификации
type SpecValidator struct {
	baseURL string
}

// NewSpecValidator создает новый валидатор
func NewSpecValidator(baseURL string) *SpecValidator {
	return &SpecValidator{baseURL: baseURL}
}

// ValidateAll проверяет все endpoints на соответствие спецификации
func (v *SpecValidator) ValidateAll() error {
	log.Println("Начинаю валидацию API на соответствие спецификации...")

	// Проверяем Events endpoints
	if err := v.validateEvents(); err != nil {
		return fmt.Errorf("Events validation failed: %w", err)
	}

	// Проверяем Bookings endpoints
	if err := v.validateBookings(); err != nil {
		return fmt.Errorf("Bookings validation failed: %w", err)
	}

	// Проверяем Seats endpoints
	if err := v.validateSeats(); err != nil {
		return fmt.Errorf("Seats validation failed: %w", err)
	}

	// Проверяем Payments endpoints
	if err := v.validatePayments(); err != nil {
		return fmt.Errorf("Payments validation failed: %w", err)
	}

	log.Println("✅ Все endpoints прошли валидацию успешно!")
	return nil
}

func (v *SpecValidator) validateEvents() error {
	log.Println("Проверяю Events endpoints...")

	// POST /api/events
	reqBody := models.CreateEventRequest{
		Title:    "Тестовое событие",
		External: false,
	}
	
	resp, err := v.makeRequest("POST", "/api/events", reqBody)
	if err != nil {
		return err
	}
	
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("POST /api/events: expected 201, got %d", resp.StatusCode)
	}

	var createResp models.CreateEventResponse
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		return fmt.Errorf("POST /api/events: failed to decode response: %w", err)
	}
	resp.Body.Close()

	if createResp.ID == 0 {
		return fmt.Errorf("POST /api/events: expected non-zero ID")
	}

	// GET /api/events
	resp, err = v.makeRequest("GET", "/api/events", nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET /api/events: expected 200, got %d", resp.StatusCode)
	}

	var listResp models.ListEventsResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return fmt.Errorf("GET /api/events: failed to decode response: %w", err)
	}
	resp.Body.Close()

	if len(listResp) == 0 {
		return fmt.Errorf("GET /api/events: expected non-empty list")
	}

	log.Println("✅ Events endpoints валидны")
	return nil
}

func (v *SpecValidator) validateBookings() error {
	log.Println("Проверяю Bookings endpoints...")

	// POST /api/bookings
	reqBody := models.CreateBookingRequest{
		EventID: 1,
	}
	
	resp, err := v.makeRequest("POST", "/api/bookings", reqBody)
	if err != nil {
		return err
	}
	
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("POST /api/bookings: expected 201, got %d", resp.StatusCode)
	}

	var createResp models.CreateBookingResponse
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		return fmt.Errorf("POST /api/bookings: failed to decode response: %w", err)
	}
	resp.Body.Close()

	if createResp.ID == 0 {
		return fmt.Errorf("POST /api/bookings: expected non-zero ID")
	}

	// GET /api/bookings
	resp, err = v.makeRequest("GET", "/api/bookings", nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET /api/bookings: expected 200, got %d", resp.StatusCode)
	}

	var listResp models.ListBookingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return fmt.Errorf("GET /api/bookings: failed to decode response: %w", err)
	}
	resp.Body.Close()

	if len(listResp) == 0 {
		return fmt.Errorf("GET /api/bookings: expected non-empty list")
	}

	// PATCH /api/bookings/initiatePayment
	patchReq := models.InitiatePaymentRequest{
		BookingID: 1,
	}
	
	resp, err = v.makeRequest("PATCH", "/api/bookings/initiatePayment", patchReq)
	if err != nil {
		return err
	}
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("PATCH /api/bookings/initiatePayment: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// PATCH /api/bookings/cancel
	cancelReq := models.CancelBookingRequest{
		BookingID: 1,
	}
	
	resp, err = v.makeRequest("PATCH", "/api/bookings/cancel", cancelReq)
	if err != nil {
		return err
	}
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("PATCH /api/bookings/cancel: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	log.Println("✅ Bookings endpoints валидны")
	return nil
}

func (v *SpecValidator) validateSeats() error {
	log.Println("Проверяю Seats endpoints...")

	// GET /api/seats
	resp, err := v.makeRequest("GET", "/api/seats?event_id=1&page=1&pageSize=10", nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET /api/seats: expected 200, got %d", resp.StatusCode)
	}

	var listResp models.ListSeatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return fmt.Errorf("GET /api/seats: failed to decode response: %w", err)
	}
	resp.Body.Close()

	if len(listResp) == 0 {
		return fmt.Errorf("GET /api/seats: expected non-empty list")
	}

	// PATCH /api/seats/select
	selectReq := models.SelectSeatRequest{
		BookingID: 1,
		SeatID:    1,
	}
	
	resp, err = v.makeRequest("PATCH", "/api/seats/select", selectReq)
	if err != nil {
		return err
	}
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("PATCH /api/seats/select: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// PATCH /api/seats/release
	releaseReq := models.ReleaseSeatRequest{
		SeatID: 1,
	}
	
	resp, err = v.makeRequest("PATCH", "/api/seats/release", releaseReq)
	if err != nil {
		return err
	}
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("PATCH /api/seats/release: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	log.Println("✅ Seats endpoints валидны")
	return nil
}

func (v *SpecValidator) validatePayments() error {
	log.Println("Проверяю Payments endpoints...")

	// GET /api/payments/success
	resp, err := v.makeRequest("GET", "/api/payments/success?orderId=123", nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET /api/payments/success: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// GET /api/payments/fail
	resp, err = v.makeRequest("GET", "/api/payments/fail?orderId=123", nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET /api/payments/fail: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	log.Println("✅ Payments endpoints валидны")
	return nil
}

func (v *SpecValidator) makeRequest(method, path string, body interface{}) (*http.Response, error) {
	var req *http.Request
	var err error

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		
		req, err = http.NewRequest(method, v.baseURL+path, bytes.NewBuffer(jsonBody))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, v.baseURL+path, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	return resp, nil
}

// RunValidation запускает валидацию API
func RunValidation() {
	baseURL := "http://localhost:8081"
	
	validator := NewSpecValidator(baseURL)
	if err := validator.ValidateAll(); err != nil {
		log.Fatalf("❌ Валидация не пройдена: %v", err)
	}
}
