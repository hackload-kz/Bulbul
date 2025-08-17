package external

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type TicketingClient struct {
	baseURL    string
	httpClient *http.Client
}

type TicketingConfig struct {
	BaseURL string
	Timeout time.Duration
}

// External ticketing service models based on ticketing_service_provider.md
type StartOrderResponse struct {
	OrderID string `json:"order_id"`
}

type GetOrderResponse struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	StartedAt   int64  `json:"started_at"`
	UpdatedAt   int64  `json:"updated_at"`
	PlacesCount int    `json:"places_count"`
}

type Place struct {
	ID     string `json:"id"`
	Row    int    `json:"row"`
	Seat   int    `json:"seat"`
	IsFree bool   `json:"is_free"`
}

type SelectPlaceRequest struct {
	OrderID string `json:"order_id"`
}

func NewTicketingClient(cfg TicketingConfig) *TicketingClient {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &TicketingClient{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

func (tc *TicketingClient) StartOrder() (*StartOrderResponse, error) {
	resp, err := tc.httpClient.Post(tc.baseURL+"/api/partners/v1/orders", "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start order: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result StartOrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (tc *TicketingClient) GetOrder(orderID string) (*GetOrderResponse, error) {
	resp, err := tc.httpClient.Get(tc.baseURL + "/api/partners/v1/orders/" + orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result GetOrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (tc *TicketingClient) GetPlaces(page, pageSize int) ([]Place, error) {
	url := fmt.Sprintf("%s/api/partners/v1/places?page=%d&pageSize=%d", tc.baseURL, page, pageSize)
	resp, err := tc.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get places: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var places []Place
	if err := json.NewDecoder(resp.Body).Decode(&places); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return places, nil
}

func (tc *TicketingClient) SelectPlace(placeID string, orderID string) error {
	reqBody := SelectPlaceRequest{OrderID: orderID}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("PATCH", tc.baseURL+"/api/partners/v1/places/"+placeID+"/select", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := tc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to select place: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (tc *TicketingClient) ReleasePlace(placeID string) error {
	req, err := http.NewRequest("PATCH", tc.baseURL+"/api/partners/v1/places/"+placeID+"/release", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := tc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to release place: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (tc *TicketingClient) SubmitOrder(orderID string) error {
	req, err := http.NewRequest("PATCH", tc.baseURL+"/api/partners/v1/orders/"+orderID+"/submit", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := tc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to submit order: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (tc *TicketingClient) ConfirmOrder(orderID string) error {
	req, err := http.NewRequest("PATCH", tc.baseURL+"/api/partners/v1/orders/"+orderID+"/confirm", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := tc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to confirm order: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (tc *TicketingClient) CancelOrder(orderID string) error {
	req, err := http.NewRequest("PATCH", tc.baseURL+"/api/partners/v1/orders/"+orderID+"/cancel", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := tc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
