package external

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"
)

type PaymentClient struct {
	baseURL    string
	teamSlug   string
	password   string
	httpClient *http.Client
}

type PaymentConfig struct {
	BaseURL  string
	TeamSlug string
	Password string
	Timeout  time.Duration
}

// Payment gateway models based on payment_gateway.md
type PaymentInitRequest struct {
	TeamSlug        string `json:"teamSlug"`
	Token           string `json:"token"`
	Amount          int64  `json:"amount"`
	OrderID         string `json:"orderId"`
	Currency        string `json:"currency"`
	Description     string `json:"description,omitempty"`
	Email           string `json:"email,omitempty"`
	SuccessURL      string `json:"successURL,omitempty"`
	FailURL         string `json:"failURL,omitempty"`
	NotificationURL string `json:"notificationURL,omitempty"`
	Language        string `json:"language,omitempty"`
}

type PaymentInitResponse struct {
	Success    bool   `json:"success"`
	PaymentID  string `json:"paymentId"`
	OrderID    string `json:"orderId"`
	Status     string `json:"status"`
	Amount     int64  `json:"amount"`
	Currency   string `json:"currency"`
	PaymentURL string `json:"paymentURL"`
	ExpiresAt  string `json:"expiresAt"`
	CreatedAt  string `json:"createdAt"`
}

type PaymentCheckRequest struct {
	TeamSlug  string `json:"teamSlug"`
	Token     string `json:"token"`
	PaymentID string `json:"paymentId,omitempty"`
	OrderID   string `json:"orderId,omitempty"`
}

type PaymentCheckResponse struct {
	Success      bool                `json:"success"`
	Payments     []PaymentDetails    `json:"payments"`
	TotalCount   int                 `json:"totalCount"`
	OrderID      string              `json:"orderId"`
}

type PaymentDetails struct {
	PaymentID         string `json:"paymentId"`
	OrderID           string `json:"orderId"`
	Status            string `json:"status"`
	StatusDescription string `json:"statusDescription"`
	Amount            int64  `json:"amount"`
	Currency          string `json:"currency"`
	CreatedAt         string `json:"createdAt"`
	UpdatedAt         string `json:"updatedAt"`
	ExpiresAt         string `json:"expiresAt"`
	Description       string `json:"description"`
}

func NewPaymentClient(cfg PaymentConfig) *PaymentClient {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &PaymentClient{
		baseURL:  cfg.BaseURL,
		teamSlug: cfg.TeamSlug,
		password: cfg.Password,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

func (pc *PaymentClient) generateToken(params map[string]string) string {
	// Add required parameters
	params["TeamSlug"] = pc.teamSlug
	params["Password"] = pc.password

	// Sort parameters alphabetically
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Concatenate values
	var tokenString string
	for _, key := range keys {
		tokenString += params[key]
	}

	// Generate SHA-256 hash
	hash := sha256.Sum256([]byte(tokenString))
	return hex.EncodeToString(hash[:])
}

func (pc *PaymentClient) InitPayment(amount int64, orderID, currency, description string) (*PaymentInitResponse, error) {
	// Generate token using simplified scheme (5 parameters only)
	params := map[string]string{
		"Amount":   strconv.FormatInt(amount, 10),
		"Currency": currency,
		"OrderId":  orderID,
	}
	token := pc.generateToken(params)

	req := PaymentInitRequest{
		TeamSlug:    pc.teamSlug,
		Token:       token,
		Amount:      amount,
		OrderID:     orderID,
		Currency:    currency,
		Description: description,
		Language:    "ru",
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := pc.httpClient.Post(pc.baseURL+"/api/v1/PaymentInit/init", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to init payment: %w", err)
	}
	defer resp.Body.Close()

	var result PaymentInitResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("payment init failed")
	}

	return &result, nil
}

func (pc *PaymentClient) CheckPayment(paymentID string) (*PaymentCheckResponse, error) {
	// Generate token for check request
	params := map[string]string{
		"PaymentId": paymentID,
	}
	token := pc.generateToken(params)

	req := PaymentCheckRequest{
		TeamSlug:  pc.teamSlug,
		Token:     token,
		PaymentID: paymentID,
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := pc.httpClient.Post(pc.baseURL+"/api/v1/PaymentCheck/check", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to check payment: %w", err)
	}
	defer resp.Body.Close()

	var result PaymentCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (pc *PaymentClient) ConfirmPayment(paymentID string, amount int64) error {
	// Generate token for confirm request
	params := map[string]string{
		"Amount":    strconv.FormatInt(amount, 10),
		"PaymentId": paymentID,
	}
	token := pc.generateToken(params)

	reqData := map[string]interface{}{
		"teamSlug":  pc.teamSlug,
		"token":     token,
		"paymentId": paymentID,
		"amount":    amount,
	}

	jsonBody, err := json.Marshal(reqData)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := pc.httpClient.Post(pc.baseURL+"/api/v1/PaymentConfirm/confirm", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to confirm payment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (pc *PaymentClient) CancelPayment(paymentID string, reason string) error {
	// Generate token for cancel request
	params := map[string]string{
		"PaymentId": paymentID,
	}
	token := pc.generateToken(params)

	reqData := map[string]interface{}{
		"teamSlug":  pc.teamSlug,
		"token":     token,
		"paymentId": paymentID,
		"reason":    reason,
	}

	jsonBody, err := json.Marshal(reqData)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := pc.httpClient.Post(pc.baseURL+"/api/v1/PaymentCancel/cancel", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to cancel payment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}