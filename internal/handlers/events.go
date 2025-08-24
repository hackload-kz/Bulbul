package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Events handlers

// ListEvents - GET /api/events
// Получить список событий
func (h *Handlers) ListEvents(c *gin.Context) {
	query := c.Query("query")
	date := c.Query("date")

	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	// Validate pagination parameters
	if page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page must be >= 1"})
		return
	}

	if pageSize < 1 || pageSize > 20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pageSize must be between 1 and 20"})
		return
	}

	// Check if we should use caching
	shouldCache := h.shouldCacheEventsRequest(query, date, pageSize)

	// Try to get from cache if conditions are met and cache client is available
	if shouldCache && h.valkeyClient != nil {
		// Use raw JSON to avoid unmarshaling/marshaling overhead
		// Wrap in recovery to prevent cache issues from crashing the handler
		rawJSON, err := h.valkeyClient.GetEventsListRaw(c.Request.Context(), page, pageSize)
		if err == nil {
			// Cache hit - return raw JSON data directly
			slog.Info("Cache hit for events list", "page", page, "pageSize", pageSize)
			c.Data(http.StatusOK, "application/json", rawJSON)
			return
		}
		// Cache miss or error - continue to fetch from database
		slog.Info("Cache miss for events list", "page", page, "pageSize", pageSize, "error", err)
	}

	// Fetch from database
	response, err := h.services.Events.List(c.Request.Context(), query, date, page, pageSize)
	if err != nil {
		slog.Error("Failed to list events", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list events"})
		return
	}

	// Store in cache if conditions are met and cache client is available
	if shouldCache && h.valkeyClient != nil {
		h.valkeyClient.SetEventsList(c.Request.Context(), page, pageSize, response)
	}

	c.JSON(http.StatusOK, response)
}

// shouldCacheEventsRequest determines if the request should be cached
func (h *Handlers) shouldCacheEventsRequest(query, date string, pageSize int) bool {
	// Don't cache if query or date parameters are provided
	if query != "" || date != "" {
		return false
	}

	// Only cache if pageSize is divisible by 5
	return pageSize%5 == 0
}

// GetAnalytics - GET /api/analytics
// Получить аналитику продаж для события
func (h *Handlers) GetAnalytics(c *gin.Context) {
	// Get event ID from query parameter
	eventIDStr := c.Query("id")
	if eventIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id format"})
		return
	}

	// Get analytics from service
	analytics, err := h.services.Events.GetAnalytics(c.Request.Context(), eventID)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get analytics")
		return
	}

	c.JSON(http.StatusOK, analytics)
}
