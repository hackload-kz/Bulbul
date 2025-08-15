package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"bulbul/internal/config"
	"bulbul/internal/handlers"
	"bulbul/internal/middleware"
)

// Server представляет HTTP сервер API
type Server struct {
	router *gin.Engine
	config *config.Config
}

// NewServer создает новый экземпляр сервера
func NewServer(cfg *config.Config) *Server {
	// Устанавливаем режим Gin
	gin.SetMode(cfg.GinMode)

	// Создаем роутер
	router := gin.Default()

	// Применяем middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())

	// Создаем сервер
	server := &Server{
		router: router,
		config: cfg,
	}

	// Настраиваем роуты
	server.setupRoutes()

	return server
}

// setupRoutes настраивает все API роуты
func (s *Server) setupRoutes() {
	// API routes
	api := s.router.Group("/api")
	{
		// Events endpoints
		events := api.Group("/events")
		{
			events.POST("", handlers.CreateEvent)
			events.GET("", handlers.ListEvents)
		}

		// Bookings endpoints
		bookings := api.Group("/bookings")
		{
			bookings.POST("", handlers.CreateBooking)
			bookings.GET("", handlers.ListBookings)
			bookings.PATCH("/initiatePayment", handlers.InitiatePayment)
			bookings.PATCH("/cancel", handlers.CancelBooking)
		}

		// Seats endpoints
		seats := api.Group("/seats")
		{
			seats.GET("", handlers.ListSeats)
			seats.PATCH("/select", handlers.SelectSeat)
			seats.PATCH("/release", handlers.ReleaseSeat)
		}

		// Payments endpoints
		payments := api.Group("/payments")
		{
			payments.GET("/success", handlers.NotifyPaymentCompleted)
			payments.GET("/fail", handlers.NotifyPaymentFailed)
		}
	}

	// Health check endpoint
	s.router.GET("/health", s.healthCheck)
}

// healthCheck обрабатывает health check запросы
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"service": "bulbul-api",
		"version": "1.0.0",
	})
}

// Run запускает HTTP сервер
func (s *Server) Run() error {
	addr := fmt.Sprintf(":%s", s.config.Port)
	return s.router.Run(addr)
}

// GetRouter возвращает роутер для тестирования
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}
