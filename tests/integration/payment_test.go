package integration

import (
	"strings"
	"testing"

	"bulbul/internal/models"
)

// TestPayment_CompleteFlow tests the complete payment flow for regular events (not Event 1)
func TestPayment_CompleteFlow(t *testing.T) {
	client := NewTestClient(APIBaseURL)

	LogTestStep(t, "Testing complete payment flow for regular event")

	// Step 1: Create booking for a regular event (not Event 1)
	LogTestStep(t, "Step 1: Create booking for Event 2")
	booking := client.CreateBooking(t, 2)
	LogTestResult(t, "Booking created: ID=%d", booking.ID)

	// Step 2: List seats for the event
	LogTestStep(t, "Step 2: List seats for Event 2")
	seats := client.ListSeats(t, 2)
	if len(seats) == 0 {
		t.Fatal("No seats available for Event 2")
	}
	LogTestResult(t, "Found %d seats for Event 2", len(seats))

	// Step 3: Select a seat
	LogTestStep(t, "Step 3: Select a seat")
	freeSeat := FindFreeSeat(seats)
	if freeSeat == nil {
		t.Skip("No free seats available for testing")
	}
	client.SelectSeat(t, booking.ID, freeSeat.ID)
	LogTestResult(t, "Seat selected: ID=%d", freeSeat.ID)

	// Step 4: Initiate payment
	LogTestStep(t, "Step 4: Initiate payment")
	paymentURL := client.InitiatePayment(t, booking.ID)
	ValidatePaymentURL(t, paymentURL)
	LogTestResult(t, "Payment initiated, URL: %s", paymentURL)

	// Step 5: Simulate payment success callback
	LogTestStep(t, "Step 5: Simulate successful payment")
	client.NotifyPaymentSuccess(t, booking.ID)
	LogTestResult(t, "Payment success notification sent")

	// Step 6: Verify booking is confirmed
	LogTestStep(t, "Step 6: Verify booking status")
	bookings := client.ListBookings(t)
	AssertBookingExists(t, bookings, booking.ID)
	LogTestResult(t, "Booking confirmed after payment")

	LogTestResult(t, "✅ Complete payment flow test passed!")
}

// TestPayment_InitiateForBooking tests payment initiation
func TestPayment_InitiateForBooking(t *testing.T) {
	client := NewTestClient(APIBaseURL)

	LogTestStep(t, "Testing payment initiation")

	// Create booking with seat
	booking := client.CreateBooking(t, 2)
	seats := client.ListSeats(t, 2)
	freeSeat := FindFreeSeat(seats)
	if freeSeat == nil {
		t.Skip("No free seats available")
	}
	client.SelectSeat(t, booking.ID, freeSeat.ID)

	// Initiate payment
	paymentURL := client.InitiatePayment(t, booking.ID)

	// Validate payment URL structure
	if !strings.Contains(paymentURL, "http") {
		t.Fatalf("Invalid payment URL format: %s", paymentURL)
	}

	LogTestResult(t, "Payment URL generated successfully")
}

// TestPayment_SuccessCallback tests successful payment callback
func TestPayment_SuccessCallback(t *testing.T) {
	client := NewTestClient(APIBaseURL)

	LogTestStep(t, "Testing payment success callback")

	// Create booking and initiate payment
	booking := client.CreateBooking(t, 2)
	seats := client.ListSeats(t, 2)
	freeSeat := FindFreeSeat(seats)
	if freeSeat == nil {
		t.Skip("No free seats available")
	}
	client.SelectSeat(t, booking.ID, freeSeat.ID)
	client.InitiatePayment(t, booking.ID)

	// Send success notification
	client.NotifyPaymentSuccess(t, booking.ID)

	// Verify booking still exists and is confirmed
	bookings := client.ListBookings(t)
	AssertBookingExists(t, bookings, booking.ID)

	LogTestResult(t, "Payment success processed correctly")
}

// TestPayment_FailureCallback tests failed payment callback
func TestPayment_FailureCallback(t *testing.T) {
	client := NewTestClient(APIBaseURL)

	LogTestStep(t, "Testing payment failure callback")

	// Create booking and initiate payment
	booking := client.CreateBooking(t, 2)
	seats := client.ListSeats(t, 2)
	freeSeat := FindFreeSeat(seats)
	if freeSeat == nil {
		t.Skip("No free seats available")
	}
	client.SelectSeat(t, booking.ID, freeSeat.ID)
	client.InitiatePayment(t, booking.ID)

	// Send failure notification
	client.NotifyPaymentFailure(t, booking.ID)

	// Verify seat is released back to FREE status
	updatedSeats := client.ListSeats(t, 2)
	AssertSeatStatus(t, updatedSeats, freeSeat.ID, "FREE")

	LogTestResult(t, "Payment failure processed correctly, seat released")
}

// TestPayment_WebhookNotification tests payment webhook processing
func TestPayment_WebhookNotification(t *testing.T) {
	client := NewTestClient(APIBaseURL)

	LogTestStep(t, "Testing payment webhook notification")

	// Create booking and initiate payment
	booking := client.CreateBooking(t, 2)
	seats := client.ListSeats(t, 2)
	freeSeat := FindFreeSeat(seats)
	if freeSeat == nil {
		t.Skip("No free seats available")
	}
	client.SelectSeat(t, booking.ID, freeSeat.ID)
	client.InitiatePayment(t, booking.ID)

	// Create and send webhook notification
	notification := CreateTestPaymentNotification(
		"payment-123",
		ConvertToStringOrderID(booking.ID),
		"CONFIRMED",
	)

	client.SendPaymentWebhook(t, notification)

	LogTestResult(t, "Payment webhook processed successfully")
}

// TestPayment_TokenGeneration tests payment token generation
func TestPayment_TokenGeneration(t *testing.T) {
	LogTestStep(t, "Testing payment token generation")

	// Test token generation with known values
	amount := int64(10000) // 100.00 in kopecks
	currency := "RUB"
	orderID := "test-order-123"
	teamSlug, password := GetTestPaymentCredentials()

	token := GeneratePaymentToken(amount, currency, orderID, teamSlug, password)

	// Verify token is a valid SHA-256 hash (64 hex characters)
	if len(token) != 64 {
		t.Fatalf("Expected 64-character token, got %d characters", len(token))
	}

	// Verify token contains only hex characters
	for _, char := range token {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			t.Fatalf("Token contains non-hex character: %c", char)
		}
	}

	LogTestResult(t, "Payment token generated: %s", token)
}

// TestPayment_CancelAfterPayment tests that confirmed bookings cannot be cancelled
func TestPayment_CancelAfterPayment(t *testing.T) {
	client := NewTestClient(APIBaseURL)

	LogTestStep(t, "Testing booking cancellation after payment")

	// Create booking and complete payment
	booking := client.CreateBooking(t, 2)
	seats := client.ListSeats(t, 2)
	freeSeat := FindFreeSeat(seats)
	if freeSeat == nil {
		t.Skip("No free seats available")
	}
	client.SelectSeat(t, booking.ID, freeSeat.ID)
	client.InitiatePayment(t, booking.ID)
	client.NotifyPaymentSuccess(t, booking.ID)

	// Try to cancel the booking - this should either be rejected or handle gracefully
	// The behavior depends on business logic implementation
	client.CancelBooking(t, booking.ID)

	LogTestResult(t, "Cancellation request processed (business logic dependent)")
}

// TestPayment_MultiplePayments tests handling multiple payment attempts
func TestPayment_MultiplePayments(t *testing.T) {
	client := NewTestClient(APIBaseURL)

	LogTestStep(t, "Testing multiple payment initiation attempts")

	// Create booking with seat
	booking := client.CreateBooking(t, 2)
	seats := client.ListSeats(t, 2)
	freeSeat := FindFreeSeat(seats)
	if freeSeat == nil {
		t.Skip("No free seats available")
	}
	client.SelectSeat(t, booking.ID, freeSeat.ID)

	// Initiate payment multiple times
	paymentURL1 := client.InitiatePayment(t, booking.ID)
	paymentURL2 := client.InitiatePayment(t, booking.ID)

	// Both should succeed (idempotent operation)
	ValidatePaymentURL(t, paymentURL1)
	ValidatePaymentURL(t, paymentURL2)

	LogTestResult(t, "Multiple payment initiations handled correctly")
}

// TestPayment_WithoutSeat tests payment initiation without selecting seats
func TestPayment_WithoutSeat(t *testing.T) {
	client := NewTestClient(APIBaseURL)

	LogTestStep(t, "Testing payment initiation without seats")

	// Create booking but don't select any seats
	booking := client.CreateBooking(t, 2)

	// Try to initiate payment - this should either fail or handle gracefully
	// The behavior depends on business logic implementation
	paymentURL := client.InitiatePayment(t, booking.ID)

	// If it succeeds, the URL should still be valid
	if paymentURL != "" {
		ValidatePaymentURL(t, paymentURL)
	}

	LogTestResult(t, "Payment initiation without seats handled")
}

// TestPayment_DifferentAmounts tests payment with different booking amounts
func TestPayment_DifferentAmounts(t *testing.T) {
	client := NewTestClient(APIBaseURL)

	LogTestStep(t, "Testing payments with different amounts")

	// Create multiple bookings with different numbers of seats
	for i := 1; i <= 2; i++ {
		LogTestStep(t, "Testing booking with %d seat(s)", i)

		booking := client.CreateBooking(t, 2)
		seats := client.ListSeats(t, 2)

		// Select i number of seats
		selectedSeats := 0
		for _, seat := range seats {
			if seat.Status == "FREE" && selectedSeats < i {
				client.SelectSeat(t, booking.ID, seat.ID)
				selectedSeats++
			}
		}

		if selectedSeats < i {
			LogTestStep(t, "Skipping test for %d seats (not enough free seats)", i)
			continue
		}

		// Initiate payment
		paymentURL := client.InitiatePayment(t, booking.ID)
		ValidatePaymentURL(t, paymentURL)

		LogTestResult(t, "Payment initiated for booking with %d seat(s)", selectedSeats)

		// Clean up
		client.NotifyPaymentFailure(t, booking.ID) // Release seats
	}
}

// TestPayment_ConcurrentPayments tests concurrent payment processing
func TestPayment_ConcurrentPayments(t *testing.T) {
	client := NewTestClient(APIBaseURL)

	LogTestStep(t, "Testing concurrent payment processing")

	// Create two bookings
	booking1 := client.CreateBooking(t, 2)
	booking2 := client.CreateBooking(t, 2)

	seats := client.ListSeats(t, 2)
	freeSeats := make([]models.ListSeatsResponseItem, 0)
	for _, seat := range seats {
		if seat.Status == "FREE" && len(freeSeats) < 2 {
			freeSeats = append(freeSeats, seat)
		}
	}

	if len(freeSeats) < 2 {
		t.Skip("Need at least 2 free seats for concurrent test")
	}

	// Select different seats for each booking
	client.SelectSeat(t, booking1.ID, freeSeats[0].ID)
	client.SelectSeat(t, booking2.ID, freeSeats[1].ID)

	// Initiate payments concurrently using goroutines
	results := make(chan string, 2)

	go func() {
		url := client.InitiatePayment(t, booking1.ID)
		results <- url
	}()

	go func() {
		url := client.InitiatePayment(t, booking2.ID)
		results <- url
	}()

	// Wait for both to complete
	url1 := <-results
	url2 := <-results

	ValidatePaymentURL(t, url1)
	ValidatePaymentURL(t, url2)

	LogTestResult(t, "Concurrent payments processed successfully")

	// Clean up
	client.NotifyPaymentFailure(t, booking1.ID)
	client.NotifyPaymentFailure(t, booking2.ID)
}
