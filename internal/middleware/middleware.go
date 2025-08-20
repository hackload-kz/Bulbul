package middleware

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"bulbul/internal/cache"
	"bulbul/internal/models"
	"bulbul/internal/repository"

	"github.com/gin-gonic/gin"
)

// Ctx key and helpers for authenticated user id
// Using unexported type to avoid collisions

type ctxKey string

const userIDKey ctxKey = "user_id"

func ContextWithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	v := ctx.Value(userIDKey)
	if v == nil {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}

// CORS middleware для обработки CORS запросов
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	}
}

// Logger middleware для структурированного логирования запросов
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Записываем время начала
		start := time.Now()

		// Выполняем запрос
		c.Next()

		// Логируем результат
		latency := time.Since(start)
		userID, exists := c.Get("user_id")

		logFields := []any{
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status_code", c.Writer.Status(),
			"latency_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		}

		if exists {
			logFields = append(logFields, "user_id", userID)
		}

		if c.Writer.Status() >= 400 {
			if len(c.Errors) > 0 {
				logFields = append(logFields, "error", c.Errors.String())
			}
			slog.Error("Request completed with error", logFields...)
		}
	}
}

// Recovery middleware для восстановления после паники с детальным логированием
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// Логируем панику с максимумом информации
		slog.Error("PANIC recovered",
			"panic", recovered,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"query", c.Request.URL.RawQuery,
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
			"headers", c.Request.Header,
		)

		// Отправляем правильный HTTP ответ клиенту
		if !c.Writer.Written() {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
		}
	})
}

// BasicAuth аутентифицирует пользователя по HTTP Basic Auth, проверяя логин/пароль в кеше Valkey, затем в БД
func BasicAuth(userRepo *repository.UserRepository, valkeyClient *cache.ValkeyClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			c.Header("WWW-Authenticate", "Basic realm=\"Restricted\"")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// По умолчанию используем email как username
		ctx := c.Request.Context()

		// Вычисляем SHA-256 хеш введенного пароля для поиска в кеше
		hash := sha256.Sum256([]byte(password))
		passwordHash := fmt.Sprintf("%x", hash)

		var userID int64
		var user *models.User
		var err error

		// Сначала пытаемся найти пользователя в кеше Valkey
		if valkeyClient != nil {
			userID, err = valkeyClient.GetUserIDByAuth(ctx, username, passwordHash)
			if err == nil {
				c.Set("user_id", userID)
				c.Request = c.Request.WithContext(ContextWithUserID(c.Request.Context(), userID))
				c.Next()
				return
			}
		}

		// Fallback: поиск в базе данных
		user, err = userRepo.GetByEmail(ctx, username)
		if err != nil || user == nil || !user.IsActive {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		verified := false
		if user.PasswordHash != "" {
			if passwordHash == user.PasswordHash {
				verified = true
			}
		}

		if !verified {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		c.Set("user_id", user.UserID)
		c.Request = c.Request.WithContext(ContextWithUserID(c.Request.Context(), user.UserID))

		c.Next()
	}
}
