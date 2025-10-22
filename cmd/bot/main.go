package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/ntttrang/python-genai-your-slack-assistant/infrastructure/cache"
	"github.com/ntttrang/python-genai-your-slack-assistant/infrastructure/database"
	httpInterface "github.com/ntttrang/python-genai-your-slack-assistant/interface/http"
	"github.com/ntttrang/python-genai-your-slack-assistant/infrastructure/metrics"
)

func main() {
	// Initialize logger
	log, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer log.Sync()

	log.Info("Starting Slack Translation Bot...")

	// Load environment variables
	serverPort := getEnv("SERVER_PORT", "8080")
	serverAddr := getEnv("SERVER_ADDRESS", "0.0.0.0")
	mysqlHost := getEnv("MYSQL_HOST", "localhost")
	mysqlPort := getEnvInt("MYSQL_PORT", 3306)
	mysqlUser := getEnv("MYSQL_USER", "root")
	mysqlPassword := getEnv("MYSQL_PASSWORD", "")
	mysqlDatabase := getEnv("MYSQL_DATABASE", "translation_bot")
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnvInt("REDIS_PORT", 6379)
	redisPassword := getEnv("REDIS_PASSWORD", "")
	slackSigningSecret := getEnv("SLACK_SIGNING_SECRET", "")

	// Initialize database
	dbConfig := database.DBConfig{
		Host:     mysqlHost,
		Port:     mysqlPort,
		User:     mysqlUser,
		Password: mysqlPassword,
		Database: mysqlDatabase,
	}

	db, err := database.NewDB(dbConfig)
	if err != nil {
		log.Error("Failed to initialize database", zap.Error(err))
		os.Exit(1)
	}
	defer db.Close()
	log.Info("Database connected successfully")

	// Initialize cache (which also connects to Redis)
	_, err = cache.NewRedisCache(redisHost, redisPort, redisPassword)
	if err != nil {
		log.Error("Failed to initialize cache", zap.Error(err))
		os.Exit(1)
	}
	log.Info("Redis connected successfully")

	// Create redis client for health checks
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisHost, redisPort),
		Password: redisPassword,
		DB:       0,
	})
	defer redisClient.Close()

	// Initialize metrics
	metricsManager := metrics.NewMetrics()

	// Initialize router
	r := gin.Default()

	// Health check endpoint
	healthHandler := httpInterface.NewHealthCheckHandler(db, redisClient, log)
	r.GET("/health", healthHandler.HandleHealthGin)

	// Metrics endpoint
	metricsHandler := httpInterface.NewMetricsHandler(metricsManager, log)
	r.GET("/metrics", metricsHandler.HandleMetricsGin)

	// Slack webhook with signature verification
	slackGroup := r.Group("/slack")
	slackGroup.Use(httpInterface.VerifySlackSignatureGin(slackSigningSecret))
	{
		// TODO: Initialize EventProcessor and add the Slack webhook handler
		// slackHandler := httpInterface.NewSlackWebhookHandler(eventProcessor, log)
		// slackGroup.POST("/events", slackHandler.HandleSlackEventsGin)
	}

	// Start HTTP server
	address := net.JoinHostPort(serverAddr, serverPort)
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
		log.Info("Received signal", zap.String("signal", sig.String()))
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Error("Server shutdown error", zap.Error(err))
			os.Exit(1)
		}
	}

	log.Info("Server stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
