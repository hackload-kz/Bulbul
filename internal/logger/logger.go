package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/google/uuid"
)

var defaultLogger *slog.Logger

// Init initializes the global logger with the specified level and format
func Init(level, format string) {
	// Parse log level
	var logLevel slog.Level
	switch strings.ToUpper(level) {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "INFO":
		logLevel = slog.LevelInfo
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	// Configure handler based on format
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: false, // Set to true in development if you need file:line info
	}

	if format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

// Get returns the default logger instance
func Get() *slog.Logger {
	if defaultLogger == nil {
		Init("INFO", "json")
	}
	return defaultLogger
}

// WithContext returns a logger with context-specific fields
func WithContext(ctx context.Context) *slog.Logger {
	logger := Get()
	
	// Add request ID if available
	if reqID := ctx.Value("request_id"); reqID != nil {
		logger = logger.With("request_id", reqID)
	}
	
	// Add user ID if available
	if userID := ctx.Value("user_id"); userID != nil {
		logger = logger.With("user_id", userID)
	}
	
	return logger
}

// WithRequestID returns a logger with a request ID attached
func WithRequestID(requestID string) *slog.Logger {
	return Get().With("request_id", requestID)
}

// WithUserID returns a logger with a user ID attached
func WithUserID(userID int64) *slog.Logger {
	return Get().With("user_id", userID)
}

// WithFields returns a logger with additional key-value pairs
func WithFields(fields ...any) *slog.Logger {
	return Get().With(fields...)
}

// NewRequestID generates a new UUID for request tracking
func NewRequestID() string {
	return uuid.New().String()
}

// Fatal logs an error message and exits the application
// This is a helper function since slog doesn't have Fatal level
func Fatal(msg string, args ...any) {
	Get().Error(msg, args...)
	os.Exit(1)
}

// FatalContext logs an error message with context and exits the application
func FatalContext(ctx context.Context, msg string, args ...any) {
	WithContext(ctx).Error(msg, args...)
	os.Exit(1)
}