package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"bulbul/internal/cache"
	internalErrors "bulbul/internal/errors"
	"bulbul/internal/service"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	services     *service.Services
	valkeyClient *cache.ValkeyClient
}

func NewHandlers(services *service.Services, valkeyClient *cache.ValkeyClient) *Handlers {
	return &Handlers{
		services:     services,
		valkeyClient: valkeyClient,
	}
}

// handleServiceError checks for specific service errors and returns appropriate HTTP status codes
func (h *Handlers) handleServiceError(c *gin.Context, err error, defaultMessage string) {
	if errors.Is(err, internalErrors.ErrUnauthorized) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if errors.Is(err, internalErrors.ErrForbidden) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	if errors.Is(err, internalErrors.ErrSeatIsNotAvailable) {
		c.JSON(419, gin.H{"error": err.Error()})
		return
	}

	// Default to internal server error for other errors
	slog.Error(defaultMessage, "error", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": defaultMessage})
}
