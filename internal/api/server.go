package api

import (
	"fmt"
	"log"
	"net/http"

	"bulbul/internal/config"
	"bulbul/internal/database"
	"bulbul/internal/external"
	"bulbul/internal/handlers"
	"bulbul/internal/messaging"
	"bulbul/internal/middleware"
	"bulbul/internal/repository"
	"bulbul/internal/service"

	"github.com/gin-gonic/gin"
)

// Server представляет HTTP сервер API
type Server struct {
	router   *gin.Engine
	config   *config.Config
	db       *database.DB
	nats     *messaging.NATSClient
	services *service.Services
	repos    *repository.Repositories
}

// NewServer создает новый экземпляр сервера
func NewServer(cfg *config.Config) *Server {
	// Устанавливаем режим Gin
	gin.SetMode(cfg.GinMode)

	// Подключаемся к базе данных
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Запускаем миграции
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Подключаемся к NATS
	natsClient, err := messaging.NewNATSClient(cfg.NATS)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}

	// Создаем клиенты внешних сервисов
	ticketingClient := external.NewTicketingClient(cfg.Ticketing)
	paymentClient := external.NewPaymentClient(cfg.Payment)

	// Создаем репозитории
	repos := repository.NewRepositories(db)

	// Создаем сервисы
	services := service.NewServices(repos, natsClient, ticketingClient, paymentClient)

	// Создаем роутер
	router := gin.Default()

	// Применяем middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())

	// Создаем сервер
	server := &Server{
		router:   router,
		config:   cfg,
		db:       db,
		nats:     natsClient,
		services: services,
		repos:    repos,
	}

	// Настраиваем роуты
	server.setupRoutes()

	return server
}

// setupRoutes настраивает все API роуты
func (s *Server) setupRoutes() {
	// Create handlers with services
	h := handlers.NewHandlers(s.services)

	// API routes
	api := s.router.Group("/api")
	// Обязательная Basic Auth для всех API роутов
	api.Use(middleware.BasicAuth(s.repos.Users))
	{
		// Events endpoints
		events := api.Group("/events")
		{
			events.POST("", h.CreateEvent)
			events.GET("", h.ListEvents)
		}

		// Bookings endpoints
		bookings := api.Group("/bookings")
		{
			bookings.POST("", h.CreateBooking)
			bookings.GET("", h.ListBookings)
			bookings.PATCH("/initiatePayment", h.InitiatePayment)
			bookings.PATCH("/cancel", h.CancelBooking)
		}

		// Seats endpoints
		seats := api.Group("/seats")
		{
			seats.GET("", h.ListSeats)
			seats.PATCH("/select", h.SelectSeat)
			seats.PATCH("/release", h.ReleaseSeat)
		}

		// Payments endpoints
		payments := api.Group("/payments")
		{
			payments.GET("/success", h.NotifyPaymentCompleted)
			payments.GET("/fail", h.NotifyPaymentFailed)
			payments.POST("/notifications", h.OnPaymentUpdates)
		}
	}

	// Health check endpoint
	s.router.GET("/health", s.healthCheck)
}

// healthCheck обрабатывает health check запросы
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
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

// Cleanup закрывает соединения
func (s *Server) Cleanup() error {
	if s.nats != nil {
		if err := s.nats.Close(); err != nil {
			log.Printf("Error closing NATS connection: %v", err)
		}
	}

	if s.db != nil {
		if err := s.db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
			return err
		}
	}

	return nil
}
