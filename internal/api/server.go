package api

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"bulbul/internal/cache"
	"bulbul/internal/config"
	"bulbul/internal/database"
	"bulbul/internal/external"
	"bulbul/internal/handlers"
	"bulbul/internal/logger"
	"bulbul/internal/messaging"
	"bulbul/internal/middleware"
	"bulbul/internal/repository"
	"bulbul/internal/search"
	"bulbul/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server представляет HTTP сервер API
type Server struct {
	router       *gin.Engine
	config       *config.Config
	db           *database.DB
	es           *search.ElasticsearchClient
	nats         *messaging.NATSClient
	services     *service.Services
	repos        *repository.Repositories
	valkeyClient *cache.ValkeyClient
}

// NewServer создает новый экземпляр сервера
func NewServer(cfg *config.Config) *Server {
	// Устанавливаем режим Gin
	gin.SetMode(cfg.GinMode)

	// Подключаемся к базе данных
	slog.Info("Connecting to database", "host", cfg.Database.Host, "port", cfg.Database.Port)
	db, err := database.Connect(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err, "host", cfg.Database.Host)
	}

	// Запускаем миграции
	slog.Info("Running database migrations")
	if err := db.RunMigrations(); err != nil {
		logger.Fatal("Failed to run migrations", "error", err)
	}

	// Инициализируем Elasticsearch для событий
	slog.Info("Connecting to Elasticsearch")
	esCfg := config.LoadElasticsearchConfig()
	es, err := search.NewElasticsearchClient(esCfg)
	if err != nil {
		logger.Fatal("Failed to connect to Elasticsearch", "error", err, "url", esCfg.URL)
	}

	// Подключаемся к NATS
	slog.Info("Connecting to NATS", "url", cfg.NATS.URL)
	natsClient, err := messaging.NewNATSClient(cfg.NATS)
	if err != nil {
		logger.Fatal("Failed to connect to NATS", "error", err, "url", cfg.NATS.URL)
	}

	// Создаем клиенты внешних сервисов
	ticketingClient := external.NewTicketingClient(cfg.Ticketing)
	paymentClient := external.NewPaymentClient(cfg.Payment)

	// Создаем репозитории с Elasticsearch для событий
	repos := repository.NewRepositoriesWithElasticsearch(db, es)

	// Создаем сервисы
	services := service.NewServices(repos, natsClient, ticketingClient, paymentClient)

	// Создаем Valkey клиент (может быть nil если не удалось подключиться)
	slog.Info("Connecting to Valkey cache")
	var valkeyClient *cache.ValkeyClient
	valkeyClient, err = cache.NewValkeyClient()
	if err != nil {
		log.Fatalf("Failed to connect to Valkey: %s", err)
		valkeyClient = nil
	} else {
		slog.Info("Successfully connected to Valkey cache")
	}

	// Создаем роутер с явной настройкой middleware
	router := gin.New()

	// Применяем middleware в правильном порядке
	router.Use(middleware.Logger())   // Логирование запросов
	router.Use(middleware.Recovery()) // Кастомное восстановление после паники
	// router.Use(middleware.CORS())     // CORS headers

	// Создаем сервер
	server := &Server{
		router:       router,
		config:       cfg,
		db:           db,
		es:           es,
		nats:         natsClient,
		services:     services,
		repos:        repos,
		valkeyClient: valkeyClient,
	}

	// Настраиваем роуты
	server.setupRoutes()

	return server
}

// setupRoutes настраивает все API роуты
func (s *Server) setupRoutes() {
	// Create handlers with services and cache
	h := handlers.NewHandlers(s.services, s.valkeyClient)

	// API routes
	api := s.router.Group("/api")
	// Обязательная Basic Auth для всех API роутов
	api.Use(middleware.BasicAuth(s.repos.Users, s.valkeyClient))
	{
		// Events endpoints
		events := api.Group("/events")
		{
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

	// Health check and monitoring endpoints
	s.router.GET("/health", s.healthCheck)
	s.router.GET("/health/db", s.dbHealthCheck)
	s.router.GET("/health/elasticsearch", s.elasticsearchHealthCheck)
	s.router.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

// healthCheck обрабатывает health check запросы
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "bulbul-api",
		"version": "1.0.0",
	})
}

// dbHealthCheck обрабатывает database health check запросы
func (s *Server) dbHealthCheck(c *gin.Context) {
	healthCheck := s.db.HealthCheck(c.Request.Context())

	status := http.StatusOK
	if healthCheck.Status != "healthy" {
		status = http.StatusServiceUnavailable
	}

	// Validate connection pool and add warnings if needed
	s.db.ValidateConnectionPool()

	c.JSON(status, healthCheck)
}

// elasticsearchHealthCheck обрабатывает Elasticsearch health check запросы
func (s *Server) elasticsearchHealthCheck(c *gin.Context) {
	ctx := c.Request.Context()

	err := s.es.HealthCheck(ctx)

	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "elasticsearch",
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
	slog.Info("Cleaning up server connections")

	if s.nats != nil {
		if err := s.nats.Close(); err != nil {
			slog.Error("Error closing NATS connection", "error", err)
		} else {
			slog.Info("NATS connection closed successfully")
		}
	}

	if s.db != nil {
		if err := s.db.Close(); err != nil {
			slog.Error("Error closing database connection", "error", err)
			return err
		} else {
			slog.Info("Database connection closed successfully")
		}
	}

	return nil
}
