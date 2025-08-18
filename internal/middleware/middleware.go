package middleware

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"time"

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

// Logger middleware для логирования запросов
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// Recovery middleware для восстановления после паники
func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}

// BasicAuth аутентифицирует пользователя по HTTP Basic Auth, проверяя логин/пароль в БД
func BasicAuth(userRepo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			c.Header("WWW-Authenticate", "Basic realm=\"Restricted\"")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// По умолчанию используем email как username
		ctx := c.Request.Context()
		user, err := userRepo.GetByEmail(ctx, username)
		if err != nil || user == nil || !user.IsActive {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// Проверяем пароль: SHA-256 хеш, затем fallback на plain (если задан)
		verified := false
		if user.PasswordHash != "" {
			// Вычисляем SHA-256 хеш введенного пароля
			hash := sha256.Sum256([]byte(password))
			passwordHash := fmt.Sprintf("%x", hash)

			if passwordHash == user.PasswordHash {
				verified = true
			}
		}
		if !verified && user.PasswordPlain != nil {
			if *user.PasswordPlain == password {
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
