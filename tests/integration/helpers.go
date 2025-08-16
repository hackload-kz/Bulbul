package integration

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"testing"

	"bulbul/internal/models"
)

const (
	APIBaseURL = "http://localhost:8081"
)

// FindFreeSeat finds a free seat from the list
func FindFreeSeat(seats []models.ListSeatsResponseItem) *models.ListSeatsResponseItem {
	for _, seat := range seats {
		if seat.Status == "FREE" {
			return &seat
		}
	}
	return nil
}

// FindReservedSeat finds a reserved seat from the list
func FindReservedSeat(seats []models.ListSeatsResponseItem) *models.ListSeatsResponseItem {
	for _, seat := range seats {
		if seat.Status == "RESERVED" {
			return &seat
		}
	}
	return nil
}

// GeneratePaymentToken generates SHA-256 token for payment gateway
func GeneratePaymentToken(amount int64, currency, orderID, teamSlug, password string) string {
	// According to payment_gateway.md: Amount + Currency + OrderId + Password + TeamSlug
	data := fmt.Sprintf("%d%s%s%s%s", amount, currency, orderID, password, teamSlug)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// AssertEventExists checks if an event exists in the list
func AssertEventExists(t *testing.T, events []models.ListEventsResponseItem, eventID int64) {
	for _, event := range events {
		if event.ID == eventID {
			return
		}
	}
	t.Fatalf("Event with ID %d not found in events list, %+v", eventID, events)
}

// AssertEventExistsWithTitle checks if an event with specific title exists
func AssertEventExistsWithTitle(t *testing.T, events []models.ListEventsResponseItem, eventID int64, expectedTitle string) {
	for _, event := range events {
		if event.ID == eventID {
			if event.Title != expectedTitle {
				t.Fatalf("Event %d has title '%s', expected '%s'", eventID, event.Title, expectedTitle)
			}
			return
		}
	}
	t.Fatalf("Event with ID %d not found in events list", eventID)
}

// AssertSeatStatus verifies that a seat has the expected status
func AssertSeatStatus(t *testing.T, seats []models.ListSeatsResponseItem, seatID string, expectedStatus string) {
	for _, seat := range seats {
		if seat.ID == seatID {
			if seat.Status != expectedStatus {
				t.Fatalf("Seat %s has status '%s', expected '%s'", seatID, seat.Status, expectedStatus)
			}
			return
		}
	}
	t.Fatalf("Seat with ID %s not found in seats list", seatID)
}

// AssertBookingExists checks if a booking exists in the list
func AssertBookingExists(t *testing.T, bookings []models.ListBookingsResponseItem, bookingID int64) {
	for _, booking := range bookings {
		if booking.ID == bookingID {
			return
		}
	}
	t.Fatalf("Booking with ID %d not found in bookings list", bookingID)
}

// GetTestPaymentCredentials returns test payment credentials
func GetTestPaymentCredentials() (teamSlug, password string) {
	// These should be set via environment variables in real tests
	// For now, using placeholder values
	return "test-team", "test-password"
}

// CreateTestPaymentNotification creates a test payment notification
func CreateTestPaymentNotification(paymentID, orderID, status string) models.PaymentNotificationPayload {
	return models.PaymentNotificationPayload{
		PaymentID: paymentID,
		Status:    status,
		TeamSlug:  "test-team",
		Timestamp: "2025-08-16T12:00:00Z",
		Data: map[string]interface{}{
			"orderId": orderID,
		},
	}
}

// LogTestStep logs a test step for better debugging
func LogTestStep(t *testing.T, step string, args ...interface{}) {
	t.Logf("🔹 "+step, args...)
}

// LogTestResult logs a test result
func LogTestResult(t *testing.T, result string, args ...interface{}) {
	t.Logf("✅ "+result, args...)
}

// LogTestError logs a test error
func LogTestError(t *testing.T, err string, args ...interface{}) {
	t.Logf("❌ "+err, args...)
}

// ConvertToStringOrderID converts int64 booking ID to string for external services
func ConvertToStringOrderID(bookingID int64) string {
	return strconv.FormatInt(bookingID, 10)
}

// ValidatePaymentURL validates that a payment URL is well-formed
func ValidatePaymentURL(t *testing.T, paymentURL string) {
	if paymentURL == "" {
		t.Fatal("Payment URL is empty")
	}
	// Should contain payment gateway base URL
	if len(paymentURL) < 10 {
		t.Fatalf("Payment URL seems too short: %s", paymentURL)
	}
	LogTestResult(t, "Payment URL generated: %s", paymentURL)
}
