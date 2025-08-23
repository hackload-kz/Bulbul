package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ResetDatabase - POST /api/reset
// Сбросить базу данных в начальное состояние
func (h *Handlers) ResetDatabase(c *gin.Context) {
	err := h.services.Reset.ResetDatabase(c.Request.Context())
	if err != nil {
		slog.Error("Failed to reset database", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Database reset successfully"})
}