package config

import (
	"os"
	"strconv"
	"time"

	"bulbul/internal/database"
	"bulbul/internal/external"
	"bulbul/internal/messaging"
)

// Config содержит конфигурацию приложения
type Config struct {
	Port           string
	GinMode        string
	LogLevel       string
	LogFormat      string
	RequestTimeout time.Duration

	// Performance monitoring
	PprofEnabled bool
	PprofPort    string

	Database  database.Config
	NATS      messaging.Config
	Ticketing external.TicketingConfig
	Payment   external.PaymentConfig
}

// Load загружает конфигурацию из переменных окружения
func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8081"),
		GinMode:        getEnv("GIN_MODE", "debug"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		LogFormat:      getEnv("LOG_FORMAT", "json"),
		RequestTimeout: time.Duration(getEnvInt("REQUEST_TIMEOUT_SEC", 30)) * time.Second,

		// Performance monitoring
		PprofEnabled: getEnv("PPROF_ENABLED", "false") == "true",
		PprofPort:    getEnv("PPROF_PORT", "6060"),

		Database: database.Config{
			Host:               getEnv("DB_HOST", "localhost"),
			Port:               getEnvInt("DB_PORT", 5432),
			User:               getEnv("DB_USER", "bulbul"),
			Password:           getEnv("DB_PASSWORD", "bulbul123"),
			DBName:             getEnv("DB_NAME", "bulbul"),
			SSLMode:            getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:       getEnvInt("DB_MAX_OPEN_CONNS", 100),
			MaxIdleConns:       getEnvInt("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetimeMin: getEnvInt("DB_CONN_MAX_LIFETIME_MIN", 5),
			ConnMaxIdleTimeMin: getEnvInt("DB_CONN_MAX_IDLE_TIME_MIN", 1),
		},

		NATS: messaging.Config{
			URL:       getEnv("NATS_URL", "nats://localhost:4222"),
			ClusterID: getEnv("NATS_CLUSTER_ID", "bulbul"),
			ClientID:  getEnv("NATS_CLIENT_ID", "biletter-api"),
		},

		Ticketing: external.TicketingConfig{
			BaseURL: getEnv("TICKETING_SERVICE_URL", "https://hub.hackload.kz/event-provider/common"),
			Timeout: time.Duration(getEnvInt("TICKETING_TIMEOUT_SEC", 30)) * time.Second,
		},

		Payment: external.PaymentConfig{
			BaseURL:  getEnv("PAYMENT_GATEWAY_URL", "https://hub.hackload.kz/payment-provider/common"),
			TeamSlug: getEnv("PAYMENT_TEAM_SLUG", ""),
			Password: getEnv("PAYMENT_PASSWORD", ""),
			Timeout:  time.Duration(getEnvInt("PAYMENT_TIMEOUT_SEC", 30)) * time.Second,
		},
	}
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt получает целочисленное значение переменной окружения
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
