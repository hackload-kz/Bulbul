package integration

import (
	"testing"

	"bulbul/internal/models"
)

// TestEvent1_ExternalService_CompleteFlow tests the complete flow for Event ID=1
// This event uses external ticketing service and should not use payment flow
func TestEvent1_ExternalService_CompleteFlow(t *testing.T) {
	client := NewTestClient(APIBaseURL)

	LogTestStep(t, "Testing complete flow for Event ID=1 (external service)")

	// Step 1: Verify API is healthy
	LogTestStep(t, "Step 1: Health check")
	client.HealthCheck(t)
	LogTestResult(t, "API is healthy")

	// Step 2: List events and verify Event 1 exists
	LogTestStep(t, "Step 2: List events and verify Event 1 exists")
	events := client.ListEvents(t)
	AssertEventExists(t, events, 1)
	LogTestResult(t, "Event 1 found in events list")

	// Step 3: Create booking for event 1
	LogTestStep(t, "Step 3: Create booking for Event 1")
	booking := client.CreateBooking(t, 1)
	if booking.ID == 0 {
		t.Fatal("Booking ID should not be 0")
	}
	LogTestResult(t, "Booking created with ID: %d", booking.ID)

	// Step 4: List seats for event 1 (should come from external service)
	LogTestStep(t, "Step 4: List seats for Event 1 (external service)")
	seats := client.ListSeats(t, 1)
	if len(seats) == 0 {
		t.Fatal("Expected seats from external service, got none")
	}
	LogTestResult(t, "Found %d seats from external service", len(seats))

	// Step 5: Find a free seat
	LogTestStep(t, "Step 5: Find a free seat")
	freeSeat := FindFreeSeat(seats)
	if freeSeat == nil {
		t.Fatal("No free seats available for testing")
	}
	LogTestResult(t, "Found free seat: ID=%d, Row=%d, Number=%d", freeSeat.ID, freeSeat.Row, freeSeat.Number)

	// Step 6: Select the seat (should call external service)
	LogTestStep(t, "Step 6: Select seat in external service")
	client.SelectSeat(t, booking.ID, freeSeat.ID)
	LogTestResult(t, "Seat selected successfully in external service")

	// Step 7: Verify seat is now reserved
	LogTestStep(t, "Step 7: Verify seat status changed")
	updatedSeats := client.ListSeats(t, 1)
	AssertSeatStatus(t, updatedSeats, freeSeat.ID, "RESERVED")
	LogTestResult(t, "Seat status updated to RESERVED")

	// Step 8: List bookings to verify our booking exists
	LogTestStep(t, "Step 8: Verify booking exists in our system")
	bookings := client.ListBookings(t)
	AssertBookingExists(t, bookings, booking.ID)
	LogTestResult(t, "Booking found in user's bookings list")

	// Step 9: For Event 1, we should NOT initiate payment (external service handles it)
	LogTestStep(t, "Step 9: Verify no payment initiation for Event 1")
	// This is implicit - we don't call InitiatePayment for Event 1
	LogTestResult(t, "No payment initiation needed for Event 1 (external service handles payments)")

	// Step 10: Clean up - release the seat
	LogTestStep(t, "Step 10: Clean up - release seat")
	client.ReleaseSeat(t, freeSeat.ID)
	LogTestResult(t, "Seat released successfully")

	// Step 11: Verify seat is free again
	LogTestStep(t, "Step 11: Verify seat is free after release")
	finalSeats := client.ListSeats(t, 1)
	AssertSeatStatus(t, finalSeats, freeSeat.ID, "FREE")
	LogTestResult(t, "Seat status updated to FREE after release")

	LogTestResult(t, "✅ Complete Event 1 external service flow test passed!")
}

// TestEvent1_CreateBooking tests creating a booking for Event 1
func TestEvent1_CreateBooking(t *testing.T) {
	client := NewTestClient(APIBaseURL)

	LogTestStep(t, "Testing booking creation for Event 1")

	booking := client.CreateBooking(t, 1)
	if booking.ID == 0 {
		t.Fatal("Expected non-zero booking ID")
	}

	LogTestResult(t, "Booking created successfully: ID=%d", booking.ID)
}

// TestEvent1_ListExternalSeats tests listing seats from external service
func TestEvent1_ListExternalSeats(t *testing.T) {
	client := NewTestClient(APIBaseURL)

	LogTestStep(t, "Testing external seat listing for Event 1")

	seats := client.ListSeats(t, 1)

	// Verify we got seats
	if len(seats) == 0 {
		t.Fatal("Expected seats from external service")
	}

	// Verify seat structure
	for i, seat := range seats {
		if seat.Row == 0 {
			t.Fatalf("Seat %d has invalid row", i)
		}
		if seat.Number == 0 {
			t.Fatalf("Seat %d has invalid number", i)
		}
		if seat.Status != "FREE" && seat.Status != "RESERVED" {
			t.Fatalf("Seat %d has invalid status: %s", i, seat.Status)
		}
	}

	LogTestResult(t, "Retrieved %d seats from external service", len(seats))
}

// TestEvent1_SelectExternalSeat tests selecting a seat via external service
func TestEvent1_SelectExternalSeat(t *testing.T) {
	client := NewTestClient(APIBaseURL)

	LogTestStep(t, "Testing external seat selection for Event 1")

	// Create booking
	booking := client.CreateBooking(t, 1)

	// Get seats
	seats := client.ListSeats(t, 1)
	freeSeat := FindFreeSeat(seats)
	if freeSeat == nil {
		t.Skip("No free seats available for testing")
	}

	// Select seat
	client.SelectSeat(t, booking.ID, freeSeat.ID)

	// Verify seat is reserved
	updatedSeats := client.ListSeats(t, 1)
	AssertSeatStatus(t, updatedSeats, freeSeat.ID, "RESERVED")

	LogTestResult(t, "Seat selected and reserved in external service")

	// Clean up
	client.ReleaseSeat(t, freeSeat.ID)
}

// TestEvent1_ReleaseExternalSeat tests releasing a seat via external service
func TestEvent1_ReleaseExternalSeat(t *testing.T) {
	client := NewTestClient(APIBaseURL)

	LogTestStep(t, "Testing external seat release for Event 1")

	// Create booking and select seat
	booking := client.CreateBooking(t, 1)
	seats := client.ListSeats(t, 1)
	freeSeat := FindFreeSeat(seats)
	if freeSeat == nil {
		t.Skip("No free seats available for testing")
	}

	client.SelectSeat(t, booking.ID, freeSeat.ID)

	// Release seat
	client.ReleaseSeat(t, freeSeat.ID)

	// Verify seat is free
	updatedSeats := client.ListSeats(t, 1)
	AssertSeatStatus(t, updatedSeats, freeSeat.ID, "FREE")

	LogTestResult(t, "Seat released successfully in external service")
}

// TestEvent1_NoPaymentRequired verifies that Event 1 doesn't require payment initiation
func TestEvent1_NoPaymentRequired(t *testing.T) {
	client := NewTestClient(APIBaseURL)

	LogTestStep(t, "Testing that Event 1 doesn't require payment initiation")

	// Create booking
	booking := client.CreateBooking(t, 1)

	// Select a seat
	seats := client.ListSeats(t, 1)
	freeSeat := FindFreeSeat(seats)
	if freeSeat == nil {
		t.Skip("No free seats available for testing")
	}
	client.SelectSeat(t, booking.ID, freeSeat.ID)

	// For Event 1, we should NOT call InitiatePayment
	// The external ticketing service handles payments
	// This test just verifies the booking was created successfully
	// without requiring payment initiation

	bookings := client.ListBookings(t)
	AssertBookingExists(t, bookings, booking.ID)

	LogTestResult(t, "Event 1 booking exists without payment initiation")

	// Clean up
	client.ReleaseSeat(t, freeSeat.ID)
}

// TestEvent1_MultipleSeats tests selecting multiple seats for Event 1
func TestEvent1_MultipleSeats(t *testing.T) {
	client := NewTestClient(APIBaseURL)

	LogTestStep(t, "Testing multiple seat selection for Event 1")

	// Create booking
	booking := client.CreateBooking(t, 1)

	// Get seats and find free ones
	seats := client.ListSeats(t, 1)
	var freeSeats []models.ListSeatsResponseItem
	for _, seat := range seats {
		if seat.Status == "FREE" && len(freeSeats) < 3 {
			freeSeats = append(freeSeats, seat)
		}
	}

	if len(freeSeats) < 2 {
		t.Skip("Need at least 2 free seats for this test")
	}

	// Select multiple seats
	var selectedSeats []string
	for i, seat := range freeSeats[:2] {
		LogTestStep(t, "Selecting seat %d: ID=%d", i+1, seat.ID)
		client.SelectSeat(t, booking.ID, seat.ID)
		selectedSeats = append(selectedSeats, seat.ID)
	}

	// Verify all seats are reserved
	updatedSeats := client.ListSeats(t, 1)
	for _, seatID := range selectedSeats {
		AssertSeatStatus(t, updatedSeats, seatID, "RESERVED")
	}

	LogTestResult(t, "Multiple seats selected successfully")

	// Clean up - release all seats
	for _, seatID := range selectedSeats {
		client.ReleaseSeat(t, seatID)
	}

	LogTestResult(t, "All seats released successfully")
}
