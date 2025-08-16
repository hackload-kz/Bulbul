# Integration Tests

This directory contains comprehensive integration tests for the Biletter API that make real calls to external services.

## Overview

The tests are designed to run against a live API server and make actual HTTP calls to:
- Local API server (localhost:8081)
- External ticketing service (https://hub.hackload.kz/event-provider/common)
- External payment service (https://hub.hackload.kz/payment-provider/common)

## Test Structure

### Core Files

- **`client.go`** - HTTP client with helper methods for all API endpoints
- **`helpers.go`** - Utility functions for test assertions and data generation
- **`README.md`** - This documentation

### Test Suites

- **`event1_test.go`** - Tests for Event ID=1 (external ticketing service)
- **`payment_test.go`** - Payment flow tests for regular events
- **`booking_test.go`** - Booking lifecycle and operations tests
- **`api_test.go`** - Complete API integration tests

## Prerequisites

1. **Services Running**: All services must be running locally:
   ```bash
   docker-compose up -d
   mise run load-data
   go run cmd/api/main.go
   ```

2. **Database Setup**: Database should be seeded with test data:
   - Event ID=1 (external service)
   - Event ID=2+ (regular events with seats)
   - User data

3. **External Service Access**: Tests will make real calls to:
   - Ticketing service for Event ID=1
   - Payment gateway for regular events

## Running Tests

### Run All Integration Tests
```bash
go test -v ./tests/integration/...
```

### Run Specific Test Suites
```bash
# Event 1 (external service) tests
go test -v ./tests/integration/ -run TestEvent1

# Payment flow tests
go test -v ./tests/integration/ -run TestPayment

# Booking operation tests
go test -v ./tests/integration/ -run TestBooking

# Complete API flow tests
go test -v ./tests/integration/ -run TestAPI
```

### Run Individual Tests
```bash
# Complete Event 1 flow
go test -v ./tests/integration/ -run TestEvent1_ExternalService_CompleteFlow

# Complete payment flow
go test -v ./tests/integration/ -run TestPayment_CompleteFlow

# API health check
go test -v ./tests/integration/ -run TestAPI_HealthCheck
```

### Run with Timeout
```bash
go test -v -timeout 5m ./tests/integration/...
```

## Test Coverage

### Event ID=1 (External Service)
- ✅ Complete external service flow
- ✅ Booking creation
- ✅ External seat listing
- ✅ External seat selection/release
- ✅ No payment requirement verification
- ✅ Multiple seat operations
- ✅ Cleanup and state verification

### Payment Flow (Regular Events)
- ✅ Complete payment cycle
- ✅ Payment initiation
- ✅ Success/failure callbacks
- ✅ Webhook notifications
- ✅ Token generation (SHA-256)
- ✅ Multiple payment scenarios
- ✅ Concurrent payment handling
- ✅ Error scenarios

### Booking Operations
- ✅ Booking creation for all event types
- ✅ Multiple seat selection
- ✅ Booking cancellation with seat release
- ✅ User booking listing
- ✅ Concurrent booking scenarios
- ✅ Invalid operation handling
- ✅ Payment requirement verification

### API Integration
- ✅ Health check verification
- ✅ Event listing
- ✅ Complete flows for both event types
- ✅ Multi-event operations
- ✅ Error handling
- ✅ Booking lifecycle management

## Key Test Scenarios

### External Service Flow (Event 1)
1. Create booking for Event 1
2. List seats from external ticketing service
3. Select seat via external API
4. Verify seat reservation in external system
5. Verify no payment initiation needed
6. Release seat through external API
7. Verify cleanup

### Internal Payment Flow (Regular Events)
1. Create booking for regular event
2. List seats from local database
3. Select seat locally
4. Initiate payment with real gateway
5. Generate valid payment token
6. Handle payment webhook
7. Verify booking confirmation
8. Test failure scenarios

### Error Scenarios
- Invalid event/seat/booking IDs
- Concurrent seat selection conflicts
- Payment failures and recovery
- External service timeouts
- Invalid request formats

## Test Data Requirements

Tests expect the following data to be available:

1. **Event 1**: External event that uses ticketing service
2. **Event 2+**: Regular events with local seat management
3. **Seats**: Available seats for selection
4. **External Services**: Accessible ticketing and payment APIs

## Environment Configuration

Tests use these constants (can be overridden with environment variables):

- `APIBaseURL`: "http://localhost:8081"
- External service URLs from config
- Test timeout: 30 seconds per operation

## Debugging

Tests include extensive logging:
- 🔹 Test steps
- ✅ Successful operations
- ❌ Error conditions
- Request/response details

Example output:
```
=== RUN   TestEvent1_ExternalService_CompleteFlow
integration_test.go:15: 🔹 Testing complete flow for Event ID=1 (external service)
integration_test.go:18: 🔹 Step 1: Health check
integration_test.go:20: ✅ API is healthy
integration_test.go:23: 🔹 Step 2: List events and verify Event 1 exists
integration_test.go:26: ✅ Event 1 found in events list
...
```

## Known Limitations

1. **External Service Dependencies**: Tests depend on external services being available
2. **Data State**: Tests may affect shared test data
3. **Rate Limits**: External services may have rate limits
4. **Cleanup**: Some tests may leave test data that needs manual cleanup

## Troubleshooting

### Common Issues

1. **API Not Running**: Ensure API server is running on localhost:8081
2. **Database Not Seeded**: Run `mise run load-data` before testing
3. **External Service Unavailable**: Check network connectivity and service status
4. **Timeout Errors**: Increase test timeout or check service performance

### Debug Commands

```bash
# Check API health
curl http://localhost:8081/health

# Check if events exist
curl http://localhost:8081/api/events

# Check database connectivity
docker-compose logs postgres

# Check external service connectivity
curl https://hub.hackload.kz/event-provider/common/api/partners/v1/orders
```

## Contributing

When adding new tests:

1. Use the existing client methods
2. Add proper logging with LogTestStep/LogTestResult
3. Include cleanup for any test data created
4. Test both success and error scenarios
5. Update this README with new test descriptions