package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"bulbul/internal/models"
)

// TestClient provides methods for testing the API
type TestClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewTestClient creates a new test client
func NewTestClient(baseURL string) *TestClient {
	return &TestClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// makeRequest makes an HTTP request and returns the response
func (c *TestClient) makeRequest(t *testing.T, method, path string, body interface{}) *http.Response {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	return resp
}

// ListEvents lists all events
func (c *TestClient) ListEvents(t *testing.T) []models.ListEventsResponseItem {
	resp := c.makeRequest(t, "GET", "/api/events", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var events []models.ListEventsResponseItem
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		t.Fatalf("Failed to decode events response: %v", err)
	}

	return events
}

// CreateBooking creates a new booking
func (c *TestClient) CreateBooking(t *testing.T, eventID int64) *models.CreateBookingResponse {
	req := models.CreateBookingRequest{
		EventID: eventID,
	}

	resp := c.makeRequest(t, "POST", "/api/bookings", req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 201, got %d. Body: %s", resp.StatusCode, string(body))
	}

	var booking models.CreateBookingResponse
	if err := json.NewDecoder(resp.Body).Decode(&booking); err != nil {
		t.Fatalf("Failed to decode booking response: %v", err)
	}

	return &booking
}

// ListSeats lists seats for an event
func (c *TestClient) ListSeats(t *testing.T, eventID int64) []models.ListSeatsResponseItem {
	path := fmt.Sprintf("/api/seats?event_id=%d&page=1&pageSize=20", eventID)
	resp := c.makeRequest(t, "GET", path, nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	var seats []models.ListSeatsResponseItem
	if err := json.NewDecoder(resp.Body).Decode(&seats); err != nil {
		t.Fatalf("Failed to decode seats response: %v", err)
	}

	return seats
}

// SelectSeat selects a seat for a booking
func (c *TestClient) SelectSeat(t *testing.T, bookingID int64, seatID string) {
	req := models.SelectSeatRequest{
		BookingID: bookingID,
		SeatID:    seatID,
	}

	resp := c.makeRequest(t, "PATCH", "/api/seats/select", req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}
}

// ReleaseSeat releases a seat
func (c *TestClient) ReleaseSeat(t *testing.T, seatID string) {
	req := models.ReleaseSeatRequest{
		SeatID: seatID,
	}

	resp := c.makeRequest(t, "PATCH", "/api/seats/release", req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}
}

// ListBookings lists bookings for a user
func (c *TestClient) ListBookings(t *testing.T) []models.ListBookingsResponseItem {
	resp := c.makeRequest(t, "GET", "/api/bookings", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	var bookings []models.ListBookingsResponseItem
	if err := json.NewDecoder(resp.Body).Decode(&bookings); err != nil {
		t.Fatalf("Failed to decode bookings response: %v", err)
	}

	return bookings
}

// InitiatePayment initiates payment for a booking
func (c *TestClient) InitiatePayment(t *testing.T, bookingID int64) string {
	req := models.InitiatePaymentRequest{
		BookingID: bookingID,
	}

	resp := c.makeRequest(t, "PATCH", "/api/bookings/initiatePayment", req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 302, got %d. Body: %s", resp.StatusCode, string(body))
	}

	// Get the Location header for payment URL
	location := resp.Header.Get("Location")
	if location == "" {
		t.Fatal("Expected Location header in payment response")
	}

	return location
}

// CancelBooking cancels a booking
func (c *TestClient) CancelBooking(t *testing.T, bookingID int64) {
	req := models.CancelBookingRequest{
		BookingID: bookingID,
	}

	resp := c.makeRequest(t, "PATCH", "/api/bookings/cancel", req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}
}

// NotifyPaymentSuccess simulates a successful payment notification
func (c *TestClient) NotifyPaymentSuccess(t *testing.T, orderID int64) {
	path := fmt.Sprintf("/api/payments/success?orderId=%d", orderID)
	resp := c.makeRequest(t, "GET", path, nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}
}

// NotifyPaymentFailure simulates a failed payment notification
func (c *TestClient) NotifyPaymentFailure(t *testing.T, orderID int64) {
	path := fmt.Sprintf("/api/payments/fail?orderId=%d", orderID)
	resp := c.makeRequest(t, "GET", path, nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}
}

// SendPaymentWebhook sends a payment webhook notification
func (c *TestClient) SendPaymentWebhook(t *testing.T, notification models.PaymentNotificationPayload) {
	resp := c.makeRequest(t, "POST", "/api/payments/notifications", notification)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}
}

// HealthCheck checks if the API is healthy
func (c *TestClient) HealthCheck(t *testing.T) {
	resp := c.makeRequest(t, "GET", "/health", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Health check failed with status %d", resp.StatusCode)
	}
}