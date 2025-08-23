package search

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"bulbul/internal/config"
	"bulbul/internal/models"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// ElasticsearchClient представляет клиент для работы с Elasticsearch
type ElasticsearchClient struct {
	client *elasticsearch.Client
	config config.ElasticsearchConfig
}

// NewElasticsearchClient создает новый клиент Elasticsearch
func NewElasticsearchClient(cfg config.ElasticsearchConfig) (*ElasticsearchClient, error) {
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses:     []string{cfg.URL},
		Username:      cfg.Username,
		Password:      cfg.Password,
		RetryOnStatus: []int{502, 503, 504, 429},
		MaxRetries:    cfg.MaxRetries,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	client := &ElasticsearchClient{
		client: es,
		config: cfg,
	}

	// Check connection and create index if needed
	if err := client.ensureIndex(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure index exists: %w", err)
	}

	return client, nil
}

// ensureIndex создает индекс если он не существует
func (c *ElasticsearchClient) ensureIndex(ctx context.Context) error {
	// Check if index exists
	req := esapi.IndicesExistsRequest{
		Index: []string{c.config.Index},
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		slog.Info("Elasticsearch index already exists", "index", c.config.Index)
		return nil
	}

	// Create index with Russian analyzer
	mapping := map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
			"analysis": map[string]interface{}{
				"analyzer": map[string]interface{}{
					"russian_analyzer": map[string]interface{}{
						"type":      "custom",
						"tokenizer": "standard",
						"filter":    []string{"lowercase", "russian_stop", "russian_stemmer"},
					},
				},
				"filter": map[string]interface{}{
					"russian_stop": map[string]interface{}{
						"type":      "stop",
						"stopwords": "_russian_",
					},
					"russian_stemmer": map[string]interface{}{
						"type":     "stemmer",
						"language": "russian",
					},
				},
			},
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "long",
				},
				"title": map[string]interface{}{
					"type":     "text",
					"analyzer": "russian_analyzer",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type":         "keyword",
							"ignore_above": 256,
						},
					},
				},
				"description": map[string]interface{}{
					"type":     "text",
					"analyzer": "russian_analyzer",
				},
				"type": map[string]interface{}{
					"type": "keyword",
				},
				"datetime_start": map[string]interface{}{
					"type":   "date",
					"format": "strict_date_optional_time||epoch_millis",
				},
				"provider": map[string]interface{}{
					"type": "keyword",
				},
				"external": map[string]interface{}{
					"type": "boolean",
				},
				"total_seats": map[string]interface{}{
					"type": "integer",
				},
				"created_at": map[string]interface{}{
					"type": "date",
				},
				"updated_at": map[string]interface{}{
					"type": "date",
				},
			},
		},
	}

	mappingJSON, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("failed to marshal mapping: %w", err)
	}

	createReq := esapi.IndicesCreateRequest{
		Index: c.config.Index,
		Body:  strings.NewReader(string(mappingJSON)),
	}

	createRes, err := createReq.Do(ctx, c.client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer createRes.Body.Close()

	if createRes.IsError() {
		return fmt.Errorf("failed to create index: %s", createRes.String())
	}

	slog.Info("Created Elasticsearch index", "index", c.config.Index)
	return nil
}

// GetByID получает событие по ID
func (c *ElasticsearchClient) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	req := esapi.GetRequest{
		Index:      c.config.Index,
		DocumentID: strconv.FormatInt(id, 10),
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return nil, nil
	}

	if res.IsError() {
		return nil, fmt.Errorf("Elasticsearch error: %s", res.String())
	}

	var response struct {
		Source models.Event `json:"_source"`
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response.Source, nil
}

// Search выполняет поиск событий
func (c *ElasticsearchClient) Search(ctx context.Context, query, date string, page, pageSize int) ([]models.Event, error) {
	searchQuery := c.buildSearchQuery(query, date)
	
	// Calculate offset
	from := 0
	if page > 0 && pageSize > 0 {
		from = (page - 1) * pageSize
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	searchRequest := map[string]interface{}{
		"query": searchQuery,
		"sort": c.buildSortQuery(query),
		"from": from,
		"size": pageSize,
	}

	searchJSON, err := json.Marshal(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{c.config.Index},
		Body:  strings.NewReader(string(searchJSON)),
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search error: %s", res.String())
	}

	var response struct {
		Hits struct {
			Hits []struct {
				Source models.Event `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	events := make([]models.Event, len(response.Hits.Hits))
	for i, hit := range response.Hits.Hits {
		events[i] = hit.Source
	}

	return events, nil
}

// buildSearchQuery строит поисковый запрос
func (c *ElasticsearchClient) buildSearchQuery(query, date string) map[string]interface{} {
	mustQueries := []map[string]interface{}{}

	// Add text search query
	if query != "" {
		mustQueries = append(mustQueries, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":    query,
				"fields":   []string{"title^2", "description"},
				"analyzer": "russian_analyzer",
				"fuzziness": "AUTO",
			},
		})
	}

	// Add date filter
	if date != "" {
		mustQueries = append(mustQueries, map[string]interface{}{
			"range": map[string]interface{}{
				"datetime_start": map[string]interface{}{
					"gte": date + "T00:00:00",
					"lte": date + "T23:59:59",
				},
			},
		})
	}

	if len(mustQueries) == 0 {
		return map[string]interface{}{
			"match_all": map[string]interface{}{},
		}
	}

	return map[string]interface{}{
		"bool": map[string]interface{}{
			"must": mustQueries,
		},
	}
}

// buildSortQuery строит сортировку
func (c *ElasticsearchClient) buildSortQuery(query string) []map[string]interface{} {
	if query != "" {
		// Sort by relevance when searching
		return []map[string]interface{}{
			{"_score": map[string]interface{}{"order": "desc"}},
			{"id": map[string]interface{}{"order": "asc"}},
		}
	}

	// Sort by ID when not searching
	return []map[string]interface{}{
		{"id": map[string]interface{}{"order": "asc"}},
	}
}

// IndexEvent индексирует событие
func (c *ElasticsearchClient) IndexEvent(ctx context.Context, event *models.Event) error {
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	if event.UpdatedAt.IsZero() {
		event.UpdatedAt = time.Now()
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      c.config.Index,
		DocumentID: strconv.FormatInt(event.ID, 10),
		Body:       strings.NewReader(string(eventJSON)),
		Refresh:    "wait_for",
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return fmt.Errorf("failed to index event: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("indexing error: %s", res.String())
	}

	return nil
}

// UpdateEvent обновляет событие
func (c *ElasticsearchClient) UpdateEvent(ctx context.Context, event *models.Event) error {
	event.UpdatedAt = time.Now()
	return c.IndexEvent(ctx, event)
}

// DeleteEvent удаляет событие
func (c *ElasticsearchClient) DeleteEvent(ctx context.Context, id int64) error {
	req := esapi.DeleteRequest{
		Index:      c.config.Index,
		DocumentID: strconv.FormatInt(id, 10),
		Refresh:    "wait_for",
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("delete error: %s", res.String())
	}

	return nil
}

// Count возвращает количество документов
func (c *ElasticsearchClient) Count(ctx context.Context, query, date string) (int64, error) {
	searchQuery := c.buildSearchQuery(query, date)

	countRequest := map[string]interface{}{
		"query": searchQuery,
	}

	countJSON, err := json.Marshal(countRequest)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal count query: %w", err)
	}

	req := esapi.CountRequest{
		Index: []string{c.config.Index},
		Body:  strings.NewReader(string(countJSON)),
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return 0, fmt.Errorf("failed to execute count: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return 0, fmt.Errorf("count error: %s", res.String())
	}

	var response struct {
		Count int64 `json:"count"`
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("failed to decode count response: %w", err)
	}

	return response.Count, nil
}

// HealthCheck проверяет состояние Elasticsearch
func (c *ElasticsearchClient) HealthCheck(ctx context.Context) error {
	req := esapi.ClusterHealthRequest{
		WaitForStatus: "yellow",
		Timeout:       10 * time.Second,
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("health check error: %s", res.String())
	}

	return nil
}