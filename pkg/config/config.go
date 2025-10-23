package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	Slack       SlackConfig
	Gemini      GeminiConfig
	Application ApplicationConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port    string
	Address string
}

// DatabaseConfig holds MySQL database configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
}

// SlackConfig holds Slack API configuration
type SlackConfig struct {
	BotToken      string
	SigningSecret string
	WebhookPath   string
}

// GeminiConfig holds Google Gemini AI configuration
type GeminiConfig struct {
	APIKey string
	Model  string
}

// ApplicationConfig holds general application configuration
type ApplicationConfig struct {
	LogLevel                  string
	Environment               string
	CacheTTLTranslation       time.Duration
	CacheTTLChannelConfig     time.Duration
	RateLimitPerUser          int
	RateLimitPerChannel       int
	MaxMessageLength          int
}

// Load reads configuration from environment variables with default values
func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:    getEnv("SERVER_PORT", "8080"),
			Address: getEnv("SERVER_ADDRESS", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("MYSQL_HOST", "localhost"),
			Port:     getEnvInt("MYSQL_PORT", 3306),
			User:     getEnv("MYSQL_USER", "root"),
			Password: getEnv("MYSQL_PASSWORD", ""),
			Database: getEnv("MYSQL_DATABASE", "translation_bot"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
		},
		Slack: SlackConfig{
			BotToken:      getEnv("SLACK_BOT_TOKEN", ""),
			SigningSecret: getEnv("SLACK_SIGNING_SECRET", ""),
			WebhookPath:   getEnv("SLACK_WEBHOOK_PATH", "/slack/events"),
		},
		Gemini: GeminiConfig{
			APIKey: getEnv("GEMINI_API_KEY", ""),
			Model:  getEnv("GEMINI_MODEL", "gemini-1.5-flash"),
		},
		Application: ApplicationConfig{
			LogLevel:                  getEnv("LOG_LEVEL", "info"),
			Environment:               getEnv("ENVIRONMENT", "development"),
			CacheTTLTranslation:       time.Duration(getEnvInt("CACHE_TTL_TRANSLATION", 86400)) * time.Second,
			CacheTTLChannelConfig:     time.Duration(getEnvInt("CACHE_TTL_CHANNEL_CONFIG", 3600)) * time.Second,
			RateLimitPerUser:          getEnvInt("RATE_LIMIT_PER_USER", 10),
			RateLimitPerChannel:       getEnvInt("RATE_LIMIT_PER_CHANNEL", 30),
			MaxMessageLength:          getEnvInt("MAX_MESSAGE_LENGTH", 10240),
		},
	}

	// Validate required configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// Validate checks if required configuration values are set
func (c *Config) Validate() error {
	if c.Slack.SigningSecret == "" {
		return fmt.Errorf("SLACK_SIGNING_SECRET is required")
	}

	if c.Database.Host == "" {
		return fmt.Errorf("MYSQL_HOST is required")
	}

	if c.Redis.Host == "" {
		return fmt.Errorf("REDIS_HOST is required")
	}

	return nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt retrieves an integer environment variable or returns a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
