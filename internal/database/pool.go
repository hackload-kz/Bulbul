package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

type PoolStats struct {
	MaxOpenConns    int `json:"max_open_connections"`
	OpenConns       int `json:"open_connections"`
	InUse           int `json:"in_use"`
	Idle            int `json:"idle"`
	WaitCount       int64 `json:"wait_count"`
	WaitDuration    time.Duration `json:"wait_duration"`
	MaxIdleClosed   int64 `json:"max_idle_closed"`
	MaxLifetimeClosed int64 `json:"max_lifetime_closed"`
}

type HealthCheck struct {
	Status       string        `json:"status"`
	ResponseTime time.Duration `json:"response_time"`
	Error        string        `json:"error,omitempty"`
	Stats        PoolStats     `json:"stats"`
	Timestamp    time.Time     `json:"timestamp"`
}

func (db *DB) GetPoolStats() PoolStats {
	stats := db.Stats()
	return PoolStats{
		MaxOpenConns:      stats.MaxOpenConnections,
		OpenConns:         stats.OpenConnections,
		InUse:             stats.InUse,
		Idle:              stats.Idle,
		WaitCount:         stats.WaitCount,
		WaitDuration:      stats.WaitDuration,
		MaxIdleClosed:     stats.MaxIdleClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
	}
}

func (db *DB) HealthCheck(ctx context.Context) HealthCheck {
	start := time.Now()
	healthCheck := HealthCheck{
		Timestamp: start,
		Stats:     db.GetPoolStats(),
	}

	// Perform database ping with timeout
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := db.PingContext(pingCtx)
	healthCheck.ResponseTime = time.Since(start)

	if err != nil {
		healthCheck.Status = "unhealthy"
		healthCheck.Error = err.Error()
		slog.Error("Database health check failed", "error", err)
	} else {
		healthCheck.Status = "healthy"
	}

	return healthCheck
}

func (db *DB) ValidateConnectionPool() error {
	stats := db.Stats()
	
	// Check for potential connection leaks
	if stats.InUse > int(float64(stats.MaxOpenConnections)*0.9) {
		slog.Warn("High connection usage detected", 
			"in_use", stats.InUse, "max_open", stats.MaxOpenConnections)
	}

	// Check for high wait times
	if stats.WaitCount > 0 && stats.WaitDuration > time.Second {
		slog.Warn("High database wait times detected", 
			"wait_count", stats.WaitCount, "wait_duration", stats.WaitDuration)
	}

	// Check for excessive idle connection closures
	if stats.MaxIdleClosed > 1000 {
		slog.Info("Many idle connections have been closed - consider adjusting MaxIdleConns", 
			"closed", stats.MaxIdleClosed)
	}

	return nil
}

func (db *DB) ExecuteWithRetry(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	const maxRetries = 3
	const backoffDelay = 100 * time.Millisecond

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		rows, err := db.QueryContext(ctx, query, args...)
		if err == nil {
			return rows, nil
		}

		lastErr = err
		
		// Check if error is retryable (connection issues)
		if !isRetryableError(err) {
			return nil, fmt.Errorf("non-retryable error on attempt %d: %w", attempt, err)
		}

		if attempt < maxRetries {
			slog.Warn("Database query failed, retrying", 
				"attempt", attempt, "max_retries", maxRetries, "error", err)
			time.Sleep(time.Duration(attempt) * backoffDelay)
		}
	}

	return nil, fmt.Errorf("query failed after %d attempts: %w", maxRetries, lastErr)
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for connection-related errors that might be temporary
	errStr := err.Error()
	retryableErrors := []string{
		"connection refused",
		"connection reset",
		"broken pipe",
		"timeout",
		"driver: bad connection",
	}
	
	for _, retryable := range retryableErrors {
		if contains(errStr, retryable) {
			return true
		}
	}
	
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
				indexOf(s, substr) != -1)))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}