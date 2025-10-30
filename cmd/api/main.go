package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/ntttrang/go-genai-slack-assistant/internal/controller"
	"github.com/ntttrang/go-genai-slack-assistant/internal/middleware"
	"github.com/ntttrang/go-genai-slack-assistant/internal/queue"
	gormmysql "github.com/ntttrang/go-genai-slack-assistant/internal/repository/gorm-mysql"
	"github.com/ntttrang/go-genai-slack-assistant/internal/service"
	slackservice "github.com/ntttrang/go-genai-slack-assistant/internal/service/slack"
	"github.com/ntttrang/go-genai-slack-assistant/pkg/ai"
	"github.com/ntttrang/go-genai-slack-assistant/pkg/cache"
	"github.com/ntttrang/go-genai-slack-assistant/pkg/config"
	"github.com/ntttrang/go-genai-slack-assistant/pkg/database"
	"github.com/ntttrang/go-genai-slack-assistant/pkg/metrics"
	"github.com/ntttrang/go-genai-slack-assistant/pkg/security"
)

func main() {
	// Initialize logger
	log, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer func() {
		_ = log.Sync()
	}()

	log.Info("Starting Slack Translation Bot...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Error("Failed to load configuration", zap.Error(err))
		os.Exit(1)
	}
	log.Info("Configuration loaded successfully",
		zap.String("environment", cfg.Application.Environment),
		zap.String("server_address", fmt.Sprintf("%s:%s", cfg.Server.Address, cfg.Server.Port)))

	// Initialize database
	dbConfig := database.DBConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Database: cfg.Database.Database,
	}

	gormDB, err := database.NewGormDB(dbConfig)
	if err != nil {
		log.Error("Failed to initialize GORM database", zap.Error(err))
		os.Exit(1)
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Error("Failed to get sql.DB from GORM", zap.Error(err))
		os.Exit(1)
	}
	defer func() {
		_ = sqlDB.Close()
	}()
	log.Info("Database connected successfully")

	// Initialize cache (which also connects to Redis)
	_, err = cache.NewRedisCache(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password)
	if err != nil {
		log.Error("Failed to initialize cache", zap.Error(err))
		os.Exit(1)
	}
	log.Info("Redis connected successfully")

	// Create redis client for health checks
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       0,
	})
	defer func() {
		_ = redisClient.Close()
	}()

	// Initialize metrics
	metricsManager := metrics.NewMetrics()

	// Initialize AI provider (Gemini)
	geminiProvider, err := ai.NewGeminiProvider(cfg.Gemini.APIKey, cfg.Gemini.Model)
	if err != nil {
		log.Error("Failed to initialize Gemini provider", zap.Error(err))
		os.Exit(1)
	}
	defer func() {
		_ = geminiProvider.Close()
	}()
	log.Info("Gemini provider initialized successfully")

	// Initialize cache instance
	cacheInstance, err := cache.NewRedisCache(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password)
	if err != nil {
		log.Error("Failed to initialize cache instance", zap.Error(err))
		os.Exit(1)
	}

	// Initialize translation repository (implements model.TranslationRepository interface)
	translationRepo := gormmysql.NewTranslationRepository(gormDB)

	// Initialize security components
	inputValidator := security.NewInputValidator(cfg.Security.MaxInputLength)
	outputValidator := security.NewOutputValidator(cfg.Security.MaxOutputLength)
	securityMiddleware := middleware.NewSecurityMiddleware(inputValidator, outputValidator, log, cfg.Security.BlockHighThreat, cfg.Security.LogSuspiciousActivity)

	// Initialize translation use case
	cacheTTL := int64(cfg.Application.CacheTTLTranslation)
	translationUseCase := service.NewTranslationUseCase(log, translationRepo, cacheInstance, geminiProvider, cacheTTL, securityMiddleware)

	// Initialize Slack client
	slackClient := slackservice.NewSlackClient(cfg.Slack.BotToken)

	// Initialize event processor (implements slack.EventProcessor interface)
	eventProc := slackservice.NewEventProcessor(translationUseCase, slackClient, log)

	// Initialize worker pool for ordered message processing
	workerPool := queue.NewWorkerPool(
		eventProc,
		cfg.Application.QueueBufferSize,
		cfg.Application.QueueIdleTimeout,
		log,
	)
	log.Info("Worker pool initialized",
		zap.Int("buffer_size", cfg.Application.QueueBufferSize),
		zap.Duration("idle_timeout", cfg.Application.QueueIdleTimeout))

	// Initialize router
	r := gin.Default()

	// Health check endpoint
	healthHandler := controller.NewHealthCheckHandler(sqlDB, redisClient, log)
	r.GET("/health", healthHandler.HandleHealthGin)

	// Metrics endpoint
	metricsHandler := controller.NewMetricsHandler(metricsManager, log)
	r.GET("/metrics", metricsHandler.HandleMetricsGin)

	// Slack webhook with signature verification
	slackGroup := r.Group("/slack")
	slackGroup.Use(middleware.VerifySlackSignatureGin(cfg.Slack.SigningSecret))
	{
		slackHandler := controller.NewSlackWebhookHandler(workerPool, log)
		slackGroup.POST("/events", slackHandler.HandleSlackEventsGin)
	}

	// Start HTTP server
	address := net.JoinHostPort(cfg.Server.Address, cfg.Server.Port)
	server := &http.Server{
		Addr:         address,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Info("Starting HTTP server", zap.String("address", address))

	// Channel to listen for server errors
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	// Channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for either server error or interrupt signal
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			log.Error("Server error", zap.Error(err))
			os.Exit(1)
		}
	case sig := <-sigChan:
		log.Info("Received shutdown signal", zap.String("signal", sig.String()))

		// Step 1: Shutdown worker pool (drain remaining messages)
		log.Info("Shutting down worker pool...")
		if err := workerPool.Shutdown(30 * time.Second); err != nil {
			log.Error("Worker pool shutdown error", zap.Error(err))
		} else {
			log.Info("Worker pool stopped successfully")
		}

		// Step 2: Shutdown HTTP server
		log.Info("Shutting down HTTP server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Error("Server shutdown error", zap.Error(err))
			os.Exit(1)
		}
	}

	log.Info("Application stopped gracefully")
}
