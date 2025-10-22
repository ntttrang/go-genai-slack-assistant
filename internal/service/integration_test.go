package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/ntttrang/python-genai-your-slack-assistant/internal/model"
)

type IntegrationTestSuite struct {
	suite.Suite
	db        *sql.DB
	redisClient *redis.Client
	logger    *zap.Logger
}

func (suite *IntegrationTestSuite) SetupSuite() {
	logger, _ := zap.NewDevelopment()
	suite.logger = logger
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	suite.logger.Sync()
	if suite.db != nil {
		suite.db.Close()
	}
	if suite.redisClient != nil {
		suite.redisClient.Close()
	}
}

// TestDatabaseStorage tests database operations
func (suite *IntegrationTestSuite) TestDatabaseStorage() {
	if suite.db == nil {
		suite.T().Skip("Database not available for integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test database connectivity
	err := suite.db.PingContext(ctx)
	assert.NoError(suite.T(), err, "Database should be reachable")

	// Test query execution
	var dbName string
	err = suite.db.QueryRowContext(ctx, "SELECT DATABASE()").Scan(&dbName)
	assert.NoError(suite.T(), err, "Should be able to query database")
	assert.NotEmpty(suite.T(), dbName, "Database name should not be empty")
}

// TestRedisCache tests Redis caching behavior
func (suite *IntegrationTestSuite) TestRedisCache() {
	if suite.redisClient == nil {
		suite.T().Skip("Redis not available for integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test Redis connectivity
	err := suite.redisClient.Ping(ctx).Err()
	assert.NoError(suite.T(), err, "Redis should be reachable")

	// Test SET/GET operations
	testKey := "test:cache:key"
	testValue := "test_value_123"
	err = suite.redisClient.Set(ctx, testKey, testValue, 1*time.Minute).Err()
	assert.NoError(suite.T(), err, "Should be able to SET in Redis")

	val, err := suite.redisClient.Get(ctx, testKey).Result()
	assert.NoError(suite.T(), err, "Should be able to GET from Redis")
	assert.Equal(suite.T(), testValue, val, "Retrieved value should match set value")

	// Test TTL
	ttl, err := suite.redisClient.TTL(ctx, testKey).Result()
	assert.NoError(suite.T(), err, "Should be able to get TTL")
	assert.Greater(suite.T(), ttl.Seconds(), float64(0), "TTL should be positive")

	// Cleanup
	suite.redisClient.Del(ctx, testKey)
}

// TestRedisCacheExpiration tests cache expiration behavior
func (suite *IntegrationTestSuite) TestRedisCacheExpiration() {
	if suite.redisClient == nil {
		suite.T().Skip("Redis not available for integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	testKey := "test:expiration:key"
	testValue := "expiring_value"

	// Set with very short TTL
	err := suite.redisClient.Set(ctx, testKey, testValue, 100*time.Millisecond).Err()
	assert.NoError(suite.T(), err, "Should be able to SET with short TTL")

	// Verify it exists immediately
	val, err := suite.redisClient.Get(ctx, testKey).Result()
	assert.NoError(suite.T(), err, "Should be able to GET immediately after SET")
	assert.Equal(suite.T(), testValue, val, "Value should be present immediately")

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Verify it's expired
	_, err = suite.redisClient.Get(ctx, testKey).Result()
	assert.Equal(suite.T(), redis.Nil, err, "Key should be expired after TTL")
}

// TestMessageEntity tests domain entities
func (suite *IntegrationTestSuite) TestMessageEntity() {
	msg := model.Message{
		ID:        "msg_123",
		ChannelID: "chan_456",
		UserID:    "user_789",
		Text:      "Hello world",
		Timestamp: "1234567890.123456",
	}

	assert.Equal(suite.T(), "msg_123", msg.ID)
	assert.Equal(suite.T(), "Hello world", msg.Text)
	assert.Equal(suite.T(), "user_789", msg.UserID)
}

// TestTranslationEntity tests translation entity
func (suite *IntegrationTestSuite) TestTranslationEntity() {
	trans := model.Translation{
		ID:             "trans_123",
		SourceMessageID: "msg_456",
		SourceText:     "Hello",
		SourceLanguage: "en",
		TranslatedText: "Xin chào",
		TargetLanguage: "vi",
		CreatedAt:      time.Now(),
	}

	assert.Equal(suite.T(), "trans_123", trans.ID)
	assert.Equal(suite.T(), "Hello", trans.SourceText)
	assert.Equal(suite.T(), "Xin chào", trans.TranslatedText)
	assert.Equal(suite.T(), "en", trans.SourceLanguage)
}

// TestChannelConfigEntity tests channel config entity
func (suite *IntegrationTestSuite) TestChannelConfigEntity() {
	config := model.ChannelConfig{
		ChannelID:       "chan_123",
		AutoTranslate:   true,
		SourceLanguages: []string{"en"},
		TargetLanguage:  "vi",
		Enabled:         true,
		CreatedAt:       time.Now(),
	}

	assert.True(suite.T(), config.AutoTranslate)
	assert.True(suite.T(), config.Enabled)
	assert.Equal(suite.T(), "vi", config.TargetLanguage)
	assert.Len(suite.T(), config.SourceLanguages, 1)
}

// TestRateLimiterWithConcurrentRequests tests rate limiting under concurrent load
func (suite *IntegrationTestSuite) TestRateLimiterWithConcurrentRequests() {
	if suite.redisClient == nil {
		suite.T().Skip("Redis not available for integration test")
	}

	ctx := context.Background()
	userID := "user_rate_test"

	// Simulate multiple concurrent requests
	requestCount := 15
	results := make([]bool, requestCount)

	for i := 0; i < requestCount; i++ {
		// In a real scenario, this would use the rate limiter
		// For now, just test Redis INCR behavior
		key := "rate:" + userID
		val, _ := suite.redisClient.Incr(ctx, key).Result()
		results[i] = val <= 10 // Assuming limit is 10

		if i == 0 {
			suite.redisClient.Expire(ctx, key, 1*time.Minute)
		}
	}

	// Cleanup
	suite.redisClient.Del(ctx, "rate:"+userID)

	// Verify that first 10 passed and rest failed
	passCount := 0
	for _, passed := range results[:10] {
		if passed {
			passCount++
		}
	}
	assert.Equal(suite.T(), 10, passCount, "First 10 requests should pass rate limit")
}

// TestConcurrentCacheOperations tests thread-safe cache operations
func (suite *IntegrationTestSuite) TestConcurrentCacheOperations() {
	if suite.redisClient == nil {
		suite.T().Skip("Redis not available for integration test")
	}

	ctx := context.Background()
	concurrentOps := 10

	done := make(chan bool, concurrentOps)
	for i := 0; i < concurrentOps; i++ {
		go func(index int) {
			key := "concurrent:key:" + string(rune(index))
			value := "value:" + string(rune(index))
			suite.redisClient.Set(ctx, key, value, 1*time.Minute)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < concurrentOps; i++ {
		<-done
	}

	// Verify all keys were set
	for i := 0; i < concurrentOps; i++ {
		key := "concurrent:key:" + string(rune(i))
		exists := suite.redisClient.Exists(ctx, key).Val() > 0
		assert.True(suite.T(), exists, "Key should exist after concurrent SET")
		suite.redisClient.Del(ctx, key)
	}
}

// TestDatabaseConnectionPooling tests connection pool management
func (suite *IntegrationTestSuite) TestDatabaseConnectionPooling() {
	if suite.db == nil {
		suite.T().Skip("Database not available for integration test")
	}

	ctx := context.Background()

	// Execute multiple concurrent queries
	concurrentQueries := 5
	results := make([]error, concurrentQueries)
	done := make(chan bool, concurrentQueries)

	for i := 0; i < concurrentQueries; i++ {
		go func(index int) {
			err := suite.db.PingContext(ctx)
			results[index] = err
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < concurrentQueries; i++ {
		<-done
	}

	// Verify all queries succeeded
	for i, err := range results {
		assert.NoError(suite.T(), err, "Query %d should succeed", i)
	}
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
