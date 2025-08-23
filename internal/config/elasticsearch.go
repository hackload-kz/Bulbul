package config

import (
	"os"
	"strconv"
	"time"
)

// ElasticsearchConfig содержит конфигурацию для подключения к Elasticsearch
type ElasticsearchConfig struct {
	URL        string
	Index      string
	Username   string
	Password   string
	MaxRetries int
	Timeout    time.Duration
}

// LoadElasticsearchConfig загружает конфигурацию Elasticsearch из переменных окружения
func LoadElasticsearchConfig() ElasticsearchConfig {
	maxRetries := 3
	if val := os.Getenv("ELASTICSEARCH_MAX_RETRIES"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			maxRetries = parsed
		}
	}

	timeout := 30 * time.Second
	if val := os.Getenv("ELASTICSEARCH_TIMEOUT"); val != "" {
		if parsed, err := time.ParseDuration(val); err == nil {
			timeout = parsed
		}
	}

	return ElasticsearchConfig{
		URL:        getEnv("ELASTICSEARCH_URL", "http://localhost:9200"),
		Index:      getEnv("ELASTICSEARCH_INDEX", "events"),
		Username:   os.Getenv("ELASTICSEARCH_USERNAME"),
		Password:   os.Getenv("ELASTICSEARCH_PASSWORD"),
		MaxRetries: maxRetries,
		Timeout:    timeout,
	}
}

