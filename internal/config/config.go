package config

import (
	"os"
	"strconv"
)

// Config содержит конфигурацию приложения
type Config struct {
	Port      string
	GinMode   string
	LogLevel  string
	LogFormat string
}

// Load загружает конфигурацию из переменных окружения
func Load() *Config {
	return &Config{
		Port:      getEnv("PORT", "8081"),
		GinMode:   getEnv("GIN_MODE", "debug"),
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),
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
